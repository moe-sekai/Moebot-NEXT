package commandparser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"moebot-next/internal/assets"
	"moebot-next/internal/config"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/renderer"
	"moebot-next/internal/servers"
)

// EntityResult is a lightweight result row returned by command parsing.
type EntityResult struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Type     string `json:"type"`
	Payload  any    `json:"-"`
}

// ParseResult describes how a raw input maps to a command definition.
type ParseResult struct {
	RawInput                 string         `json:"raw_input"`
	CommandPrefix            string         `json:"command_prefix"`
	CommandText              string         `json:"command_text"`
	MatchedCommand           string         `json:"matched_command"`
	MatchedBase              string         `json:"matched_base"`
	MatchSource              string         `json:"match_source"`
	Region                   string         `json:"region"`
	RegionLabel              string         `json:"region_label"`
	Argument                 string         `json:"argument"`
	Definition               *Definition    `json:"definition,omitempty"`
	Results                  []EntityResult `json:"results"`
	Selected                 *EntityResult  `json:"selected,omitempty"`
	CanRender                bool           `json:"can_render"`
	RenderMode               string         `json:"render_mode"`
	PreviewFallbackAvailable bool           `json:"preview_fallback_available"`
	Message                  string         `json:"message"`
	Warnings                 []string       `json:"warnings"`
	Suggestions              []string       `json:"suggestions"`
}

// DefinitionsResponse is returned by the command definitions API.
type DefinitionsResponse struct {
	Data          []Definition `json:"data"`
	Total         int          `json:"total"`
	CommandPrefix string       `json:"command_prefix"`
	Regions       []RegionInfo `json:"regions"`
	RiskMessage   string       `json:"risk_message"`
	RestartNote   string       `json:"restart_note"`
}

// RegionInfo describes a supported server region for WebUI.
type RegionInfo struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// ParseResponse is returned by the parse API.
type ParseResponse struct {
	OK      bool        `json:"ok"`
	Parsed  ParseResult `json:"parsed"`
	Message string      `json:"message"`
}

// Service parses command text and builds renderer requests.
type Service struct {
	CommandPrefix string
	Definitions   []Definition
	Servers       *servers.Manager
	Store         *masterdata.Store
	Renderer      *renderer.Client
}

// NewService creates a parser service with merged definitions.
func NewService(prefix string, customAliases map[string][]string, manager *servers.Manager, store *masterdata.Store, rendererClient *renderer.Client) *Service {
	if strings.TrimSpace(prefix) == "" {
		prefix = "/"
	}
	return &Service{
		CommandPrefix: prefix,
		Definitions:   Definitions(customAliases),
		Servers:       manager,
		Store:         store,
		Renderer:      rendererClient,
	}
}

// DefinitionsPayload returns metadata for the WebUI.
func (s *Service) DefinitionsPayload() DefinitionsResponse {
	regions := make([]RegionInfo, 0, len(config.RegionKeys()))
	for _, region := range config.RegionKeys() {
		regions = append(regions, RegionInfo{Key: region, Label: config.RegionLabel(region)})
	}
	return DefinitionsResponse{
		Data:          s.Definitions,
		Total:         len(s.Definitions),
		CommandPrefix: s.CommandPrefix,
		Regions:       regions,
		RiskMessage:   RiskMessage,
		RestartNote:   RestartNote,
	}
}

// Parse maps a raw text command to a definition and optional search result.
func (s *Service) Parse(input string) ParseResult {
	result := ParseResult{
		RawInput:      input,
		CommandPrefix: s.CommandPrefix,
		Region:        s.defaultRegion(),
		RegionLabel:   config.RegionLabel(s.defaultRegion()),
		Results:       []EntityResult{},
		Suggestions:   s.defaultSuggestions(),
	}

	text := strings.TrimSpace(input)
	if text == "" {
		result.Message = "请输入要解析的命令，例如 /查卡 1204。"
		return result
	}

	body := s.stripCommandPrefix(text)
	commandText, argument := splitCommand(body)
	result.CommandText = commandText
	result.Argument = argument
	if commandText == "" {
		result.Message = "没有读取到指令名，请输入类似 /查卡 1204 的内容。"
		return result
	}

	match, ok := s.matchCommand(commandText)
	if !ok {
		result.Message = fmt.Sprintf("未匹配到指令「%s」。", commandText)
		result.Suggestions = s.suggestionsFor(commandText)
		return result
	}

	def := match.Definition
	result.MatchedCommand = match.Name
	result.MatchedBase = match.Base
	result.MatchSource = match.Source
	if match.Region != "" {
		result.Region = match.Region
		result.RegionLabel = config.RegionLabel(match.Region)
	}
	result.Definition = &def
	result.RenderMode = def.RenderMode
	result.PreviewFallbackAvailable = def.PreviewID != ""

	if def.RequiresArgument && argument == "" {
		result.Message = def.ArgumentHint
		result.Warnings = append(result.Warnings, "该功能需要参数，暂不渲染真实数据。")
		result.CanRender = def.RenderMode == RenderModePreview && def.PreviewID != ""
		return result
	}

	if def.RenderMode == RenderModePreview || def.SearchType == SearchTypeNone {
		result.CanRender = def.PreviewID != ""
		result.Message = "该功能会使用 Satori 静态样例预览；聊天端会按真实上下文执行。"
		return result
	}

	store, assetResolver, storeMessage := s.storeForRegion(result.Region)
	if storeMessage != "" {
		result.Message = storeMessage
		result.CanRender = def.PreviewID != ""
		result.Warnings = append(result.Warnings, "无法命中真实数据，将使用静态预览兜底。")
		return result
	}

	results, selected := searchAndBuild(def, store, assetResolver, argument)
	result.Results = results
	if selected != nil {
		result.Selected = selected
		result.CanRender = true
		result.Message = fmt.Sprintf("已解析为「%s」，并命中 %s #%d。", def.Name, selected.Type, selected.ID)
		return result
	}

	result.Message = fmt.Sprintf("已解析为「%s」，但没有找到与「%s」匹配的数据。", def.Name, argument)
	result.CanRender = def.PreviewID != ""
	result.Warnings = append(result.Warnings, "没有搜索命中，将使用静态预览兜底。")
	return result
}

// Render renders the parsed command. It uses real search data when available and falls back to preview images.
func (s *Service) Render(input string, width int, height int) (*renderer.PreviewRenderResult, ParseResult, error) {
	parsed := s.Parse(input)
	if s.Renderer == nil {
		return nil, parsed, fmt.Errorf("renderer client is not configured")
	}
	if parsed.Definition == nil {
		return nil, parsed, errors.New(parsed.Message)
	}
	def := *parsed.Definition
	if parsed.Selected != nil && parsed.Selected.Payload != nil && def.Template != "" {
		request := renderer.RenderRequest{Template: def.Template, Data: parsed.Selected.Payload, Width: width, Height: height}
		result, err := s.Renderer.RenderWithTrace(request)
		if err == nil {
			return result, parsed, nil
		}
		parsed.Warnings = append(parsed.Warnings, "真实数据渲染失败，尝试使用静态预览兜底。")
	}
	if def.PreviewID == "" {
		return nil, parsed, fmt.Errorf("该功能没有可用的静态预览")
	}
	result, err := s.Renderer.RenderPreviewWithTrace(def.PreviewID, width, height)
	return result, parsed, err
}

type commandMatch struct {
	Name       string
	Base       string
	Region     string
	Source     string
	Definition Definition
}

func (s *Service) matchCommand(commandText string) (commandMatch, bool) {
	if match, ok := s.matchBase(commandText, ""); ok {
		return match, true
	}
	for _, region := range config.RegionKeys() {
		if strings.HasPrefix(commandText, region) && len(commandText) > len(region) {
			base := strings.TrimPrefix(commandText, region)
			if match, ok := s.matchBase(base, region); ok {
				match.Name = commandText
				match.Region = region
				return match, true
			}
		}
	}
	return commandMatch{}, false
}

func (s *Service) matchBase(base string, region string) (commandMatch, bool) {
	baseKey := normalizeName(base)
	for _, def := range s.Definitions {
		for _, command := range def.Commands {
			if normalizeName(command) == baseKey {
				return commandMatch{Name: base, Base: command, Region: region, Source: MatchPrimary, Definition: def}, true
			}
		}
		for _, alias := range def.PresetAliases {
			if normalizeName(alias) == baseKey {
				return commandMatch{Name: base, Base: alias, Region: region, Source: MatchPresetAlias, Definition: def}, true
			}
		}
		for _, alias := range def.CustomAliases {
			if normalizeName(alias) == baseKey {
				return commandMatch{Name: base, Base: alias, Region: region, Source: MatchCustomAlias, Definition: def}, true
			}
		}
	}
	return commandMatch{}, false
}

func (s *Service) stripCommandPrefix(text string) string {
	prefix := strings.TrimSpace(s.CommandPrefix)
	if prefix != "" && strings.HasPrefix(text, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(text, prefix))
	}
	if strings.HasPrefix(text, "/") {
		return strings.TrimSpace(strings.TrimPrefix(text, "/"))
	}
	return text
}

func splitCommand(body string) (string, string) {
	fields := strings.Fields(body)
	if len(fields) == 0 {
		return "", ""
	}
	command := fields[0]
	argument := strings.TrimSpace(strings.TrimPrefix(body, command))
	return command, argument
}

func (s *Service) defaultRegion() string {
	if s.Servers != nil {
		return s.Servers.DefaultRegion()
	}
	return config.RegionJP
}

func (s *Service) storeForRegion(region string) (*masterdata.Store, *assets.Resolver, string) {
	if s.Servers != nil {
		runtime := s.Servers.Get(region)
		if runtime == nil || runtime.Store == nil || !runtime.Enabled {
			return nil, nil, runtimeUnavailableMessage(runtime)
		}
		if !runtime.Store.IsLoaded() {
			return runtime.Store, runtime.Assets, fmt.Sprintf("%s Masterdata 尚未加载，暂时无法搜索。", runtime.Label)
		}
		return runtime.Store, runtime.Assets, ""
	}
	if s.Store == nil || !s.Store.IsLoaded() {
		return s.Store, nil, "Masterdata 尚未加载，暂时无法搜索。"
	}
	return s.Store, nil, ""
}

func runtimeUnavailableMessage(runtime *servers.Runtime) string {
	if runtime == nil {
		return "服务器配置不可用。"
	}
	if !runtime.Enabled {
		return fmt.Sprintf("%s 暂未启用。", runtime.Label)
	}
	return "服务器配置不可用。"
}

func searchAndBuild(def Definition, store *masterdata.Store, resolver *assets.Resolver, argument string) ([]EntityResult, *EntityResult) {
	switch def.SearchType {
	case SearchTypeCard:
		cards := store.SearchCards(argument)
		rows := make([]EntityResult, 0, len(cards))
		for _, card := range cards {
			rows = append(rows, EntityResult{ID: card.ID, Title: card.Prefix, Subtitle: fmt.Sprintf("角色 #%d · %s", card.CharacterID, card.CardRarityType), Type: "card"})
		}
		if len(cards) == 0 {
			return rows, nil
		}
		payload := renderer.BuildCardDetailPayloadWithAssets(store, cards[0], resolver)
		selected := rows[0]
		selected.Payload = payload
		return rows, &selected
	case SearchTypeMusic:
		musics := store.SearchMusics(argument)
		rows := make([]EntityResult, 0, len(musics))
		for _, music := range musics {
			rows = append(rows, EntityResult{ID: music.ID, Title: music.Title, Subtitle: strings.Join(nonEmpty(music.Composer, music.Lyricist, music.Arranger), " / "), Type: "music"})
		}
		if len(musics) == 0 {
			return rows, nil
		}
		payload := renderer.BuildMusicDetailPayloadWithAssets(store, musics[0], resolver)
		selected := rows[0]
		selected.Payload = payload
		return rows, &selected
	case SearchTypeEvent:
		events := store.SearchEvents(argument)
		rows := make([]EntityResult, 0, len(events))
		for _, event := range events {
			rows = append(rows, EntityResult{ID: event.ID, Title: event.Name, Subtitle: fmt.Sprintf("%s · %s", event.EventType, event.Unit), Type: "event"})
		}
		if len(events) == 0 {
			return rows, nil
		}
		payload := renderer.BuildEventInfoPayloadWithAssets(store, events[0], resolver)
		selected := rows[0]
		selected.Payload = payload
		return rows, &selected
	case SearchTypeGacha:
		gachas := store.SearchGachas(argument)
		rows := make([]EntityResult, 0, len(gachas))
		for _, gacha := range gachas {
			rows = append(rows, EntityResult{ID: gacha.ID, Title: gacha.Name, Subtitle: gachaSubtitle(gacha), Type: "gacha"})
		}
		if len(gachas) == 0 {
			return rows, nil
		}
		payload := renderer.BuildGachaInfoPayloadWithAssets(store, gachas[0], resolver)
		selected := rows[0]
		selected.Payload = payload
		return rows, &selected
	default:
		return []EntityResult{}, nil
	}
}

func gachaSubtitle(gacha masterdata.GachaInfo) string {
	parts := nonEmpty(gacha.GachaType)
	if gacha.StartAt > 0 || gacha.EndAt > 0 {
		parts = append(parts, fmt.Sprintf("%s - %s", formatMillis(gacha.StartAt), formatMillis(gacha.EndAt)))
	}
	return strings.Join(parts, " · ")
}

func nonEmpty(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func formatMillis(ms int64) string {
	if ms <= 0 {
		return "-"
	}
	return time.Unix(0, ms*int64(time.Millisecond)).Format("2006-01-02 15:04")
}

func (s *Service) defaultSuggestions() []string {
	out := make([]string, 0, len(s.Definitions))
	for _, def := range s.Definitions {
		if len(def.Examples) > 0 {
			out = append(out, def.Examples[0])
		}
	}
	return out
}

func (s *Service) suggestionsFor(commandText string) []string {
	needle := strings.ToLower(commandText)
	out := []string{}
	for _, def := range s.Definitions {
		names := append(append([]string{}, def.Commands...), def.PresetAliases...)
		names = append(names, def.CustomAliases...)
		for _, name := range names {
			if strings.Contains(strings.ToLower(name), needle) || strings.Contains(needle, strings.ToLower(name)) {
				out = append(out, firstExample(def))
				break
			}
		}
	}
	if len(out) == 0 {
		return s.defaultSuggestions()
	}
	return out
}

func firstExample(def Definition) string {
	if len(def.Examples) > 0 {
		return def.Examples[0]
	}
	return def.Usage
}

func ParseWidth(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
