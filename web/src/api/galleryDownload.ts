// 浏览器端画廊压缩包下载工具。
//
// 思路：
//   1. 分页拉取目标画廊的全部 PID 列表（GET /pics?offset&limit）。
//   2. 限定并发地 fetch 每张原图（带 cookie 凭据）。
//   3. 把每个 Response 直接交给 client-zip 的 downloadZip()，得到流式 zip 的
//      Response 对象；浏览器边压边写入磁盘，**不会**把整包驻留在内存里。
//   4. 优先使用 File System Access API（Chrome/Edge）落盘到用户选择的位置；
//      浏览器不支持时回退到经典的 <a download> Blob URL 下载。
//
// 该模块**完全不需要后端额外接口**，沿用现有 listGalleryPics + image URL。

import { downloadZip } from "client-zip";
import { listGalleryPics, galleryPicImageUrl } from "./client";

// client-zip 的 InputWithSizeMeta（接受 Uint8Array/ArrayBuffer/Blob 等
// BufferLike），其类型未在 d.ts 中导出，这里就地复刻最小子集。
type ZipEntry = {
	name: string;
	input: Uint8Array | ArrayBuffer | Blob | string;
	lastModified?: Date;
	size?: number;
};

export interface DownloadProgress {
	/** 已抓取（成功+失败） */
	fetched: number;
	/** 失败数 */
	failed: number;
	/** 总张数 */
	total: number;
	/** 阶段：listing 列表 / fetching 拉图 / writing 写盘 / done 完成 */
	phase: "listing" | "fetching" | "writing" | "done";
}

export interface DownloadOptions {
	galleryName: string;
	/** 同时进行的图片下载数，默认 4，太大会撑爆服务端 / 浏览器 */
	concurrency?: number;
	/** 单页 PID 拉取上限（后端默认 100） */
	pageSize?: number;
	onProgress?: (p: DownloadProgress) => void;
	signal?: AbortSignal;
}

interface PicMeta {
	pid: number;
	path: string;
}

/**
 * 抓取目标画廊的全部 PID 列表。后端返回 total 字段，所以一次拉一页拿到 total
 * 后再决定是否继续翻页。
 */
async function fetchAllPics(
	name: string,
	pageSize: number,
	signal?: AbortSignal,
): Promise<PicMeta[]> {
	const all: PicMeta[] = [];
	let offset = 0;
	while (true) {
		if (signal?.aborted) throw new DOMException("aborted", "AbortError");
		const { pics, total } = await listGalleryPics(name, offset, pageSize);
		for (const p of pics) {
			all.push({ pid: p.pid, path: p.path || "" });
		}
		offset += pics.length;
		if (!pics.length || all.length >= total) break;
	}
	return all;
}

/**
 * 根据原图 path 推断扩展名；后端是按上传时的 hash + 扩展存的，这里直接用 path
 * 后缀（如 .jpg / .png / .gif），拿不到就退化成 .jpg。
 */
function pickExt(path: string): string {
	const m = /\.([a-zA-Z0-9]{1,5})(?:\?|$)/.exec(path);
	if (!m) return ".jpg";
	const ext = m[1].toLowerCase();
	if (["jpg", "jpeg", "png", "gif", "webp", "bmp"].includes(ext)) {
		return "." + ext;
	}
	return ".jpg";
}

/**
 * 主入口：下载整个画廊为 zip。
 *
 * 注意：client-zip 接受 `{ name, input }` 项，input 可以是 Response，库内部会
 * 直接消费它的 body（ReadableStream），所以**整个 zip 是流式生成的**，浏览器
 * 内存里同时只会有当前正在写入的那张图。
 */
export async function downloadGalleryAsZip(opts: DownloadOptions): Promise<void> {
	const {
		galleryName,
		concurrency = 4,
		pageSize = 200,
		onProgress,
		signal,
	} = opts;

	const emit = (p: Partial<DownloadProgress>, base: DownloadProgress) => {
		Object.assign(base, p);
		onProgress?.({ ...base });
	};

	const progress: DownloadProgress = {
		fetched: 0,
		failed: 0,
		total: 0,
		phase: "listing",
	};
	onProgress?.({ ...progress });

	// 1) 列表
	const pics = await fetchAllPics(galleryName, pageSize, signal);
	progress.total = pics.length;
	emit({ phase: "fetching" }, progress);

	if (pics.length === 0) {
		throw new Error("该画廊没有图片");
	}

	// 2) 并发抓取 -> 收集成 client-zip 的 input 列表
	//    用一个简单的"工人池"：N 个 worker 从队列里取下一个任务执行。
	const queue = [...pics];
	const entries: ZipEntry[] = new Array(pics.length);

	async function worker() {
		while (queue.length) {
			if (signal?.aborted) throw new DOMException("aborted", "AbortError");
			const next = queue.shift();
			if (!next) return;
			const idx = pics.indexOf(next);
			try {
				const resp = await fetch(galleryPicImageUrl(next.pid), {
					credentials: "include",
					signal,
				});
				if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
				const buf = await resp.arrayBuffer();
				entries[idx] = {
					name: `${next.pid}${pickExt(next.path)}`,
					input: new Uint8Array(buf),
					lastModified: new Date(),
				};
			} catch (e) {
				progress.failed++;
				console.warn(`[gallery-zip] pid=${next.pid} 下载失败:`, e);
			} finally {
				progress.fetched++;
				onProgress?.({ ...progress });
			}
		}
	}

	const workers = Array.from(
		{ length: Math.max(1, Math.min(concurrency, pics.length)) },
		() => worker(),
	);
	await Promise.all(workers);

	// 过滤掉失败的空槽
	const finalEntries = entries.filter((x): x is ZipEntry => !!x);
	if (finalEntries.length === 0) {
		throw new Error("所有图片都下载失败");
	}

	// 3) 流式打包 + 落盘
	emit({ phase: "writing" }, progress);
	const zipResp = downloadZip(finalEntries);
	const fileName = `${sanitizeFileName(galleryName)}.zip`;

	await saveStreamingResponse(zipResp, fileName, signal);

	emit({ phase: "done" }, progress);
}

function sanitizeFileName(name: string): string {
	return name.replace(/[\\/:*?"<>|]/g, "_") || "gallery";
}

/**
 * 把 client-zip 的 Response 写到磁盘：
 *   - 优先 File System Access API（真正的边压边写盘，超大画廊也不爆内存）
 *   - 否则回退 Blob + a[download]（受浏览器单 Blob 大小限制，但够用）
 */
async function saveStreamingResponse(
	resp: Response,
	suggestedName: string,
	signal?: AbortSignal,
) {
	// File System Access API
	const w = window as any;
	if (typeof w.showSaveFilePicker === "function" && resp.body) {
		try {
			const handle = await w.showSaveFilePicker({
				suggestedName,
				types: [
					{
						description: "Zip archive",
						accept: { "application/zip": [".zip"] },
					},
				],
			});
			const writable = await handle.createWritable();
			await resp.body.pipeTo(writable, { signal });
			return;
		} catch (e: any) {
			// 用户取消选择文件 -> AbortError，直接抛
			if (e?.name === "AbortError") throw e;
			console.warn(
				"[gallery-zip] showSaveFilePicker 失败，回退 Blob 下载:",
				e,
			);
		}
	}

	// Fallback: 一次性 Blob
	const blob = await resp.blob();
	const url = URL.createObjectURL(blob);
	const a = document.createElement("a");
	a.href = url;
	a.download = suggestedName;
	document.body.appendChild(a);
	a.click();
	a.remove();
	setTimeout(() => URL.revokeObjectURL(url), 1000);
}
