package commandparser

import (
	"fmt"
	"sort"
	"strings"

	"moebot-next/internal/config"
)

const (
	MatchPrimary     = "primary"
	MatchPresetAlias = "preset_alias"
	MatchCustomAlias = "custom_alias"

	RenderModeSearch  = "search"
	RenderModePreview = "preview"
	RenderModeAction  = "action"
)

// SearchType identifies a masterdata-backed parser target.
type SearchType string

const (
	SearchTypeCard        SearchType = "card"
	SearchTypeMusic       SearchType = "music"
	SearchTypeEvent       SearchType = "event"
	SearchTypeGacha       SearchType = "gacha"
	SearchTypeVirtualLive SearchType = "virtual_live"
	SearchTypeNone        SearchType = ""
)

// Definition describes a command surface shared by bot registration and WebUI parsing.
type Definition struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	PrimaryCommand   string     `json:"primary_command"`
	Commands         []string   `json:"commands"`
	Usage            string     `json:"usage"`
	Template         string     `json:"template"`
	PreviewID        string     `json:"preview_id"`
	PresetAliases    []string   `json:"preset_aliases"`
	CustomAliases    []string   `json:"custom_aliases"`
	Examples         []string   `json:"examples"`
	RequiresArgument bool       `json:"requires_argument"`
	ArgumentHint     string     `json:"argument_hint"`
	RequiresBinding  bool       `json:"requires_binding"`
	BindingKind      string     `json:"binding_kind,omitempty"`
	BindingHint      string     `json:"binding_hint,omitempty"`
	SearchType       SearchType `json:"search_type"`
	RenderMode       string     `json:"render_mode"`
}

// BotCommand binds an actual ZeroBot trigger to a canonical command definition.
type BotCommand struct {
	Name          string
	Base          string
	Primary       string
	Region        string
	MatchSource   string
	DefinitionID  string
	RequiresArg   bool
	ArgumentHint  string
	CanonicalName string
}

// CommandAliasEntry is used by the alias editor API.
type CommandAliasEntry struct {
	Command string   `json:"command"`
	Aliases []string `json:"aliases"`
}

// AliasConfigResponse summarizes effective alias settings for WebUI.
type AliasConfigResponse struct {
	Data         []Definition        `json:"data"`
	Custom       map[string][]string `json:"custom"`
	Preset       map[string][]string `json:"preset"`
	Protected    []string            `json:"protected"`
	RiskMessage  string              `json:"risk_message"`
	RestartNote  string              `json:"restart_note"`
	Warnings     []string            `json:"warnings"`
	CommandNames []string            `json:"command_names"`
}

// AliasUpdateRequest is accepted by update/import endpoints.
type AliasUpdateRequest struct {
	Aliases map[string][]string `json:"aliases"`
}

// AliasUpdateResponse is returned after alias config mutations.
type AliasUpdateResponse struct {
	OK      bool                `json:"ok"`
	Message string              `json:"message"`
	Aliases map[string][]string `json:"aliases"`
	Config  AliasConfigResponse `json:"config"`
}

const RiskMessage = "自定义关键词会影响聊天端指令触发。请避免使用过短、常见聊天词或与现有指令冲突的词；保存后 WebUI 解析立即生效，聊天端通常需要重启后生效。"
const RestartNote = "如果机器人已经启动，新增/删除的自定义聊天指令别名需要重启 Bot 后生效；WebUI 指令解析会立即使用最新配置。"

var baseDefinitions = []Definition{
	{
		ID:               "card-detail",
		Name:             "查卡/卡牌列表",
		Description:      "搜索卡牌信息；纯数字 ID 显示详情，其它角色别名、卡名和筛选条件显示列表。",
		PrimaryCommand:   "查卡",
		Commands:         []string{"查卡"},
		Usage:            "/查卡 [ID/筛选条件]",
		Template:         "card_detail",
		PreviewID:        "card-detail",
		PresetAliases:    []string{"card", "cardinfo"},
		Examples:         []string{"/查卡 1204", "/查卡 mnr 4 蓝 限定", "/查卡 纯mmj", "/查卡 event123"},
		RequiresArgument: true,
		ArgumentHint:     "请输入卡牌 ID、角色别名、卡名或筛选条件",
		SearchType:       SearchTypeCard,
		RenderMode:       RenderModeSearch,
	},
	{
		ID:               "music-detail",
		Name:             "查曲详情图",
		Description:      "搜索曲目信息，支持曲目 ID、标题、读音、作者和别名关键词。",
		PrimaryCommand:   "查曲",
		Commands:         []string{"查曲", "查歌"},
		Usage:            "/查曲 [ID/关键词]",
		Template:         "music_detail",
		PreviewID:        "music-detail",
		PresetAliases:    []string{"music", "musicinfo", "song", "songinfo"},
		Examples:         []string{"/查曲 Tell Your World", "/查歌 千本樱", "/song 1", "/songinfo 谷歌"},
		RequiresArgument: true,
		ArgumentHint:     "请输入曲目 ID、曲名、读音、作者或关键词",
		SearchType:       SearchTypeMusic,
		RenderMode:       RenderModeSearch,
	},
	{
		ID:               "chart-detail",
		Name:             "查谱详情图",
		Description:      "查询曲目谱面等级、Notes 数与谱面预览；关键词后可追加 ex/ma/mas/apd/ap/hd/nm/ez 或原称指定难度。",
		PrimaryCommand:   "查谱",
		Commands:         []string{"查谱"},
		Usage:            "/查谱 [ID/关键词]",
		Template:         "chart_detail",
		PreviewID:        "chart-detail",
		PresetAliases:    []string{"chart", "chartinfo", "谱面", "谱面预览"},
		Examples:         []string{"/查谱 Tell Your World", "/chart 1", "/谱面 739", "/谱面预览 739 ex", "/谱面预览 千本樱 mas"},
		RequiresArgument: true,
		ArgumentHint:     "请输入曲目 ID、曲名或谱面关键词",
		SearchType:       SearchTypeMusic,
		RenderMode:       RenderModeSearch,
	},
	{
		ID:               "event-info",
		Name:             "活动信息图",
		Description:      "搜索活动信息，支持活动 ID、相对索引、活动名、当前、类型和团组关键词。",
		PrimaryCommand:   "查活动",
		Commands:         []string{"查活动"},
		Usage:            "/查活动 [ID/关键词]",
		Template:         "event_info",
		PreviewID:        "event-info",
		PresetAliases:    []string{"event", "eventinfo"},
		Examples:         []string{"/查活动", "/查活动 +1", "/event 周年"},
		RequiresArgument: false,
		ArgumentHint:     "可选：活动 ID、当前、相对索引、活动名或关键词",
		SearchType:       SearchTypeEvent,
		RenderMode:       RenderModeSearch,
	},
	{
		ID:               "gacha-info",
		Name:             "卡池信息图",
		Description:      "搜索卡池/扭蛋信息，支持 ID、负数索引、当前、年份、类型和 pickup 相关关键词。",
		PrimaryCommand:   "查卡池",
		Commands:         []string{"查卡池", "查扭蛋"},
		Usage:            "/查卡池 [ID/关键词]",
		Template:         "gacha_info",
		PreviewID:        "gacha-info",
		PresetAliases:    []string{"gacha", "gachainfo"},
		Examples:         []string{"/查卡池 700", "/查扭蛋 当前", "/gacha fes"},
		RequiresArgument: false,
		ArgumentHint:     "可选：卡池 ID、当前、年份、类型或关键词",
		SearchType:       SearchTypeGacha,
		RenderMode:       RenderModeSearch,
	},
	{
		ID:               "virtual-live-list",
		Name:             "虚拟 Live 列表",
		Description:      "查询近期演唱会/虚拟 Live，支持 ID、名称、当前、未来和年份过滤。",
		PrimaryCommand:   "查演唱会",
		Commands:         []string{"查演唱会", "演唱会", "虚拟live", "查虚拟live"},
		Usage:            "/查演唱会 [ID/关键词]",
		Template:         "virtual_live_list",
		PreviewID:        "virtual-live-list",
		PresetAliases:    []string{"vlive", "live"},
		Examples:         []string{"/查演唱会", "/虚拟live 当前", "/vlive 1"},
		RequiresArgument: false,
		ArgumentHint:     "可选：虚拟 Live ID、名称、当前、未来或年份",
		SearchType:       SearchTypeVirtualLive,
		RenderMode:       RenderModeSearch,
	},
	{
		ID:             "ranking-list",
		Name:           "实时排行图",
		Description:    "查询活动整体榜线，支持五服和 WL 前缀。WebUI 解析页使用静态预览兜底，聊天端走公开实时榜线接口。",
		PrimaryCommand: "榜线",
		Commands:       []string{"榜线", "排行", "sk线", "skl"},
		Usage:          "/榜线 或 /sk线",
		Template:       "ranking_list",
		PreviewID:      "ranking-list",
		PresetAliases:  []string{"rank", "ranking", "skline"},
		Examples:       []string{"/榜线", "/sk线", "/cnskl", "/wlsk线"},
		ArgumentHint:   "可选：WL 查询时可输入角色 ID/章节序号",
		SearchType:     SearchTypeNone,
		RenderMode:     RenderModePreview,
	},
	{
		ID:             "ranking-target",
		Name:           "sk 指定榜线",
		Description:    "按排名、范围或绑定 UID 查询实时榜线，支持 1k/1w 简写与 WL 前缀。",
		PrimaryCommand: "sk",
		Commands:       []string{"sk"},
		Usage:          "/sk [排名/范围/UID]",
		Template:       "ranking_list",
		PreviewID:      "ranking-list",
		PresetAliases:  []string{},
		Examples:       []string{"/sk 100", "/sk 1k", "/sk 1-10", "/cnwlsk 1 100"},
		ArgumentHint:   "可选：排名、范围或 UID；留空时使用绑定账号",
		SearchType:     SearchTypeNone,
		RenderMode:     RenderModePreview,
	},
	{
		ID:             "churn-ranking",
		Name:           "查房",
		Description:    "查询周回、时速和最近分数变化，支持 cf 别名和 WL 前缀。",
		PrimaryCommand: "查房",
		Commands:       []string{"查房", "cf"},
		Usage:          "/查房 [排名/范围/UID]",
		Template:       "churn_ranking_list",
		PreviewID:      "ranking-list",
		PresetAliases:  []string{},
		Examples:       []string{"/cf", "/查房 100", "/cncf 1k", "/wlcf 1"},
		ArgumentHint:   "可选：排名、范围或 UID；留空时使用绑定账号/默认档位",
		SearchType:     SearchTypeNone,
		RenderMode:     RenderModePreview,
	},
	{
		ID:             "water-table",
		Name:           "查水表",
		Description:    "查询单个玩家小时周回与停车区间，支持 csb 别名和 WL 前缀。",
		PrimaryCommand: "查水表",
		Commands:       []string{"查水表", "停车时间"},
		Usage:          "/查水表 [排名/UID]",
		Template:       "water_table",
		PreviewID:      "ranking-list",
		PresetAliases:  []string{"csb"},
		Examples:       []string{"/csb", "/查水表 100", "/cncsb 1k", "/wlcsb 1 100"},
		ArgumentHint:   "可选：单个排名或 UID；留空时使用绑定账号",
		SearchType:     SearchTypeNone,
		RenderMode:     RenderModePreview,
	},
	{
		ID:             "forecast-ranking",
		Name:           "榜线预测",
		Description:    "查询公开预测 API 的预测/最终榜线；预测仅支持国服和日服。",
		PrimaryCommand: "榜线预测",
		Commands:       []string{"榜线预测", "sk预测"},
		Usage:          "/sk预测 [活动ID]",
		Template:       "forecast_ranking_list",
		PreviewID:      "ranking-list",
		PresetAliases:  []string{"skp"},
		Examples:       []string{"/skp", "/sk预测 165", "/jpskp"},
		ArgumentHint:   "可选：活动 ID；留空时使用当前/最近活动",
		SearchType:     SearchTypeNone,
		RenderMode:     RenderModePreview,
	},
	{
		ID:              "profile-card",
		Name:            "个人信息图",
		Description:     "查看已绑定账号的玩家资料。WebUI 可输入临时区服与游戏 UID 调试真实渲染，聊天端会按用户绑定账号查询。",
		PrimaryCommand:  "个人信息",
		Commands:        []string{"个人信息"},
		Usage:           "/个人信息",
		Template:        "profile_card",
		PreviewID:       "profile-card",
		PresetAliases:   []string{"profile"},
		Examples:        []string{"/个人信息", "/profile"},
		RequiresBinding: true,
		BindingKind:     "profile",
		BindingHint:     "输入区服与游戏 UID 后，WebUI 会使用 Sekai API 拉取资料生成真实预览；不会保存绑定。",
		SearchType:      SearchTypeNone,
		RenderMode:      RenderModePreview,
	},
	{
		ID:              "suite-status",
		Name:            "Suite 公开状态",
		Description:     "查询已绑定账号的 Haruki Suite 公开数据更新时间和来源。WebUI 可输入临时区服与游戏 UID 调试。",
		PrimaryCommand:  "抓包状态",
		Commands:        []string{"抓包状态", "抓包数据", "抓包信息"},
		Usage:           "/抓包状态",
		Template:        "suite_panel",
		PreviewID:       "suite-panel",
		PresetAliases:   []string{"suite"},
		Examples:        []string{"/抓包状态", "/cn抓包状态", "/suite"},
		RequiresBinding: true,
		BindingKind:     "suite",
		BindingHint:     "输入区服与游戏 UID 后，WebUI 会使用 Haruki 公开 API 拉取真实 Suite 数据；不会保存绑定。",
		SearchType:      SearchTypeNone,
		RenderMode:      RenderModePreview,
	},
	{
		ID:             "suite-hide",
		Name:           "隐藏抓包",
		Description:    "隐藏自己的 Suite 抓包信息。",
		PrimaryCommand: "隐藏抓包",
		Commands:       []string{"隐藏抓包"},
		Usage:          "/隐藏抓包",
		SearchType:     SearchTypeNone,
		RenderMode:     RenderModeAction,
	},
	{
		ID:             "suite-show",
		Name:           "展示抓包",
		Description:    "重新展示自己的 Suite 抓包信息。",
		PrimaryCommand: "展示抓包",
		Commands:       []string{"展示抓包", "显示抓包"},
		Usage:          "/展示抓包",
		SearchType:     SearchTypeNone,
		RenderMode:     RenderModeAction,
	},
	{
		ID:              "bond-list",
		Name:            "羁绊查询",
		Description:     "查询绑定账号的 Suite 羁绊等级 TOP 列表。WebUI 可输入临时区服与游戏 UID 调试。",
		PrimaryCommand:  "羁绊",
		Commands:        []string{"羁绊", "羁绊等级", "牵绊", "牵绊等级"},
		Usage:           "/羁绊",
		Template:        "suite_panel",
		PreviewID:       "suite-panel",
		PresetAliases:   []string{"bond", "bonds"},
		Examples:        []string{"/羁绊", "/cn羁绊"},
		RequiresBinding: true,
		BindingKind:     "suite",
		BindingHint:     "输入区服与游戏 UID 后，WebUI 会使用 Haruki 公开 API 拉取真实 Suite 数据；不会保存绑定。",
		SearchType:      SearchTypeNone,
		RenderMode:      RenderModePreview,
	},
	{
		ID:             "music-progress",
		Name:           "打歌进度 / 歌曲奖励",
		Description:    "查询绑定账号的 Suite 打歌进度与歌曲奖励剩余统计。WebUI 可输入临时区服与游戏 UID 调试。",
		PrimaryCommand: "打歌进度",
		Commands: []string{
			"打歌进度", "歌曲进度", "打歌信息",
			"歌曲奖励", "打歌奖励", "歌曲挖矿", "打歌挖矿",
		},
		Usage:           "/打歌进度 或 /歌曲奖励",
		Template:        "suite_panel",
		PreviewID:       "suite-panel",
		PresetAliases:   []string{"progress", "musicreward"},
		Examples:        []string{"/打歌进度", "/歌曲奖励", "/cn打歌进度"},
		RequiresBinding: true,
		BindingKind:     "suite",
		BindingHint:     "输入区服与游戏 UID 后，WebUI 会使用 Haruki 公开 API 拉取真实 Suite 数据；不会保存绑定。",
		SearchType:      SearchTypeNone,
		RenderMode:      RenderModePreview,
	},
	{
		ID:              "challenge-info",
		Name:            "挑战信息",
		Description:     "查询绑定账号的 Suite 挑战 Live 进度统计。WebUI 可输入临时区服与游戏 UID 调试。",
		PrimaryCommand:  "挑战信息",
		Commands:        []string{"挑战信息", "挑战等级", "挑战进度", "挑战详情", "每日挑战"},
		Usage:           "/挑战信息",
		Template:        "suite_panel",
		PreviewID:       "suite-panel",
		PresetAliases:   []string{"challenge"},
		Examples:        []string{"/挑战信息", "/cn挑战信息"},
		RequiresBinding: true,
		BindingKind:     "suite",
		BindingHint:     "输入区服与游戏 UID 后，WebUI 会使用 Haruki 公开 API 拉取真实 Suite 数据；不会保存绑定。",
		SearchType:      SearchTypeNone,
		RenderMode:      RenderModePreview,
	},
	{
		ID:              "event-record",
		Name:            "活动记录",
		Description:     "查询绑定账号的 Suite 活动 PT 和 WL 章节记录。WebUI 可输入临时区服与游戏 UID 调试。",
		PrimaryCommand:  "活动记录",
		Commands:        []string{"活动记录", "活动履历", "冲榜记录"},
		Usage:           "/活动记录",
		Template:        "suite_panel",
		PreviewID:       "suite-panel",
		PresetAliases:   []string{"eventrecord"},
		Examples:        []string{"/活动记录", "/冲榜记录", "/cn活动记录"},
		RequiresBinding: true,
		BindingKind:     "suite",
		BindingHint:     "输入区服与游戏 UID 后，WebUI 会使用 Haruki 公开 API 拉取真实 Suite 数据；不会保存绑定。",
		SearchType:      SearchTypeNone,
		RenderMode:      RenderModePreview,
	},
	{
		ID:              "leader-count",
		Name:            "队长次数",
		Description:     "查询绑定账号的 Suite 角色队长游玩次数排行。WebUI 可输入临时区服与游戏 UID 调试。",
		PrimaryCommand:  "队长次数",
		Commands:        []string{"队长次数", "角色次数", "队长游玩次数", "角色游玩次数"},
		Usage:           "/队长次数",
		Template:        "suite_panel",
		PreviewID:       "suite-panel",
		PresetAliases:   []string{"leadercount"},
		Examples:        []string{"/队长次数", "/cn队长次数"},
		RequiresBinding: true,
		BindingKind:     "suite",
		BindingHint:     "输入区服与游戏 UID 后，WebUI 会使用 Haruki 公开 API 拉取真实 Suite 数据；不会保存绑定。",
		SearchType:      SearchTypeNone,
		RenderMode:      RenderModePreview,
	},
	{
		ID:              "suite-card-box",
		Name:            "卡牌一览",
		Description:     "查询绑定账号的 Suite 卡牌持有一览，支持 box/id/before/mr/sl/time 与查卡筛选条件。WebUI 可输入临时区服与游戏 UID 调试。",
		PrimaryCommand:  "卡牌一览",
		Commands:        []string{"卡牌一览", "卡面一览", "卡一览", "持有卡牌", "我的卡牌"},
		Usage:           "/卡牌一览 [box/id/before/mr/sl/time/筛选条件]",
		Template:        "suite_card_box",
		PreviewID:       "suite-card-box",
		PresetAliases:   []string{"cardbox", "box", "mycards"},
		Examples:        []string{"/卡牌一览", "/卡牌一览 四星 限定", "/卡牌一览 box", "/卡牌一览 miku mr"},
		RequiresBinding: true,
		BindingKind:     "suite",
		BindingHint:     "输入区服与游戏 UID 后，WebUI 会使用 Haruki 公开 API 拉取真实 Suite 数据；不会保存绑定。",
		SearchType:      SearchTypeNone,
		RenderMode:      RenderModePreview,
	},
	{
		ID:             "gacha-result",
		Name:           "抽卡结果图",
		Description:    "生成娱乐向抽卡结果预览。",
		PrimaryCommand: "抽卡模拟",
		Commands:       []string{"抽卡模拟"},
		Usage:          "/抽卡模拟",
		Template:       "gacha_result",
		PreviewID:      "gacha-result",
		PresetAliases:  []string{"抽卡", "roll"},
		Examples:       []string{"/抽卡模拟", "/roll"},
		SearchType:     SearchTypeNone,
		RenderMode:     RenderModePreview,
	},
	{
		ID:             "help-card",
		Name:           "帮助菜单图",
		Description:    "显示指令帮助与功能列表。",
		PrimaryCommand: "帮助",
		Commands:       []string{"帮助"},
		Usage:          "/帮助",
		Template:       "help_card",
		PreviewID:      "help-card",
		PresetAliases:  []string{"help"},
		Examples:       []string{"/帮助", "/help"},
		SearchType:     SearchTypeNone,
		RenderMode:     RenderModePreview,
	},
}

// Definitions returns command definitions merged with user custom aliases.
func Definitions(custom map[string][]string) []Definition {
	defs := cloneDefinitions(baseDefinitions)
	cleaned, _, err := ValidateAliases(custom)
	if err != nil {
		cleaned = map[string][]string{}
	}
	for i := range defs {
		aliases := cleaned[defs[i].PrimaryCommand]
		defs[i].CustomAliases = append([]string(nil), aliases...)
		defs[i].Examples = withPrefixExamples(defs[i].Examples, "/")
	}
	return defs
}

// BaseDefinitions returns the protected built-in command definitions without custom aliases.
func BaseDefinitions() []Definition {
	return cloneDefinitions(baseDefinitions)
}

// DefinitionByID returns a definition matching id or nil.
func DefinitionByID(defs []Definition, id string) *Definition {
	for i := range defs {
		if defs[i].ID == id {
			return &defs[i]
		}
	}
	return nil
}

// DefinitionsByPrimary maps primary command names to definitions.
func DefinitionsByPrimary(defs []Definition) map[string]Definition {
	out := make(map[string]Definition, len(defs))
	for _, def := range defs {
		out[def.PrimaryCommand] = def
	}
	return out
}

// PresetAliasMap returns protected preset aliases by primary command.
func PresetAliasMap() map[string][]string {
	out := make(map[string][]string, len(baseDefinitions))
	for _, def := range baseDefinitions {
		out[def.PrimaryCommand] = append([]string(nil), def.PresetAliases...)
	}
	return out
}

// ProtectedNames returns all built-in commands and preset aliases that custom aliases cannot override.
func ProtectedNames() []string {
	seen := map[string]string{}
	for _, def := range baseDefinitions {
		for _, name := range append(append([]string{}, def.Commands...), def.PresetAliases...) {
			seen[normalizeName(name)] = strings.TrimSpace(name)
		}
	}
	for _, region := range config.RegionKeys() {
		seen[normalizeName(region)] = region
	}
	return sortedMapValues(seen)
}

// CommandNames returns all configurable base command names.
func CommandNames() []string {
	out := make([]string, 0, len(baseDefinitions))
	for _, def := range baseDefinitions {
		out = append(out, def.PrimaryCommand)
	}
	return out
}

// NormalizeAliases trims, deduplicates, and keeps only known command keys.
func NormalizeAliases(input map[string][]string) (map[string][]string, []string) {
	out := make(map[string][]string)
	warnings := []string{}
	for rawCommand, rawAliases := range input {
		command := strings.TrimSpace(rawCommand)
		primary := primaryForCommand(command)
		if primary == "" {
			warnings = append(warnings, fmt.Sprintf("未知指令 %q 已忽略", rawCommand))
			continue
		}
		seen := map[string]struct{}{}
		for _, rawAlias := range rawAliases {
			alias := sanitizeAlias(rawAlias)
			if alias == "" {
				continue
			}
			key := normalizeName(alias)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out[primary] = append(out[primary], alias)
		}
	}
	for command := range out {
		sort.SliceStable(out[command], func(i, j int) bool {
			return strings.ToLower(out[command][i]) < strings.ToLower(out[command][j])
		})
	}
	return out, warnings
}

// ValidateAliases validates user-provided custom aliases against protected names and collisions.
func ValidateAliases(input map[string][]string) (map[string][]string, []string, error) {
	cleaned, warnings := NormalizeAliases(input)
	protected := map[string]string{}
	for _, name := range ProtectedNames() {
		protected[normalizeName(name)] = name
	}

	owner := map[string]string{}
	for command, aliases := range cleaned {
		valid := make([]string, 0, len(aliases))
		for _, alias := range aliases {
			key := normalizeName(alias)
			if strings.ContainsAny(alias, " \t\r\n") {
				return nil, warnings, fmt.Errorf("自定义关键词 %q 不能包含空格", alias)
			}
			if len([]rune(alias)) < 2 {
				return nil, warnings, fmt.Errorf("自定义关键词 %q 过短，容易误触发聊天端指令", alias)
			}
			if protectedName, ok := protected[key]; ok {
				return nil, warnings, fmt.Errorf("自定义关键词 %q 与受保护指令/预设 %q 冲突", alias, protectedName)
			}
			if previous, ok := owner[key]; ok && previous != command {
				return nil, warnings, fmt.Errorf("自定义关键词 %q 同时绑定到 %s 和 %s", alias, previous, command)
			}
			owner[key] = command
			valid = append(valid, alias)
		}
		cleaned[command] = valid
	}
	return cleaned, warnings, nil
}

// AliasConfig builds a full alias editor response.
func AliasConfig(custom map[string][]string) AliasConfigResponse {
	cleaned, warnings, err := ValidateAliases(custom)
	if err != nil {
		cleaned, warnings = NormalizeAliases(custom)
		warnings = append(warnings, err.Error())
	}
	return AliasConfigResponse{
		Data:         Definitions(cleaned),
		Custom:       cleaned,
		Preset:       PresetAliasMap(),
		Protected:    ProtectedNames(),
		RiskMessage:  RiskMessage,
		RestartNote:  RestartNote,
		Warnings:     warnings,
		CommandNames: CommandNames(),
	}
}

// BotCommandsFor returns actual command names including regional variants.
func BotCommandsFor(def Definition) []BotCommand {
	names := make([]BotCommand, 0)
	add := func(name, region, source string) {
		if strings.TrimSpace(name) == "" {
			return
		}
		names = append(names, BotCommand{
			Name:          name,
			Base:          baseWithoutRegion(name),
			Primary:       def.PrimaryCommand,
			Region:        region,
			MatchSource:   source,
			DefinitionID:  def.ID,
			RequiresArg:   def.RequiresArgument,
			ArgumentHint:  def.ArgumentHint,
			CanonicalName: def.PrimaryCommand,
		})
	}

	for _, command := range def.Commands {
		add(command, "", MatchPrimary)
		for _, region := range config.RegionKeys() {
			add(region+command, region, MatchPrimary)
		}
	}
	for _, alias := range def.PresetAliases {
		add(alias, "", MatchPresetAlias)
		for _, region := range config.RegionKeys() {
			add(region+alias, region, MatchPresetAlias)
		}
	}
	for _, alias := range def.CustomAliases {
		add(alias, "", MatchCustomAlias)
		for _, region := range config.RegionKeys() {
			add(region+alias, region, MatchCustomAlias)
		}
	}
	return dedupeBotCommands(names)
}

// AllBotCommands returns actual command names for all definitions.
func AllBotCommands(defs []Definition) []BotCommand {
	out := []BotCommand{}
	for _, def := range defs {
		out = append(out, BotCommandsFor(def)...)
	}
	return dedupeBotCommands(out)
}

func primaryForCommand(command string) string {
	command = strings.TrimSpace(command)
	for _, def := range baseDefinitions {
		if def.PrimaryCommand == command {
			return def.PrimaryCommand
		}
		for _, alias := range def.Commands {
			if alias == command {
				return def.PrimaryCommand
			}
		}
	}
	return ""
}

func cloneDefinitions(defs []Definition) []Definition {
	out := make([]Definition, len(defs))
	for i, def := range defs {
		out[i] = def
		out[i].Commands = append([]string(nil), def.Commands...)
		out[i].PresetAliases = append([]string(nil), def.PresetAliases...)
		out[i].CustomAliases = append([]string(nil), def.CustomAliases...)
		out[i].Examples = append([]string(nil), def.Examples...)
	}
	return out
}

func sanitizeAlias(raw string) string {
	return strings.TrimPrefix(strings.TrimSpace(raw), "/")
}

func normalizeName(value string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(value), "/"))
}

func baseWithoutRegion(command string) string {
	for _, region := range config.RegionKeys() {
		if strings.HasPrefix(command, region) && len(command) > len(region) {
			return strings.TrimPrefix(command, region)
		}
	}
	return command
}

func dedupeBotCommands(commands []BotCommand) []BotCommand {
	seen := map[string]struct{}{}
	out := make([]BotCommand, 0, len(commands))
	for _, command := range commands {
		key := normalizeName(command.Name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, command)
	}
	return out
}

func sortedMapValues(values map[string]string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i]) < strings.ToLower(out[j])
	})
	return out
}

func withPrefixExamples(examples []string, prefix string) []string {
	out := make([]string, len(examples))
	for i, example := range examples {
		if prefix == "/" || strings.HasPrefix(example, prefix) || strings.HasPrefix(example, "/") {
			out[i] = example
			continue
		}
		out[i] = prefix + strings.TrimPrefix(example, "/")
	}
	return out
}
