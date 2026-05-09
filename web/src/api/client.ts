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
	FilterAppListResponse,
	FilterAppPayload,
	FilterEvent,
	FilterGatewayPayload,
	FilterImportYAMLResponse,
	FilterTemplateListResponse,
	FilterTemplatePayload,
	FilterRegexTestPayload,
	FilterRegexTestResponse,
	FilterStatus,
	GroupRecentCommandsResponse,
	GroupRow,
	HealthResponse,
	UpdateGroupPayload,
	LogsQuery,
	LogsResponse,
	ConfigUpdateResponse,
	MasterdataReloadResponse,
	MasterdataSummary,
	PaginatedResponse,
	PublicConfig,
	RecentCommandsResponse,
	UpdatePublicConfigPayload,
	RendererFontsResponse,
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

export async function getRendererFonts() {
	const { data } = await api.get<RendererFontsResponse>("/renderer/fonts");
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

export async function getGroups(page = 1, limit = 20, statsDays = 7) {
	const { data } = await api.get<PaginatedResponse<GroupRow>>("/groups", {
		params: { page, limit, stats_days: statsDays },
	});
	return data;
}

export async function updateGroup(id: number, payload: UpdateGroupPayload) {
	const { data } = await api.put<{ data: GroupRow; message: string }>(
		`/groups/${id}`,
		payload,
	);
	return data;
}

export async function deleteGroup(id: number) {
	const { data } = await api.delete<{ message: string }>(`/groups/${id}`);
	return data;
}

export async function getGroupRecentCommands(id: number, limit = 20) {
	const { data } = await api.get<GroupRecentCommandsResponse>(
		`/groups/${id}/commands`,
		{ params: { limit } },
	);
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

export async function getLogs(query: LogsQuery = {}) {
	const params: Record<string, string | number> = {};
	if (query.levels && query.levels.length > 0) {
		params.level = query.levels.join(",");
	}
	if (query.q) params.q = query.q;
	if (query.limit) params.limit = query.limit;
	if (query.sinceSeq && query.sinceSeq > 0) params.since_seq = query.sinceSeq;
	const { data } = await api.get<LogsResponse>("/logs", { params });
	return data;
}

// --- Filter (OneBot gateway) ---

export async function getFilterStatus() {
	const { data } = await api.get<FilterStatus>("/filter/status");
	return data;
}

export async function getFilterGateway() {
	const { data } = await api.get<FilterGatewayPayload>("/filter/gateway");
	return data;
}

export async function updateFilterGateway(payload: FilterGatewayPayload) {
	const { data } = await api.put<FilterGatewayPayload>(
		"/filter/gateway",
		payload,
	);
	return data;
}

export async function listFilterApps() {
	const { data } = await api.get<FilterAppListResponse>("/filter/apps");
	return data.items;
}

export async function createFilterApp(payload: Partial<FilterAppPayload>) {
	const { data } = await api.post<FilterAppPayload>("/filter/apps", payload);
	return data;
}

export async function updateFilterApp(
	id: number,
	payload: Partial<FilterAppPayload>,
) {
	const { data } = await api.put<FilterAppPayload>(
		`/filter/apps/${id}`,
		payload,
	);
	return data;
}

export async function deleteFilterApp(id: number) {
	const { data } = await api.delete<{ ok: boolean }>(`/filter/apps/${id}`);
	return data;
}

export async function listFilterTemplates() {
	const { data } = await api.get<FilterTemplateListResponse>("/filter/templates");
	return data.items;
}

export async function getFilterTemplate(id: number) {
	const { data } = await api.get<FilterTemplatePayload>(`/filter/templates/${id}`);
	return data;
}

export async function createFilterTemplate(payload: Partial<FilterTemplatePayload>) {
	const { data } = await api.post<FilterTemplatePayload>(
		"/filter/templates",
		payload,
	);
	return data;
}

export async function updateFilterTemplate(
	id: number,
	payload: Partial<FilterTemplatePayload>,
) {
	const { data } = await api.put<FilterTemplatePayload>(
		`/filter/templates/${id}`,
		payload,
	);
	return data;
}

export async function deleteFilterTemplate(id: number) {
	const { data } = await api.delete<{ ok: boolean }>(`/filter/templates/${id}`);
	return data;
}

export async function getFilterRecentEvents(limit = 100) {
	const { data } = await api.get<{ items: FilterEvent[] }>(
		"/filter/events/recent",
		{ params: { limit } },
	);
	return data.items;
}

export async function testFilterRegex(payload: FilterRegexTestPayload) {
	const { data } = await api.post<FilterRegexTestResponse>(
		"/filter/test-regex",
		payload,
	);
	return data;
}

export async function exportFilterYAML(): Promise<Blob> {
	const response = await api.get<Blob>("/filter/export-yaml", {
		responseType: "blob",
	});
	return response.data;
}

export async function importFilterYAML(yaml: string, replaceAll: boolean) {
	const { data } = await api.post<FilterImportYAMLResponse>(
		"/filter/import-yaml",
		{ yaml, replace_all: replaceAll },
	);
	return data;
}

/** Returns an EventSource for `/api/filter/events`. Caller is responsible for closing it. */
export function openFilterEvents(replay = 50): EventSource {
	return new EventSource(`/api/filter/events?replay=${replay}`);
}

function debugBindingParams(binding?: CommandDebugBindingPayload) {
	if (!binding) return {};
	return {
		...(binding.region ? { debug_region: binding.region } : {}),
		...(binding.game_id ? { debug_game_id: binding.game_id } : {}),
	};
}

// ---------------------------------------------------------------------------
// Plugins
// ---------------------------------------------------------------------------

export interface PluginManifest {
	name: string;
	title: string;
	version: string;
	author?: string;
	category: "official" | "market" | "third";
	description?: string;
	repo?: string;
	homepage?: string;
	tags?: string[];
	settings_route?: string;
}

export interface PluginListItem extends PluginManifest {
	enabled: boolean;
	loaded: boolean;
	configurable: boolean;
}

export interface PluginSettingChoice {
	label: string;
	value: string;
}

export interface PluginSettingField {
	key: string;
	label: string;
	type: "string" | "int" | "float" | "bool" | "select" | "textarea";
	default?: unknown;
	description?: string;
	group?: string;
	options?: PluginSettingChoice[];
}

export interface PluginSettingsResponse {
	name: string;
	schema: PluginSettingField[];
	values: Record<string, unknown>;
	configurable: boolean;
}

export interface MarketPluginEntry {
	name: string;
	title?: string;
	path: string;
	html_url: string;
	import_path: string;
	priority?: "high" | "medium" | "low" | "";
	description?: string;
	commands?: string[];
	source?: "zerobot-plugin" | "zbputils" | string;
	loaded: boolean;
	enabled: boolean;
}

export interface MarketPluginListResponse {
	source: string;
	repo: string;
	branch: string;
	fetched_at: string;
	items: MarketPluginEntry[];
}

export async function listMarketPlugins(refresh = false): Promise<MarketPluginListResponse> {
	const { data } = await api.get<MarketPluginListResponse>("/plugins/market", {
		params: refresh ? { refresh: 1 } : undefined,
		timeout: 30_000,
	});
	return data;
}

export async function listPlugins(): Promise<PluginListItem[]> {
	const { data } = await api.get<{ plugins: PluginListItem[] }>("/plugins");
	return data.plugins ?? [];
}

export async function setPluginEnabled(name: string, enabled: boolean) {
	const action = enabled ? "enable" : "disable";
	const { data } = await api.post<{
		name: string;
		enabled: boolean;
		loaded: boolean;
		requires_restart: boolean;
	}>(`/plugins/${encodeURIComponent(name)}/${action}`);
	return data;
}

export async function getPluginConfig(name: string) {
	const { data } = await api.get<{ name: string; path: string; yaml: string }>(
		`/plugins/${encodeURIComponent(name)}/config`,
	);
	return data;
}

export async function getPluginSettings(name: string): Promise<PluginSettingsResponse> {
	const { data } = await api.get<PluginSettingsResponse>(
		`/plugins/${encodeURIComponent(name)}/settings`,
	);
	return data;
}

export async function updatePluginSettings(
	name: string,
	values: Record<string, unknown>,
): Promise<PluginSettingsResponse> {
	const { data } = await api.put<PluginSettingsResponse>(
		`/plugins/${encodeURIComponent(name)}/settings`,
		{ values },
	);
	return data;
}

// --- AutoChat 插件 per-group 配置 ---

export interface AutochatGroupSetting {
	group_id: number
	persona?: string
	willing_threshold?: number | null
	model?: string
	template?: string
	chat_enabled: boolean
	auto_enabled: boolean
}

export interface AutochatGroupListResponse {
	groups: AutochatGroupSetting[]
	default_threshold: number
}

export interface AutochatGroupUpsertPayload {
	persona?: string
	willing_threshold?: number | null
	clear_willing?: boolean
	model?: string
	template?: string
	chat_enabled?: boolean
	auto_enabled?: boolean
}

export async function listAutochatGroups() {
	const { data } = await api.get<AutochatGroupListResponse>(
		"/plugins/autochat/groups",
	)
	return data
}

export async function upsertAutochatGroup(
	gid: number | string,
	payload: AutochatGroupUpsertPayload,
) {
	const { data } = await api.put<AutochatGroupSetting>(
		`/plugins/autochat/groups/${gid}`,
		payload,
	)
	return data
}

export async function deleteAutochatGroup(gid: number | string) {
	const { data } = await api.delete<{ ok: boolean }>(
		`/plugins/autochat/groups/${gid}`,
	)
	return data
}

// --- AutoChat Overview ---

export interface AutochatOverview {
	primary_model: string
	models_count: number
	providers: {
		openai: boolean
		anthropic: boolean
		embedding: boolean
		rerank: boolean
		vector: boolean
		image_caption: boolean
	}
	keywords_count: number
	willing_threshold: number
	group_overrides: number
	persona_overrides: number
	token_stats_today: { prompt: number; completion: number; requests: number; total: number }
	token_stats_7days: { prompt: number; completion: number; requests: number; total: number }
}

export async function getAutochatOverview() {
	const { data } = await api.get<AutochatOverview>('/plugins/autochat/overview')
	return data
}

// --- AutoChat Providers ---

export interface AutochatProvider {
	name: string
	type: 'openai' | 'anthropic'
	base_url: string
	api_key: string
	timeout: number
	anthropic_version?: string
}

export interface AutochatProviders {
	provider_list: AutochatProvider[]
	llm: { models: string[]; multimodal_models: string[]; max_tokens: number; reasoning: boolean; timeout: number }
	embedding: {
		enabled: boolean
		provider: string  // 引用 provider_list[].name；空则用下方独立 base_url/api_key
		base_url: string
		api_key: string
		model: string
		dimensions: number
		timeout: number
	}
	rerank: {
		enabled: boolean
		provider: string
		base_url: string
		api_key: string
		model: string
		threshold: number
		timeout: number
	}
	vector: { enabled: boolean; dimensions: number; top_k: number }
	image_caption: { enabled: boolean; model: string; timeout: number; max_tokens: number; prompt: string }
	rag_summary: { enabled: boolean; model: string; timeout: number; max_tokens: number }
	format_repair: { enabled: boolean; model: string; timeout: number; max_tokens: number; prompt: string }
}

export async function getAutochatProviders() {
	const { data } = await api.get<AutochatProviders>('/plugins/autochat/providers')
	return data
}

export async function updateAutochatProviders(payload: AutochatProviders) {
	const { data } = await api.put<AutochatProviders>('/plugins/autochat/providers', payload)
	return data
}

// --- AutoChat Persona ---

export interface AutochatPersona {
	default_persona: string
	framework: string
	group_personas: Record<string, string>
	rag_summary: { enabled: boolean; model: string; timeout: number; max_tokens: number; prompt: string }
}

export async function getAutochatPersona() {
	const { data } = await api.get<AutochatPersona>('/plugins/autochat/persona')
	return data
}

export async function updateAutochatPersona(payload: AutochatPersona) {
	const { data } = await api.put<AutochatPersona>('/plugins/autochat/persona', payload)
	return data
}

// --- AutoChat Triggers ---

export interface AutochatTriggers {
	willing_threshold: number
	at_delta: number
	keyword_delta: number
	random_delta_max: number
	chat_cd_seconds: number
	tts_cd_seconds: number
	context_size: number
	buffer_limit: number
	reply_max_length: number
	keywords: string[]
	ignore_prefixes: string[]
	ignore_patterns: string[]
}

export async function getAutochatTriggers() {
	const { data } = await api.get<AutochatTriggers>('/plugins/autochat/triggers')
	return data
}

export async function updateAutochatTriggers(payload: AutochatTriggers) {
	const { data } = await api.put<AutochatTriggers>('/plugins/autochat/triggers', payload)
	return data
}

// --- AutoChat Provider Test & Models ---

export interface TestProviderResult {
	ok: boolean
	error?: string
	reachable?: boolean
	status?: number
	message?: string
}

export async function testAutochatProvider(payload: {
	type: string; base_url: string; api_key: string; timeout?: number
}) {
	const { data } = await api.post<TestProviderResult>('/plugins/autochat/test-provider', payload)
	return data
}

export interface ModelOption { id: string; name: string }

export async function listAutochatModels(payload: {
	type: string; base_url: string; api_key: string; timeout?: number; prefix?: string
}) {
	const { data } = await api.post<{ models: ModelOption[]; source: string }>(
		'/plugins/autochat/list-models', payload,
	)
	return data
}

// --- AutoChat YAML (settings 高级 tab) ---

export async function getAutochatYAML() {
	const { data } = await api.get<{ path: string; yaml: string }>('/plugins/autochat/yaml')
	return data
}

// --- AutoChat Memory ---

export interface AutochatMemoryGroup {
	group_id: number
	count: number
}

export interface AutochatMemoryItem {
	id: number
	group_id: number
	user_id: number
	user_name?: string
	type: 'user_memory' | 'summary'
	text: string
	timestamp: number
	score?: number
}

export async function listAutochatMemoryGroups() {
	const { data } = await api.get<{ groups: AutochatMemoryGroup[]; vector_enabled: boolean }>(
		'/plugins/autochat/memory/groups',
	)
	return data
}

export interface MemoryQueryParams {
	group_id?: number
	user_id?: number
	type?: 'user_memory' | 'summary' | ''
	q?: string
	limit?: number
}

export async function queryAutochatMemory(params: MemoryQueryParams) {
	const search: Record<string, string | number> = {}
	if (params.group_id) search.group_id = params.group_id
	if (params.user_id) search.user_id = params.user_id
	if (params.type) search.type = params.type
	if (params.q) search.q = params.q
	if (params.limit) search.limit = params.limit
	const { data } = await api.get<{
		items: AutochatMemoryItem[]
		total: number
		vector_enabled: boolean
		mode: 'semantic' | 'recent'
	}>('/plugins/autochat/memory', { params: search })
	return data
}

export async function deleteAutochatMemory(id: number) {
	const { data } = await api.delete<{ ok: boolean; id: number }>(
		`/plugins/autochat/memory/${id}`,
	)
	return data
}

// --- AutoChat Templates ---

export interface AutochatTemplate {
	name: string
	persona: string
	models: string[]
	multimodal?: boolean | null
	willing_threshold?: number | null
	at_delta?: number | null
	keyword_delta?: number | null
	random_delta_max?: number | null
	keywords: string[]
	used_by_groups: string[]
}

export async function listAutochatTemplates() {
	const { data } = await api.get<{ templates: AutochatTemplate[] }>(
		'/plugins/autochat/templates',
	)
	return data
}

export async function upsertAutochatTemplate(name: string, payload: AutochatTemplate) {
	const { data } = await api.put<AutochatTemplate>(
		`/plugins/autochat/templates/${encodeURIComponent(name)}`,
		payload,
	)
	return data
}

export async function deleteAutochatTemplate(name: string) {
	const { data } = await api.delete<{ ok: boolean; name: string }>(
		`/plugins/autochat/templates/${encodeURIComponent(name)}`,
	)
	return data
}

export async function updatePluginConfig(name: string, yaml: string) {
	const { data } = await api.put<{
		name: string;
		path: string;
		requires_restart: boolean;
	}>(`/plugins/${encodeURIComponent(name)}/config`, { yaml });
	return data;
}

function parseHeaderNumber(value: unknown) {
	const raw = Array.isArray(value) ? value[0] : value;
	if (raw === undefined || raw === null || raw === "") return null;
	const numberValue = Number(raw);
	return Number.isFinite(numberValue) ? numberValue : null;
}
