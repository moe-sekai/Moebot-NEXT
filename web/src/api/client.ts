import axios from "axios";
import type {
	CommandAliasConfig,
	CommandAliasPayload,
	CommandAliasUpdateResponse,
	CommandDefinitionsResponse,
	CommandDebugBindingPayload,
	CommandParseResponse,
	CommandStatsResponse,
	DashboardData,
	GroupRow,
	HealthResponse,
	ConfigUpdateResponse,
	MasterdataReloadResponse,
	MasterdataSummary,
	PaginatedResponse,
	PublicConfig,
	RecentCommandsResponse,
	UpdatePublicConfigPayload,
	RendererHealth,
	RendererCardThumbnailCacheStatus,
	RendererPreviewImageResult,
	RuntimeStatus,
	SearchResponse,
	SearchType,
	SekaiSystemTestPayload,
	SekaiSystemTestResponse,
	UserRow,
} from "./types";

export const api = axios.create({
	baseURL: "/api",
	timeout: 10_000,
});

export type { DashboardData } from "./types";

export async function getHealth() {
	const { data } = await api.get<HealthResponse>("/health");
	return data;
}

export async function getDashboard() {
	const { data } = await api.get<DashboardData>("/dashboard");
	return data;
}

export async function getStatus() {
	const { data } = await api.get<RuntimeStatus>("/status");
	return data;
}

export async function getMasterdataSummary() {
	const { data } = await api.get<MasterdataSummary>("/masterdata/summary");
	return data;
}

export async function getRendererHealth() {
	const { data } = await api.get<RendererHealth>("/renderer/health");
	return data;
}

export async function getRendererCardThumbnailCacheStatus(region?: string) {
	const { data } = await api.get<RendererCardThumbnailCacheStatus>(
		"/renderer/cache/card-thumbnails",
		{ params: region ? { region } : undefined },
	);
	return data;
}

export async function preloadRendererCardThumbnails(region?: string) {
	const { data } = await api.post<RendererCardThumbnailCacheStatus>(
		"/renderer/cache/card-thumbnails/preload",
		null,
		{ params: region ? { region } : undefined, timeout: 0 },
	);
	return data;
}

export async function getRecentCommands(limit = 10) {
	const { data } = await api.get<RecentCommandsResponse>("/commands/recent", {
		params: { limit },
	});
	return data;
}

export async function getCommandDefinitions() {
	const { data } = await api.get<CommandDefinitionsResponse>(
		"/commands/definitions",
	);
	return data;
}

export async function parseCommand(
	input: string,
	debugBinding?: CommandDebugBindingPayload,
) {
	const { data } = await api.get<CommandParseResponse>("/commands/parse", {
		params: { q: input, ...debugBindingParams(debugBinding) },
	});
	return data;
}

export async function renderParsedCommand(
	input: string,
	width?: number,
	height?: number,
	debugBinding?: CommandDebugBindingPayload,
): Promise<RendererPreviewImageResult> {
	const startedAt = performance.now();
	const response = await api.get<Blob>("/commands/parse/image", {
		params: {
			q: input,
			...(width ? { width } : {}),
			...(height ? { height } : {}),
			...debugBindingParams(debugBinding),
			ts: Date.now(),
		},
		responseType: "blob",
		timeout: 0,
	});
	const networkMs = Math.round(performance.now() - startedAt);
	const blob = response.data;
	return {
		blob,
		url: URL.createObjectURL(blob),
			timings: {
			fonts_ms: parseHeaderNumber(response.headers["x-render-fonts-ms"]),
			images_ms: parseHeaderNumber(response.headers["x-render-images-ms"]),
			satori_ms: parseHeaderNumber(response.headers["x-render-satori-ms"]),
			resvg_ms: parseHeaderNumber(response.headers["x-render-resvg-ms"]),
			total_ms: parseHeaderNumber(response.headers["x-render-total-ms"]),
			proxy_ms: parseHeaderNumber(response.headers["x-render-proxy-ms"]),
			network_ms: networkMs,
			size_bytes:
				parseHeaderNumber(response.headers["x-render-size-bytes"]) ?? blob.size,
			image_total: parseHeaderNumber(response.headers["x-render-image-total"]),
			image_remote: parseHeaderNumber(response.headers["x-render-image-remote"]),
			image_cache_hits: parseHeaderNumber(response.headers["x-render-image-cache-hits"]),
			image_cache_misses: parseHeaderNumber(response.headers["x-render-image-cache-misses"]),
			image_cache_errors: parseHeaderNumber(response.headers["x-render-image-cache-errors"]),
		},

	};
}

export async function getCommandAliases() {
	const { data } = await api.get<CommandAliasConfig>("/commands/aliases");
	return data;
}

export async function updateCommandAliases(payload: CommandAliasPayload) {
	const { data } = await api.put<CommandAliasUpdateResponse>(
		"/commands/aliases",
		payload,
	);
	return data;
}

export async function resetCommandAliases() {
	const { data } = await api.post<CommandAliasUpdateResponse>(
		"/commands/aliases/reset",
	);
	return data;
}

export async function exportCommandAliases() {
	const { data } = await api.get<CommandAliasPayload>(
		"/commands/aliases/export",
	);
	return data;
}

export async function importCommandAliases(payload: CommandAliasPayload) {
	const { data } = await api.post<CommandAliasUpdateResponse>(
		"/commands/aliases/import",
		payload,
	);
	return data;
}

export function downloadCommandAliases(payload: CommandAliasPayload) {
	const blob = new Blob([JSON.stringify(payload, null, 2)], {
		type: "application/json",
	});
	const url = URL.createObjectURL(blob);
	const link = document.createElement("a");
	link.href = url;
	link.download = "moebot-command-aliases.json";
	link.click();
	URL.revokeObjectURL(url);
}

export async function getPublicConfig() {
	const { data } = await api.get<PublicConfig>("/config/public");
	return data;
}

export async function updatePublicConfig(payload: UpdatePublicConfigPayload) {
	const { data } = await api.put<ConfigUpdateResponse>(
		"/config/public",
		payload,
	);
	return data;
}

export async function testSekaiSystem(payload: SekaiSystemTestPayload) {
	const { data } = await api.post<SekaiSystemTestResponse>(
		"/config/sekai/test-system",
		payload,
		{ timeout: 0 },
	);
	return data;
}

export async function reloadMasterdata(region?: string) {
	const { data } = await api.post<MasterdataReloadResponse>(
		"/masterdata/reload",
		null,
		{
			params: region ? { region } : undefined,
		},
	);
	return data;
}

export async function searchMasterdata(type: SearchType, q: string) {
	const { data } = await api.get<SearchResponse>(`/search/${type}`, {
		params: { q },
	});
	return data;
}

export async function getGroups(page = 1, limit = 20) {
	const { data } = await api.get<PaginatedResponse<GroupRow>>("/groups", {
		params: { page, limit },
	});
	return data;
}

export async function getUsers(page = 1, limit = 20) {
	const { data } = await api.get<PaginatedResponse<UserRow>>("/users", {
		params: { page, limit },
	});
	return data;
}

export async function getCommandStats(days = 7) {
	const { data } = await api.get<CommandStatsResponse>("/stats/commands", {
		params: { days },
	});
	return data;
}

function debugBindingParams(binding?: CommandDebugBindingPayload) {
	if (!binding) return {};
	return {
		...(binding.region ? { debug_region: binding.region } : {}),
		...(binding.game_id ? { debug_game_id: binding.game_id } : {}),
	};
}

function parseHeaderNumber(value: unknown) {
	const raw = Array.isArray(value) ? value[0] : value;
	if (raw === undefined || raw === null || raw === "") return null;
	const numberValue = Number(raw);
	return Number.isFinite(numberValue) ? numberValue : null;
}
