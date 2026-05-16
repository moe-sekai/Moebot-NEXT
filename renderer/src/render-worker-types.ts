import type { DeckRecommendCalculateRequest, DeckRecommendCalculateResponse } from "./deck-recommend/types";
import type { DeckRecommendMasterSnapshot, DeckRecommendMusicMetasSnapshot, DeckRecommendSnapshotRequest, DeckRecommendSnapshotResponse } from "./deck-recommend/types";
import type { RenderPreviewResult } from "./preview";
import type { ChartSvgRenderTrace } from "./chart-svg-renderer";
import type { ChartRenderRequest, RenderRequest, RenderTemplateJobResult } from "./render-jobs";

export interface WorkerConfigState {
	deployer?: string;
	fontBody?: string;
	fontScore?: string;
	prepareConcurrency?: number;
}

export type WorkerJobPayload =
	| { type: "render"; request: RenderRequest; config?: WorkerConfigState }
	| { type: "preview"; id: string; width?: number; height?: number; precision?: number; config?: WorkerConfigState }
	| { type: "chart"; request: ChartRenderRequest; config?: WorkerConfigState }
	| { type: "deck-calculate"; request: DeckRecommendCalculateRequest; config?: WorkerConfigState }
	| { type: "snapshot"; request: DeckRecommendSnapshotRequest; config?: WorkerConfigState }
	| { type: "config"; config: WorkerConfigState };

export type WorkerJobResult =
	| RenderTemplateJobResult
	| RenderPreviewResult
	| ChartSvgRenderTrace
	| DeckRecommendCalculateResponse
	| DeckRecommendSnapshotResponse
	| { ok: true; config: WorkerConfigState };

export interface WorkerRequestMessage {
	id: number;
	job: WorkerJobPayload;
}

export interface WorkerSuccessMessage {
	id: number;
	ok: true;
	result: WorkerJobResult;
}

export interface WorkerErrorMessage {
	id: number;
	ok: false;
	error: string;
	stack?: string;
}

export type WorkerResponseMessage = WorkerSuccessMessage | WorkerErrorMessage;

export type SnapshotReplay = {
	region: string;
	master?: DeckRecommendMasterSnapshot;
	musicMetas?: DeckRecommendMusicMetasSnapshot;
};
