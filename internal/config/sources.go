package config

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

const (
	RegionCN = "cn"
	RegionJP = "jp"
	RegionTW = "tw"
	RegionKR = "kr"
	RegionEN = "en"

	// Suite 抓包数据固定使用 Haruki 公开 API；保留旧模式常量仅用于兼容历史配置/数据库。
	SuiteModeHaruki   = "haruki"
	SuiteModeLatest   = SuiteModeHaruki
	SuiteModeLocal    = SuiteModeHaruki
	SuiteModeMoeSekai = SuiteModeHaruki

	DefaultSuiteAPIURL = "https://suite-api.haruki.seiunx.com/public/{region}/suite/{uid}"

	MasterdataSourceMoeSekai = "moesekai"
	MasterdataSourceHaruki   = "haruki"
	MasterdataSource8823     = "8823"
	MasterdataSourceCustom   = "custom"

	AssetSourceMoeSekai  = "moesekai"
	AssetSourceSekaiBest = "sekai_best"
	AssetSourceCustom    = "custom"

	AssetMirrorMain           = "main"
	AssetMirrorBackup         = "backup"
	AssetMirrorOverseas       = "overseas"
	AssetMirrorOverseasBackup = "overseas_backup"
)

// PublicOption describes a selectable settings preset in the admin panel.
type PublicOption struct {
	Key         string   `json:"key"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Regions     []string `json:"regions,omitempty"`
}

// ResolvedEndpoint is a concrete URL endpoint after applying region/source presets.
type ResolvedEndpoint struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

// ResolvedMasterdata contains the effective masterdata source settings.
type ResolvedMasterdata struct {
	Region            string             `json:"region"`
	RegionLabel       string             `json:"region_label"`
	Source            string             `json:"source"`
	SourceLabel       string             `json:"source_label"`
	URL               string             `json:"url"`
	FallbackURL       string             `json:"fallback_url"`
	CustomURL         string             `json:"custom_url"`
	CustomFallbackURL string             `json:"custom_fallback_url"`
	LocalPath         string             `json:"local_path"`
	RefreshInterval   int                `json:"refresh_interval"`
	Endpoints         []ResolvedEndpoint `json:"endpoints"`
}

// ResolvedAssets contains the effective asset CDN settings.
type ResolvedAssets struct {
	Region        string `json:"region"`
	RegionLabel   string `json:"region_label"`
	Source        string `json:"source"`
	SourceLabel   string `json:"source_label"`
	Mirror        string `json:"mirror"`
	MirrorLabel   string `json:"mirror_label"`
	BaseURL       string `json:"base_url"`
	CustomBaseURL string `json:"custom_base_url"`
	RendererKey   string `json:"renderer_key"`
	CDNSource     string `json:"cdn_source"`
}

type masterdataPreset struct {
	URL         string
	FallbackURL string
}

type assetMirrorPreset struct {
	Domain string
}

var regionOptions = []PublicOption{
	{Key: RegionCN, Label: "国服", Description: "简体中文服 / CN"},
	{Key: RegionJP, Label: "日服", Description: "日本服 / JP"},
	{Key: RegionTW, Label: "台服", Description: "繁体中文服 / TW/TC"},
	{Key: RegionKR, Label: "韩服", Description: "韩国服 / KR"},
	{Key: RegionEN, Label: "国际服", Description: "Global / EN"},
}

var masterdataSourceOptions = []PublicOption{
	{Key: MasterdataSourceMoeSekai, Label: "MoeSekai", Description: "MoeSekai / Exmeaning masterdata 镜像", Regions: []string{RegionJP, RegionCN}},
	{Key: MasterdataSourceHaruki, Label: "Haruki GitHub", Description: "Team-Haruki 的全服 masterdata 仓库", Regions: []string{RegionJP, RegionCN, RegionTW, RegionKR, RegionEN}},
	{Key: MasterdataSource8823, Label: "8823 GitHub", Description: "kotori8823 的 JP/CN/TW masterdata 仓库", Regions: []string{RegionJP, RegionCN, RegionTW}},
	{Key: MasterdataSourceCustom, Label: "自定义", Description: "填写包含 cards.json 等文件的目录 URL", Regions: []string{RegionJP, RegionCN, RegionTW, RegionKR, RegionEN}},
}

var assetSourceOptions = []PublicOption{
	{Key: AssetSourceMoeSekai, Label: "MoeSekai", Description: "Exmeaning / pjsk.moe 资源镜像", Regions: []string{RegionJP, RegionCN}},
	{Key: AssetSourceSekaiBest, Label: "sekai.best", Description: "storage.sekai.best 五服资源", Regions: []string{RegionJP, RegionCN, RegionTW, RegionKR, RegionEN}},
	{Key: AssetSourceCustom, Label: "自定义", Description: "填写资源服务器 base URL", Regions: []string{RegionJP, RegionCN, RegionTW, RegionKR, RegionEN}},
}

var assetMirrorOptions = []PublicOption{
	{Key: AssetMirrorMain, Label: "主镜像", Description: "storage.exmeaning.com"},
	{Key: AssetMirrorBackup, Label: "备用镜像", Description: "storage2.exmeaning.com"},
	{Key: AssetMirrorOverseas, Label: "海外镜像", Description: "storage.pjsk.moe"},
	{Key: AssetMirrorOverseasBackup, Label: "海外备用", Description: "storage2.pjsk.moe"},
}

var masterdataPresets = map[string]map[string]masterdataPreset{
	MasterdataSourceMoeSekai: {
		RegionJP: {URL: "https://sk.exmeaning.com/master", FallbackURL: "https://sekaimaster.exmeaning.com/master"},
		RegionCN: {URL: "https://sk-cn.exmeaning.com/master", FallbackURL: "https://sekaimaster-cn.exmeaning.com/master"},
	},
	MasterdataSourceHaruki: {
		RegionJP: {URL: "https://raw.githubusercontent.com/Team-Haruki/haruki-sekai-master/main/master"},
		RegionCN: {URL: "https://raw.githubusercontent.com/Team-Haruki/haruki-sekai-sc-master/main/master"},
		RegionTW: {URL: "https://raw.githubusercontent.com/Team-Haruki/haruki-sekai-tc-master/main/master"},
		RegionKR: {URL: "https://raw.githubusercontent.com/Team-Haruki/haruki-sekai-kr-master/main/master"},
		RegionEN: {URL: "https://raw.githubusercontent.com/Team-Haruki/haruki-sekai-en-master/main/master"},
	},
	MasterdataSource8823: {
		RegionJP: {URL: "https://raw.githubusercontent.com/kotori8823/sekai-master-db/master"},
		RegionCN: {URL: "https://raw.githubusercontent.com/kotori8823/sekai-sc-master-db/master"},
		RegionTW: {URL: "https://raw.githubusercontent.com/kotori8823/sekai-tc-master-db/master"},
	},
}

var assetMirrorPresets = map[string]assetMirrorPreset{
	AssetMirrorMain:           {Domain: "https://storage.exmeaning.com"},
	AssetMirrorBackup:         {Domain: "https://storage2.exmeaning.com"},
	AssetMirrorOverseas:       {Domain: "https://storage.pjsk.moe"},
	AssetMirrorOverseasBackup: {Domain: "https://storage2.pjsk.moe"},
}

// RegionOptions returns all supported game server regions.
func RegionOptions() []PublicOption {
	return cloneOptions(regionOptions)
}

// MasterdataSourceOptions returns all supported masterdata source presets.
func MasterdataSourceOptions() []PublicOption {
	return cloneOptions(masterdataSourceOptions)
}

// AssetSourceOptions returns all supported asset source presets.
func AssetSourceOptions() []PublicOption {
	return cloneOptions(assetSourceOptions)
}

// AssetMirrorOptions returns all supported MoeSekai asset mirror presets.
func AssetMirrorOptions() []PublicOption {
	return cloneOptions(assetMirrorOptions)
}

// NormalizeRegion converts common aliases to Moebot's canonical region keys.
func NormalizeRegion(region string) string {
	switch strings.ToLower(strings.TrimSpace(region)) {
	case "cn", "sc", "c", "zh-cn", "zh_cn", "chs", "zh-hans", "zh_hans", "国服", "简中", "简体", "简体中文服":
		return RegionCN
	case "jp", "ja", "jpn", "日服", "日本", "日本服":
		return RegionJP
	case "tw", "tc", "zh-tw", "zh_tw", "zh-hant", "zh_hant", "台服", "繁中", "繁体", "繁体中文服":
		return RegionTW
	case "kr", "ko", "kor", "韩服", "韩国", "韩国服":
		return RegionKR
	case "en", "global", "intl", "ww", "gl", "international", "国际服", "全球服", "英文服":
		return RegionEN
	default:
		return strings.ToLower(strings.TrimSpace(region))
	}
}

// RegionLabel returns the display label for a canonical region key.
func RegionLabel(region string) string {
	region = NormalizeRegion(region)
	for _, option := range regionOptions {
		if option.Key == region {
			return option.Label
		}
	}
	return region
}

// IsValidRegion reports whether region is one of cn/jp/tw/kr/en.
func IsValidRegion(region string) bool {
	region = NormalizeRegion(region)
	for _, option := range regionOptions {
		if option.Key == region {
			return true
		}
	}
	return false
}

func NormalizeMasterdataSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "moesekai", "moe", "exmeaning":
		return MasterdataSourceMoeSekai
	case "haruki", "team-haruki", "team_haruki", "github_haruki", "haruki_github":
		return MasterdataSourceHaruki
	case "8823", "kotori8823", "kotori", "github_8823", "8823_github":
		return MasterdataSource8823
	case "custom", "url", "manual":
		return MasterdataSourceCustom
	default:
		return strings.ToLower(strings.TrimSpace(source))
	}
}

func NormalizeAssetSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "moesekai", "moe", "exmeaning":
		return AssetSourceMoeSekai
	case "sekai_best", "sekai-best", "sekai.best", "sekaibest":
		return AssetSourceSekaiBest
	case "custom", "url", "manual":
		return AssetSourceCustom
	default:
		return strings.ToLower(strings.TrimSpace(source))
	}
}

func NormalizeSuiteMode(mode string) string {
	// Haruki 公开 API 是唯一 Suite 数据来源；任何历史/未知模式都统一兼容到 haruki。
	return SuiteModeHaruki
}

func IsValidSuiteMode(mode string) bool {
	return NormalizeSuiteMode(mode) == SuiteModeHaruki
}

func NormalizeAssetMirror(mirror string) string {
	switch strings.ToLower(strings.TrimSpace(mirror)) {
	case "", "main", "cn", "cn_main", "cn-main", "primary":
		return AssetMirrorMain
	case "backup", "cn_backup", "cn-backup", "secondary":
		return AssetMirrorBackup
	case "overseas", "global":
		return AssetMirrorOverseas
	case "overseas_backup", "overseas-backup", "global_backup", "global-backup":
		return AssetMirrorOverseasBackup
	default:
		return strings.ToLower(strings.TrimSpace(mirror))
	}
}

// ResolveMasterdata resolves presets/custom URL settings into concrete endpoints.
func ResolveMasterdata(cfg MasterdataConfig, defaultRegion string) (ResolvedMasterdata, error) {
	region := NormalizeRegion(cfg.Region)
	if region == "" {
		region = inferMasterdataRegion(cfg)
	}
	if region == "" {
		region = NormalizeRegion(defaultRegion)
	}
	if region == "" {
		region = RegionJP
	}
	if !IsValidRegion(region) {
		return ResolvedMasterdata{}, fmt.Errorf("unsupported masterdata region %q", cfg.Region)
	}

	source := NormalizeMasterdataSource(cfg.Source)
	if source == "" {
		source = inferMasterdataSource(cfg, region)
	}

	resolved := ResolvedMasterdata{
		Region:            region,
		RegionLabel:       RegionLabel(region),
		Source:            source,
		SourceLabel:       masterdataSourceLabel(source),
		CustomURL:         normalizeBaseURL(cfg.CustomURL),
		CustomFallbackURL: normalizeBaseURL(cfg.CustomFallbackURL),
		LocalPath:         strings.TrimSpace(cfg.LocalPath),
		RefreshInterval:   cfg.RefreshInterval,
	}

	if source == MasterdataSourceCustom {
		primary := firstNonEmpty(resolved.CustomURL, normalizeBaseURL(cfg.URL))
		fallback := firstNonEmpty(resolved.CustomFallbackURL, normalizeBaseURL(cfg.FallbackURL))
		if primary == "" {
			return ResolvedMasterdata{}, fmt.Errorf("custom masterdata url is required")
		}
		if !isHTTPURL(primary) {
			return ResolvedMasterdata{}, fmt.Errorf("invalid custom masterdata url %q", primary)
		}
		resolved.URL = primary
		resolved.FallbackURL = fallback
		resolved.CustomURL = primary
		resolved.CustomFallbackURL = fallback
		resolved.Endpoints = append(resolved.Endpoints, ResolvedEndpoint{Key: "primary", Label: "主 URL", URL: primary})
		if fallback != "" {
			if !isHTTPURL(fallback) {
				return ResolvedMasterdata{}, fmt.Errorf("invalid custom masterdata fallback url %q", fallback)
			}
			resolved.Endpoints = append(resolved.Endpoints, ResolvedEndpoint{Key: "fallback", Label: "备用 URL", URL: fallback})
		}
		return resolved, nil
	}

	byRegion, ok := masterdataPresets[source]
	if !ok {
		return ResolvedMasterdata{}, fmt.Errorf("unsupported masterdata source %q", source)
	}
	preset, ok := byRegion[region]
	if !ok {
		return ResolvedMasterdata{}, fmt.Errorf("masterdata source %s does not support region %s", masterdataSourceLabel(source), RegionLabel(region))
	}
	resolved.URL = normalizeBaseURL(preset.URL)
	resolved.FallbackURL = normalizeBaseURL(preset.FallbackURL)
	resolved.Endpoints = append(resolved.Endpoints, ResolvedEndpoint{Key: "primary", Label: "主 URL", URL: resolved.URL})
	if resolved.FallbackURL != "" {
		resolved.Endpoints = append(resolved.Endpoints, ResolvedEndpoint{Key: "fallback", Label: "备用 URL", URL: resolved.FallbackURL})
	}
	return resolved, nil
}

// ResolveAssets resolves asset CDN settings into one base URL and renderer key.
func ResolveAssets(cfg AssetsConfig, defaultRegion string) (ResolvedAssets, error) {
	source := NormalizeAssetSource(cfg.Source)
	if source == "" {
		source = inferAssetSource(cfg)
	}

	region := NormalizeRegion(cfg.Region)
	if region == "" {
		region = inferAssetRegion(cfg)
	}
	if region == "" {
		region = NormalizeRegion(defaultRegion)
	}
	if region == "" {
		region = RegionJP
	}
	if !IsValidRegion(region) {
		return ResolvedAssets{}, fmt.Errorf("unsupported asset region %q", cfg.Region)
	}

	mirror := NormalizeAssetMirror(firstNonEmpty(cfg.Mirror, cfg.CDNSource))
	resolved := ResolvedAssets{
		Region:        region,
		RegionLabel:   RegionLabel(region),
		Source:        source,
		SourceLabel:   assetSourceLabel(source),
		Mirror:        mirror,
		MirrorLabel:   assetMirrorLabel(mirror),
		CustomBaseURL: normalizeBaseURL(firstNonEmpty(cfg.CustomBaseURL, cfg.BaseURL)),
	}

	switch source {
	case AssetSourceCustom:
		baseURL := firstNonEmpty(resolved.CustomBaseURL, normalizeBaseURL(cfg.CDNSource))
		if baseURL == "" {
			return ResolvedAssets{}, fmt.Errorf("custom asset base url is required")
		}
		if !isHTTPURL(baseURL) {
			return ResolvedAssets{}, fmt.Errorf("invalid custom asset base url %q", baseURL)
		}
		resolved.BaseURL = baseURL
		resolved.CustomBaseURL = baseURL
		resolved.RendererKey = baseURL
		resolved.CDNSource = baseURL
		return resolved, nil
	case AssetSourceSekaiBest:
		resolved.Mirror = ""
		resolved.MirrorLabel = ""
		resolved.BaseURL = fmt.Sprintf("https://storage.sekai.best/sekai-%s-assets", sekaiBestRegionPath(region))
		resolved.RendererKey = fmt.Sprintf("sekai-best-%s", region)
		resolved.CDNSource = resolved.RendererKey
		return resolved, nil
	case AssetSourceMoeSekai:
		if region != RegionJP && region != RegionCN {
			return ResolvedAssets{}, fmt.Errorf("asset source %s does not support region %s", assetSourceLabel(source), RegionLabel(region))
		}
		preset, ok := assetMirrorPresets[mirror]
		if !ok {
			return ResolvedAssets{}, fmt.Errorf("unsupported asset mirror %q", mirror)
		}
		resolved.BaseURL = fmt.Sprintf("%s/sekai-%s-assets", preset.Domain, region)
		resolved.RendererKey = fmt.Sprintf("%s-%s", rendererMirrorKey(mirror), region)
		resolved.CDNSource = legacyCDNSource(mirror)
		return resolved, nil
	default:
		return ResolvedAssets{}, fmt.Errorf("unsupported asset source %q", source)
	}
}

func inferMasterdataSource(cfg MasterdataConfig, region string) string {
	primary := normalizeBaseURL(firstNonEmpty(cfg.CustomURL, cfg.URL))
	if primary == "" {
		return MasterdataSourceMoeSekai
	}
	for source, byRegion := range masterdataPresets {
		if preset, ok := byRegion[region]; ok && normalizeBaseURL(preset.URL) == primary {
			return source
		}
	}
	if strings.Contains(primary, "Team-Haruki") || strings.Contains(primary, "team-haruki") || strings.Contains(primary, "haruki-sekai") {
		return MasterdataSourceHaruki
	}
	if strings.Contains(primary, "kotori8823") {
		return MasterdataSource8823
	}
	if strings.Contains(primary, "exmeaning.com/master") || strings.Contains(primary, "exmeaning.com") {
		if _, ok := masterdataPresets[MasterdataSourceMoeSekai][region]; ok {
			return MasterdataSourceMoeSekai
		}
	}
	return MasterdataSourceCustom
}

func inferMasterdataRegion(cfg MasterdataConfig) string {
	primary := normalizeBaseURL(firstNonEmpty(cfg.CustomURL, cfg.URL))
	if strings.Contains(primary, "sk-cn.exmeaning.com") || strings.Contains(primary, "sekaimaster-cn.exmeaning.com") || strings.Contains(primary, "haruki-sekai-sc-master") || strings.Contains(primary, "sekai-sc-master-db") {
		return RegionCN
	}
	if strings.Contains(primary, "haruki-sekai-tc-master") || strings.Contains(primary, "sekai-tc-master-db") {
		return RegionTW
	}
	if strings.Contains(primary, "haruki-sekai-kr-master") {
		return RegionKR
	}
	if strings.Contains(primary, "haruki-sekai-en-master") {
		return RegionEN
	}
	if primary != "" {
		return RegionJP
	}
	return ""
}

func inferAssetSource(cfg AssetsConfig) string {
	if looksLikeHTTPURL(cfg.CustomBaseURL) || looksLikeHTTPURL(cfg.BaseURL) || looksLikeHTTPURL(cfg.CDNSource) {
		return AssetSourceCustom
	}
	cdnSource := strings.ToLower(strings.TrimSpace(cfg.CDNSource))
	if strings.Contains(cdnSource, "sekai_best") || strings.Contains(cdnSource, "sekai-best") {
		return AssetSourceSekaiBest
	}
	return AssetSourceMoeSekai
}

func inferAssetRegion(cfg AssetsConfig) string {
	baseURL := strings.ToLower(firstNonEmpty(cfg.CustomBaseURL, cfg.BaseURL, cfg.CDNSource))
	switch {
	case strings.Contains(baseURL, "sekai-cn-assets"):
		return RegionCN
	case strings.Contains(baseURL, "sekai-tc-assets"), strings.Contains(baseURL, "sekai-tw-assets"):
		return RegionTW
	case strings.Contains(baseURL, "sekai-kr-assets"):
		return RegionKR
	case strings.Contains(baseURL, "sekai-en-assets"):
		return RegionEN
	case strings.Contains(baseURL, "sekai-jp-assets"):
		return RegionJP
	default:
		return ""
	}
}

func masterdataSourceLabel(source string) string {
	for _, option := range masterdataSourceOptions {
		if option.Key == source {
			return option.Label
		}
	}
	return source
}

func assetSourceLabel(source string) string {
	for _, option := range assetSourceOptions {
		if option.Key == source {
			return option.Label
		}
	}
	return source
}

func assetMirrorLabel(mirror string) string {
	for _, option := range assetMirrorOptions {
		if option.Key == mirror {
			return option.Label
		}
	}
	return mirror
}

func normalizeBaseURL(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), "/")
}

func isHTTPURL(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func looksLikeHTTPURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	return strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func sekaiBestRegionPath(region string) string {
	if NormalizeRegion(region) == RegionTW {
		return "tc"
	}
	return NormalizeRegion(region)
}

func rendererMirrorKey(mirror string) string {
	if mirror == AssetMirrorOverseasBackup {
		return "overseas-backup"
	}
	return mirror
}

func legacyCDNSource(mirror string) string {
	switch mirror {
	case AssetMirrorBackup:
		return "cn_backup"
	case AssetMirrorOverseas:
		return "overseas"
	case AssetMirrorOverseasBackup:
		return "overseas_backup"
	default:
		return "cn_main"
	}
}

func cloneOptions(options []PublicOption) []PublicOption {
	out := make([]PublicOption, len(options))
	copy(out, options)
	for i := range out {
		out[i].Regions = append([]string(nil), out[i].Regions...)
	}
	return out
}

// SupportedMasterdataRegions returns the canonical regions supported by source.
func SupportedMasterdataRegions(source string) []string {
	source = NormalizeMasterdataSource(source)
	if source == MasterdataSourceCustom {
		return allRegionKeys()
	}
	regions := make([]string, 0, len(masterdataPresets[source]))
	for region := range masterdataPresets[source] {
		regions = append(regions, region)
	}
	sort.Strings(regions)
	return regions
}

// SupportedAssetRegions returns the canonical regions supported by source.
func SupportedAssetRegions(source string) []string {
	source = NormalizeAssetSource(source)
	switch source {
	case AssetSourceMoeSekai:
		return []string{RegionCN, RegionJP}
	case AssetSourceSekaiBest, AssetSourceCustom:
		return allRegionKeys()
	default:
		return nil
	}
}

func allRegionKeys() []string {
	regions := make([]string, 0, len(regionOptions))
	for _, option := range regionOptions {
		regions = append(regions, option.Key)
	}
	return regions
}

// RegionKeys returns all canonical supported region keys in display order.
func RegionKeys() []string {
	return allRegionKeys()
}

// IsEnabled reports whether a pointer-backed enabled flag is enabled. Nil means enabled.
func IsEnabled(enabled *bool) bool {
	return enabled == nil || *enabled
}

// EnabledPtr returns a stable pointer for YAML-backed enabled settings.
func EnabledPtr(enabled bool) *bool {
	value := enabled
	return &value
}

// DefaultGameServerProfiles returns built-in per-server defaults. JP is the
// default profile; other regions are present but disabled until enabled in the
// admin panel or config file.
func DefaultGameServerProfiles() map[string]GameServerConfig {
	profiles := make(map[string]GameServerConfig, len(regionOptions))
	for _, region := range allRegionKeys() {
		enabled := region == RegionJP
		profiles[region] = defaultGameServerProfile(region, enabled)
	}
	return profiles
}

// NormalizeConfig applies compatibility defaults after loading YAML.
func NormalizeConfig(cfg *Config) {
	if cfg == nil {
		return
	}
	cfg.Server.Region = NormalizeRegion(cfg.Server.Region)
	if cfg.Server.Region == "" || !IsValidRegion(cfg.Server.Region) {
		cfg.Server.Region = RegionJP
	}
	if cfg.Renderer.Precision <= 0 {
		cfg.Renderer.Precision = DefaultRendererPrecision
	}
	if cfg.SuiteAPI.URL == "" {
		cfg.SuiteAPI.URL = DefaultSuiteAPIURL
	}
	if cfg.SuiteAPI.Timeout <= 0 {
		cfg.SuiteAPI.Timeout = 10
	}
	cfg.SuiteAPI.DefaultMode = SuiteModeHaruki

	defaults := DefaultGameServerProfiles()
	if cfg.GameServers == nil {
		cfg.GameServers = map[string]GameServerConfig{}
	}
	for rawRegion, profile := range cfg.GameServers {
		region := NormalizeRegion(rawRegion)
		if region == "" || !IsValidRegion(region) {
			delete(cfg.GameServers, rawRegion)
			continue
		}
		if rawRegion != region {
			delete(cfg.GameServers, rawRegion)
		}
		base := defaults[region]
		cfg.GameServers[region] = mergeGameServerProfile(base, profile, region)
	}

	if !hasGameServerOverrides(cfg) {
		jp := defaults[RegionJP]
		jp.Enabled = EnabledPtr(true)
		jp.Masterdata = mergeMasterdataProfile(jp.Masterdata, cfg.Masterdata, RegionJP)
		jp.Assets = mergeAssetsProfile(jp.Assets, cfg.Assets, RegionJP)
		jp.SekaiAPI = mergeSekaiAPIProfile(jp.SekaiAPI, cfg.SekaiAPI, RegionJP)
		jp.SuiteAPI = mergeSuiteAPIProfile(jp.SuiteAPI, cfg.SuiteAPI)
		jp.RankingAPI = mergeRankingAPIProfile(jp.RankingAPI, cfg.RankingAPI, RegionJP)
		cfg.GameServers[RegionJP] = jp
	}

	if cfg.Server.Region != RegionJP {
		profile := cfg.GameServers[cfg.Server.Region]
		if profile.Enabled == nil || !*profile.Enabled {
			profile.Enabled = EnabledPtr(true)
			cfg.GameServers[cfg.Server.Region] = profile
		}
	}
	if profile := cfg.GameServers[RegionJP]; profile.Enabled == nil || !*profile.Enabled {
		profile = mergeGameServerProfile(defaults[RegionJP], profile, RegionJP)
		profile.Enabled = EnabledPtr(true)
		cfg.GameServers[RegionJP] = profile
	}
	for _, region := range allRegionKeys() {
		if _, ok := cfg.GameServers[region]; !ok {
			cfg.GameServers[region] = defaults[region]
		}
	}
}

// ResolveGameServerProfile returns a normalized profile for region.
func ResolveGameServerProfile(cfg *Config, region string) GameServerConfig {
	region = NormalizeRegion(region)
	if region == "" || !IsValidRegion(region) {
		region = RegionJP
	}
	defaults := DefaultGameServerProfiles()
	profile := defaults[region]
	if cfg != nil {
		if cfg.GameServers != nil {
			if configured, ok := cfg.GameServers[region]; ok {
				profile = mergeGameServerProfile(profile, configured, region)
			}
		}
		if region == cfg.Server.Region || (region == RegionJP && cfg.Server.Region == "") {
			profile.Masterdata = mergeMasterdataProfile(profile.Masterdata, cfg.Masterdata, region)
			profile.Assets = mergeAssetsProfile(profile.Assets, cfg.Assets, region)
			profile.SekaiAPI = mergeSekaiAPIProfile(profile.SekaiAPI, cfg.SekaiAPI, region)
			profile.SuiteAPI = mergeSuiteAPIProfile(profile.SuiteAPI, cfg.SuiteAPI)
			profile.RankingAPI = mergeRankingAPIProfile(profile.RankingAPI, cfg.RankingAPI, region)
		}
	}
	return profile
}

func defaultGameServerProfile(region string, enabled bool) GameServerConfig {
	profile := GameServerConfig{
		Enabled: EnabledPtr(enabled),
		Masterdata: MasterdataConfig{
			Region:          region,
			LocalPath:       fmt.Sprintf("./data/master/%s", region),
			RefreshInterval: 3600,
		},
		SekaiAPI: SekaiAPIConfig{
			Enabled:   false,
			BaseURL:   "https://seka-api.exmeaning.com",
			Region:    region,
			Headers:   map[string]string{},
			Timeout:   10,
			RateLimit: 30,
		},
		SuiteAPI: SuiteAPIConfig{
			Enabled:     true,
			EnabledSet:  true,
			URL:         DefaultSuiteAPIURL,
			Timeout:     10,
			DefaultMode: SuiteModeHaruki,
		},
		RankingAPI: RankingAPIConfig{
			BaseURL: "https://rks.exmeaning.com",
			Region:  region,
			Timeout: 10,
		},
		Assets: AssetsConfig{
			Region:        region,
			Mirror:        AssetMirrorMain,
			MusicAliasURL: "https://moe.exmeaning.com/data/music_alias/music_aliases.json",
			StickerPath:   "./assets/stickers",
		},
	}
	if region == RegionJP || region == RegionCN {
		profile.Masterdata.Source = MasterdataSourceMoeSekai
		profile.Assets.Source = AssetSourceMoeSekai
	} else {
		profile.Masterdata.Source = MasterdataSourceHaruki
		profile.Assets.Source = AssetSourceSekaiBest
		profile.Assets.Mirror = ""
	}
	if masterResolved, err := ResolveMasterdata(profile.Masterdata, region); err == nil {
		profile.Masterdata.URL = masterResolved.URL
		profile.Masterdata.FallbackURL = masterResolved.FallbackURL
	}
	if assetResolved, err := ResolveAssets(profile.Assets, region); err == nil {
		profile.Assets.BaseURL = assetResolved.BaseURL
		profile.Assets.CDNSource = assetResolved.CDNSource
	}
	return profile
}

func hasGameServerOverrides(cfg *Config) bool {
	if cfg == nil || len(cfg.GameServers) == 0 {
		return false
	}
	for _, profile := range cfg.GameServers {
		if profile.Enabled != nil || profile.Masterdata.Source != "" || profile.Masterdata.Region != "" || profile.Masterdata.URL != "" || profile.Masterdata.LocalPath != "" || profile.Assets.Source != "" || profile.Assets.Region != "" || profile.Assets.BaseURL != "" || profile.Assets.MusicAliasURL != "" || profile.SekaiAPI.Region != "" || profile.SekaiAPI.BaseURL != "" || profile.SuiteAPI.URL != "" || profile.SuiteAPI.EnabledSet || profile.RankingAPI.Region != "" || profile.RankingAPI.BaseURL != "" {
			return true
		}
	}
	return false
}

func mergeGameServerProfile(base GameServerConfig, override GameServerConfig, region string) GameServerConfig {
	if override.Enabled != nil {
		base.Enabled = override.Enabled
	}
	base.Masterdata = mergeMasterdataProfile(base.Masterdata, override.Masterdata, region)
	base.Assets = mergeAssetsProfile(base.Assets, override.Assets, region)
	base.SekaiAPI = mergeSekaiAPIProfile(base.SekaiAPI, override.SekaiAPI, region)
	base.SuiteAPI = mergeSuiteAPIProfile(base.SuiteAPI, override.SuiteAPI)
	base.RankingAPI = mergeRankingAPIProfile(base.RankingAPI, override.RankingAPI, region)
	return base
}

func mergeMasterdataProfile(base MasterdataConfig, override MasterdataConfig, region string) MasterdataConfig {
	if override.Region != "" {
		base.Region = NormalizeRegion(override.Region)
	}
	if base.Region == "" {
		base.Region = region
	}
	if override.Source != "" {
		base.Source = NormalizeMasterdataSource(override.Source)
	}
	if override.URL != "" {
		base.URL = override.URL
	}
	if override.FallbackURL != "" {
		base.FallbackURL = override.FallbackURL
	}
	if override.CustomURL != "" {
		base.CustomURL = override.CustomURL
	}
	if override.CustomFallbackURL != "" {
		base.CustomFallbackURL = override.CustomFallbackURL
	}
	if override.LocalPath != "" {
		base.LocalPath = override.LocalPath
	}
	if override.RefreshInterval != 0 {
		base.RefreshInterval = override.RefreshInterval
	}
	return base
}

func mergeAssetsProfile(base AssetsConfig, override AssetsConfig, region string) AssetsConfig {
	if override.Region != "" {
		base.Region = NormalizeRegion(override.Region)
	}
	if base.Region == "" {
		base.Region = region
	}
	if override.Source != "" {
		base.Source = NormalizeAssetSource(override.Source)
	}
	if override.Mirror != "" {
		base.Mirror = NormalizeAssetMirror(override.Mirror)
	}
	if override.CDNSource != "" {
		base.CDNSource = override.CDNSource
	}
	if override.BaseURL != "" {
		base.BaseURL = override.BaseURL
	}
	if override.CustomBaseURL != "" {
		base.CustomBaseURL = override.CustomBaseURL
	}
	if override.MusicAliasURL != "" {
		base.MusicAliasURL = override.MusicAliasURL
	}
	if override.StickerPath != "" {
		base.StickerPath = override.StickerPath
	}
	return base
}

func mergeSekaiAPIProfile(base SekaiAPIConfig, override SekaiAPIConfig, region string) SekaiAPIConfig {
	if override.Enabled {
		base.Enabled = true
	}
	if override.BaseURL != "" {
		base.BaseURL = override.BaseURL
	}
	if override.Region != "" {
		base.Region = NormalizeRegion(override.Region)
	}
	if base.Region == "" {
		base.Region = region
	}
	if override.Headers != nil {
		base.Headers = override.Headers
	}
	if override.Timeout != 0 {
		base.Timeout = override.Timeout
	}
	if override.RateLimit != 0 {
		base.RateLimit = override.RateLimit
	}
	return base
}

func mergeSuiteAPIProfile(base SuiteAPIConfig, override SuiteAPIConfig) SuiteAPIConfig {
	if override.EnabledSet {
		base.Enabled = override.Enabled
		base.EnabledSet = true
	}
	if override.URL != "" {
		base.URL = strings.TrimSpace(override.URL)
	}
	if override.Token != "" {
		base.Token = override.Token
	}
	if override.Timeout != 0 {
		base.Timeout = override.Timeout
	}
	if override.DefaultMode != "" {
		base.DefaultMode = NormalizeSuiteMode(override.DefaultMode)
	}
	if base.URL == "" {
		base.URL = DefaultSuiteAPIURL
	}
	if base.Timeout <= 0 {
		base.Timeout = 10
	}
	base.DefaultMode = NormalizeSuiteMode(base.DefaultMode)
	if base.DefaultMode == "" || !IsValidSuiteMode(base.DefaultMode) {
		base.DefaultMode = SuiteModeHaruki
	}
	return base
}

func mergeRankingAPIProfile(base RankingAPIConfig, override RankingAPIConfig, region string) RankingAPIConfig {
	if override.BaseURL != "" {
		base.BaseURL = override.BaseURL
	}
	if override.Region != "" {
		base.Region = NormalizeRegion(override.Region)
	}
	if base.Region == "" {
		base.Region = region
	}
	if override.Timeout != 0 {
		base.Timeout = override.Timeout
	}
	return base
}
