export interface DashboardData {
	commands_total: number;
	users_total: number;
	groups_total: number;
	uptime: string;
	version: string;
}

export interface HealthResponse {
	status: string;
	version: string;
	time: string;
	uptime: string;
}

export interface MasterdataCounts {
	cards: number;
	musics: number;
	events: number;
	gachas: number;
	virtual_lives?: number;
}

export interface MasterdataSummary {
	loaded: boolean;
	loaded_at: string | null;
	counts: MasterdataCounts;
}

export interface StatusBlock {
	status: string;
	ok: boolean;
	message: string;
	[key: string]: unknown;
}

export interface BotStatus extends StatusBlock {
	driver_type: string;
	listen: string;
	url_configured: boolean;
	command_prefix: string;
	nicknames: string[];
}

export interface WebStatus extends StatusBlock {
	host: string;
	port: number;
}

export interface RendererStatus extends StatusBlock {
	base_url: string;
	status_code: number;
	latency_ms: number;
	service_port: number;
	dashboard_port: number;
	precision: number;
}

export interface MasterdataStatus extends StatusBlock {
	loaded: boolean;
	loaded_at: string | null;
	counts: MasterdataCounts;
}

export interface DatabaseStatus extends StatusBlock {
	path: string;
}

export interface RuntimeStatus {
	version: string;
	time: string;
	uptime: string;
	bot: BotStatus;
	web: WebStatus;
	renderer: RendererStatus;
	masterdata: MasterdataStatus;
	database: DatabaseStatus;
}

export interface RendererHealth {
	ok: boolean;
	status: string;
	message: string;
	base_url: string;
	status_code: number;
	latency_ms: number;
	renderer_port: number;
	dashboard_port: number;
	precision: number;
	note: string;
}

export interface RecentCommand {
	id: number;
	command: string;
	platform: string;
	user_id: string;
	group_id: string;
	args: string;
	response_ms: number;
	created_at: string;
}

export interface RecentCommandsResponse {
	data: RecentCommand[];
	total: number;
	message: string;
}

export interface ConfigOption {
	key: string;
	label: string;
	description?: string;
	regions?: string[];
}

export interface ResolvedEndpoint {
	key: string;
	label: string;
	url: string;
}

export interface PublicMasterdataConfig {
	region: string;
	region_label: string;
	source: string;
	source_label: string;
	url: string;
	fallback_url: string;
	custom_url: string;
	custom_fallback_url: string;
	url_configured: boolean;
	fallback_url_configured: boolean;
	local_path: string;
	refresh_interval: number;
	endpoints: ResolvedEndpoint[];
	supported: boolean;
	error?: string;
	load_error?: string;
}

export interface PublicAssetsConfig {
	region: string;
	region_label: string;
	source: string;
	source_label: string;
	mirror: string;
	mirror_label: string;
	cdn_source: string;
	base_url: string;
	custom_base_url: string;
	renderer_source: string;
	music_alias_url: string;
	music_alias_configured: boolean;
	chart_source_url: string;
	sticker_path: string;
	supported: boolean;
	error?: string;
}

export interface PublicSekaiAPIConfig {
	enabled: boolean;
	base_url: string;
	base_url_configured: boolean;
	region: string;
	headers: Record<string, string>;
	headers_configured: boolean;
	timeout?: number;
	rate_limit?: number;
}

export interface PublicSuiteAPIConfig {
	enabled: boolean;
	url: string;
	url_configured: boolean;
	headers: Record<string, string>;
	headers_configured: boolean;
	timeout: number;
}

export interface PublicRankingAPIConfig {
	base_url_configured: boolean;
	region: string;
	timeout: number;
}

export interface PublicServerProfile {
	region: string;
	label: string;
	enabled: boolean;
	is_default: boolean;
	loaded: boolean;
	loaded_at: string | null;
	counts: MasterdataCounts;
	masterdata: PublicMasterdataConfig;
	assets: PublicAssetsConfig;
	sekai_api: PublicSekaiAPIConfig;
	suite_api: PublicSuiteAPIConfig;
	ranking_api: PublicRankingAPIConfig;
}

export interface PublicConfig {
	version: string;
	server: {
		region: string;
		label: string;
	};
	servers?: Record<string, PublicServerProfile>;
	presets: {
		regions: ConfigOption[];
		masterdata_sources: ConfigOption[];
		asset_sources: ConfigOption[];
		asset_mirrors: ConfigOption[];
	};
	web: {
		host: string;
		port: number;
	};
	bot: {
		nickname: string[];
		command_prefix: string;
		command_aliases?: Record<string, string[]>;
		driver_type: string;
		listen: string;
		url_configured: boolean;
		token_set?: boolean;
	};
	masterdata: PublicMasterdataConfig;
	sekai_api: PublicSekaiAPIConfig;
	suite_api?: PublicSuiteAPIConfig;
	ranking_api?: PublicRankingAPIConfig;
	renderer: {
		base_url: string;
		host: string;
		port: number;
		precision: number;
		cache: {
			enabled: boolean;
			path: string;
			max_size_mb: number;
			ttl_hours: number;
		};
	};
	assets: PublicAssetsConfig;
}

export interface UpdateMasterdataPayload {
	region: string;
	source: string;
	custom_url: string;
	custom_fallback_url: string;
	local_path: string;
	refresh_interval: number;
}

export interface UpdateAssetsPayload {
	region: string;
	source: string;
	mirror: string;
	custom_base_url: string;
	music_alias_url: string;
	chart_source_url: string;
	sticker_path: string;
}

export interface UpdateSekaiAPIPayload {
	enabled: boolean;
	base_url: string;
	region: string;
	headers?: Record<string, string>;
	timeout: number;
	rate_limit: number;
}

export interface UpdateSuiteAPIPayload {
	enabled: boolean;
	url: string;
	headers?: Record<string, string>;
	timeout: number;
}

export interface UpdateRankingAPIPayload {
	timeout: number;
}

export interface SekaiSystemTestPayload {
	base_url: string;
	region: string;
	headers?: Record<string, string>;
	timeout?: number;
}

export interface SekaiSystemTestResponse {
	ok: boolean;
	url: string;
	status_code?: number;
	duration_ms: number;
	message: string;
}

export interface UpdateServerProfilePayload {
	enabled: boolean;
	masterdata: UpdateMasterdataPayload;
	assets: UpdateAssetsPayload;
	sekai_api: UpdateSekaiAPIPayload;
	suite_api: UpdateSuiteAPIPayload;
	ranking_api: UpdateRankingAPIPayload;
}

export interface UpdateRendererPayload {
	precision: number;
}

export interface UpdatePublicConfigPayload {
	server: {
		region: string;
	};
	servers?: Record<string, UpdateServerProfilePayload>;
	masterdata?: UpdateMasterdataPayload;
	assets?: UpdateAssetsPayload;
	renderer?: UpdateRendererPayload;
	reload_masterdata?: boolean;
	sync_client_regions?: boolean;
}

export interface ConfigUpdateResponse {
	ok: boolean;
	message: string;
	config: PublicConfig;
}

export interface MasterdataReloadResponse {
	ok: boolean;
	message: string;
	duration_ms: number;
	loaded_at: string | null;
	counts: MasterdataCounts;
}

export type SearchType =
	| "cards"
	| "musics"
	| "events"
	| "gachas"
	| "virtual-lives";

export interface SearchResult {
	id: number;
	title: string;
	subtitle: string;
	type: string;
	[key: string]: unknown;
}

export interface SearchResponse {
	data: SearchResult[];
	total: number;
	query: string;
	message: string;
}

export interface RenderTiming {
	fonts_ms: number | null;
	images_ms: number | null;
	satori_ms: number | null;
	resvg_ms: number | null;
	total_ms: number | null;
	proxy_ms: number | null;
	network_ms: number | null;
	size_bytes: number | null;
	image_total: number | null;
	image_remote: number | null;
	image_cache_hits: number | null;
	image_cache_misses: number | null;
	image_cache_errors: number | null;
}

export interface RendererPreviewImageResult {
	url: string;
	blob: Blob;
	timings: RenderTiming;
}

export interface RendererCardThumbnailCacheStatus {
	ok: boolean;
	message: string;
	region: string;
	region_label: string;
	total_cards: number;
	total_urls: number;
	total_composite_images?: number;
	enabled: boolean;
	running: boolean;
	cache_dir: string;
	total: number;
	cached: number;
	missing: number;
	failed: number;
	downloaded: number;
	skipped: number;
	progress: number;
	started_at: string | null;
	completed_at: string | null;
	errors: string[];
	composite_total?: number;
	composite_cached?: number;
	composite_missing?: number;
	composite_failed?: number;
	composite_generated?: number;
	composite_progress?: number;
	composite_source_downloaded?: number;
	composite_source_failed?: number;
	composite_render_ms?: number;
	renderer_message?: string;
}

export type CommandMatchSource =
	| "primary"
	| "preset_alias"
	| "custom_alias"
	| string;
export type CommandRenderMode = "search" | "preview" | string;
export type CommandSearchType =
	| "card"
	| "music"
	| "event"
	| "gacha"
	| ""
	| string;

export interface CommandDefinition {
	id: string;
	name: string;
	description: string;
	primary_command: string;
	commands: string[];
	usage: string;
	template: string;
	preview_id: string;
	preset_aliases: string[];
	custom_aliases: string[];
	examples: string[];
	requires_argument: boolean;
	argument_hint: string;
	requires_binding: boolean;
	binding_kind?: string;
	binding_hint?: string;
	search_type: CommandSearchType;
	render_mode: CommandRenderMode;
}

export interface CommandRegionInfo {
	key: string;
	label: string;
}

export interface CommandDefinitionsResponse {
	data: CommandDefinition[];
	total: number;
	command_prefix: string;
	regions: CommandRegionInfo[];
	risk_message: string;
	restart_note: string;
}

export interface ParsedCommandResult {
	id: number;
	title: string;
	subtitle: string;
	type: string;
}

export interface ParsedCommand {
	raw_input: string;
	command_prefix: string;
	command_text: string;
	matched_command: string;
	matched_base: string;
	match_source: CommandMatchSource;
	region: string;
	region_label: string;
	argument: string;
	definition?: CommandDefinition;
	results: ParsedCommandResult[];
	selected?: ParsedCommandResult;
	can_render: boolean;
	render_mode: CommandRenderMode;
	preview_fallback_available: boolean;
	requires_binding: boolean;
	binding_kind?: string;
	debug_binding_used: boolean;
	message: string;
	warnings: string[];
	suggestions: string[];
}

export interface CommandDebugBindingPayload {
	region?: string;
	game_id?: string;
}

export interface CommandParseResponse {
	ok: boolean;
	parsed: ParsedCommand;
	message: string;
}

export interface CommandAliasConfig {
	data: CommandDefinition[];
	custom: Record<string, string[]>;
	preset: Record<string, string[]>;
	protected: string[];
	risk_message: string;
	restart_note: string;
	warnings: string[];
	command_names: string[];
}

export interface CommandAliasPayload {
	aliases: Record<string, string[]>;
}

export interface CommandAliasUpdateResponse {
	ok: boolean;
	message: string;
	aliases: Record<string, string[]>;
	config: CommandAliasConfig;
}

export interface PaginatedResponse<T> {
	data: T[];
	total: number;
	page: number;
	limit: number;
}

export interface GroupRow {
	id: number;
	platform: string;
	group_id: string;
	name: string;
	enabled: boolean;
}

export interface UserRow {
	id: number;
	platform: string;
	platform_id: string;
	game_id: string;
	nickname: string;
	server_region?: string;
}

export interface CommandStatRow {
	command: string;
	count: number;
	avg_ms: number;
}

export interface CommandStatsResponse {
	data: CommandStatRow[];
	since: string;
}
