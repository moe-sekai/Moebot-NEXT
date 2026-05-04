package commandparser

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"moebot-next/internal/assets"
	"moebot-next/internal/cardquery"
	"moebot-next/internal/config"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/ranking"
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
		if inlineCommand, inlineArgument, inlineMatch, inlineOK := s.matchInlineArgumentCommand(body); inlineOK {
			commandText = inlineCommand
			argument = inlineArgument
			match = inlineMatch
			ok = true
			result.CommandText = commandText
			result.Argument = argument
		} else {
			result.Message = fmt.Sprintf("未匹配到指令「%s」。", commandText)
			result.Suggestions = s.suggestionsFor(commandText)
			return result
		}
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

	if payloadResult, ok := s.buildRealtimePayload(def, result.Region, argument); ok {
		result.Results = payloadResult.Results
		result.Selected = payloadResult.Selected
		result.CanRender = payloadResult.Selected != nil && payloadResult.Selected.Payload != nil
		result.Message = payloadResult.Message
		result.Warnings = append(result.Warnings, payloadResult.Warnings...)
		return result
	}

	if def.RenderMode == RenderModePreview || def.SearchType == SearchTypeNone {
		result.CanRender = def.PreviewID != ""
		result.Message = "该功能暂未接入实时解析数据，将使用 Satori 静态样例预览；聊天端会按真实上下文执行。"
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
		if strings.HasSuffix(selected.Type, "_list") {
			result.Message = fmt.Sprintf("已解析为「%s」，并生成 %s（%d 条候选）。", def.Name, selected.Title, len(results))
		} else {
			result.Message = fmt.Sprintf("已解析为「%s」，并命中 %s #%d。", def.Name, selected.Type, selected.ID)
		}
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
		template := templateForPayload(def.Template, parsed.Selected.Payload)
		request := renderer.RenderRequest{Template: template, Data: parsed.Selected.Payload, Width: width, Height: height}
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
		if strings.HasPrefix(commandText, region+"wl") && len(commandText) > len(region)+len("wl") {
			base := strings.TrimPrefix(commandText, region+"wl")
			if match, ok := s.matchBase(base, region); ok && isWorldLinkInlineDefinitionID(match.Definition.ID) {
				match.Name = commandText
				match.Base = "wl" + match.Base
				match.Region = region
				return match, true
			}
		}
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

func (s *Service) matchInlineArgumentCommand(body string) (string, string, commandMatch, bool) {
	body = strings.TrimSpace(body)
	if body == "" {
		return "", "", commandMatch{}, false
	}
	for _, candidate := range s.inlineArgumentCommandCandidates() {
		if !hasCommandPrefixFold(body, candidate.Name) {
			continue
		}
		rest := strings.TrimSpace(body[len(candidate.Name):])
		if !isInlineArgumentAllowed(candidate.Definition.ID, rest) {
			continue
		}
		match := candidate.Match
		match.Name = body[:len(candidate.Name)]
		return match.Name, rest, match, true
	}
	return "", "", commandMatch{}, false
}

type inlineArgumentCommandCandidate struct {
	Name       string
	Definition Definition
	Match      commandMatch
}

func (s *Service) inlineArgumentCommandCandidates() []inlineArgumentCommandCandidate {
	candidates := []inlineArgumentCommandCandidate{}
	add := func(name string, def Definition, source string, region string, base string) {
		name = strings.TrimSpace(name)
		if name == "" || !isInlineArgumentDefinitionID(def.ID) {
			return
		}
		match := commandMatch{
			Name:       name,
			Base:       base,
			Region:     region,
			Source:     source,
			Definition: def,
		}
		candidates = append(candidates, inlineArgumentCommandCandidate{Name: name, Definition: def, Match: match})
	}
	for _, def := range s.Definitions {
		for _, command := range def.Commands {
			add(command, def, MatchPrimary, "", command)
			for _, region := range config.RegionKeys() {
				add(region+command, def, MatchPrimary, region, command)
			}
		}
		for _, alias := range def.PresetAliases {
			add(alias, def, MatchPresetAlias, "", alias)
			for _, region := range config.RegionKeys() {
				add(region+alias, def, MatchPresetAlias, region, alias)
			}
		}
		for _, alias := range def.CustomAliases {
			add(alias, def, MatchCustomAlias, "", alias)
			for _, region := range config.RegionKeys() {
				add(region+alias, def, MatchCustomAlias, region, alias)
			}
		}
	}
	return sortInlineArgumentCommandCandidates(candidates)
}

func sortInlineArgumentCommandCandidates(candidates []inlineArgumentCommandCandidate) []inlineArgumentCommandCandidate {
	sort.SliceStable(candidates, func(i, j int) bool {
		if len(candidates[i].Name) != len(candidates[j].Name) {
			return len(candidates[i].Name) > len(candidates[j].Name)
		}
		priorityI := inlineArgumentDefinitionPriority(candidates[i].Definition.ID)
		priorityJ := inlineArgumentDefinitionPriority(candidates[j].Definition.ID)
		if priorityI != priorityJ {
			return priorityI < priorityJ
		}
		return strings.ToLower(candidates[i].Name) < strings.ToLower(candidates[j].Name)
	})
	return candidates
}

func inlineArgumentDefinitionPriority(id string) int {
	switch id {
	case "ranking-list", "forecast-ranking":
		return 0
	case "water-table", "churn-ranking":
		return 1
	case "ranking-target":
		return 2
	default:
		return 10
	}
}

func isInlineArgumentDefinitionID(id string) bool {
	switch id {
	case "ranking-list", "ranking-target", "churn-ranking", "water-table", "forecast-ranking":
		return true
	default:
		return false
	}
}

func isWorldLinkInlineDefinitionID(id string) bool {
	switch id {
	case "ranking-list", "ranking-target", "churn-ranking", "water-table":
		return true
	default:
		return false
	}
}

func isInlineArgumentAllowed(definitionID string, rest string) bool {
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return true
	}
	switch definitionID {
	case "ranking-list", "ranking-target", "churn-ranking", "water-table", "forecast-ranking":
		return startsWithRankingArgument(rest)
	default:
		return true
	}
}

func startsWithRankingArgument(value string) bool {
	if value == "" {
		return false
	}
	first, _ := utf8.DecodeRuneInString(value)
	return first == '#' || first == '+' || first == '-' || first == '＋' || first == '－' || (first >= '0' && first <= '9')
}

func hasCommandPrefixFold(value string, command string) bool {
	if len(value) < len(command) {
		return false
	}
	return strings.EqualFold(value[:len(command)], command)
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

type realtimePayloadResult struct {
	Results  []EntityResult
	Selected *EntityResult
	Message  string
	Warnings []string
}

func (s *Service) buildRealtimePayload(def Definition, region string, argument string) (realtimePayloadResult, bool) {
	if !isRankingDefinition(def.ID) {
		return realtimePayloadResult{}, false
	}
	runtime := s.runtimeForRegion(region)
	if runtime == nil || !runtime.Enabled || runtime.Ranking == nil {
		return realtimePayloadResult{Message: runtimeUnavailableMessage(runtime), Warnings: []string{"榜线 API 未配置，将使用静态预览兜底。"}}, true
	}
	store := runtime.Store
	resolver := runtime.Assets
	switch def.ID {
	case "ranking-list":
		board, err := runtime.Ranking.GetLatest()
		if err != nil {
			return realtimePayloadResult{Message: "榜线获取失败：" + err.Error(), Warnings: []string{"实时数据获取失败，将使用静态预览兜底。"}}, true
		}
		view := *board
		view.Rankings = filterRankingEntriesForParser(board.Rankings, defaultRankingTiersForParser())
		if len(view.Rankings) == 0 {
			view.Rankings = firstRankingEntriesForParser(board.Rankings, 10)
		}
		payload := renderer.BuildRankingListPayloadWithStore("活动榜线", view, store, resolver)
		payload.Subtitle = rankingSubtitleForParser(region, view)
		selected := EntityResult{ID: view.EventID, Title: "实时榜线", Subtitle: payload.Subtitle, Type: "ranking_list", Payload: payload}
		return realtimePayloadResult{Results: rankingRowsForParser(view.Rankings), Selected: &selected, Message: fmt.Sprintf("已获取 %s 真实榜线数据（%d 条）。", config.RegionLabel(region), len(view.Rankings))}, true
	case "ranking-target":
		board, err := runtime.Ranking.GetLatest()
		if err != nil {
			return realtimePayloadResult{Message: "sk 获取失败：" + err.Error(), Warnings: []string{"实时数据获取失败，将使用静态预览兜底。"}}, true
		}
		entries := resolveRankingArgumentForParser(board.Rankings, argument, false)
		if len(entries) == 0 {
			entries = firstRankingEntriesForParser(board.Rankings, 1)
		}
		view := *board
		view.Rankings = entries
		payload := renderer.BuildRankingListPayloadWithStore("sk", view, store, resolver)
		payload.Subtitle = rankingSubtitleForParser(region, view)
		selected := EntityResult{ID: view.EventID, Title: "实时 sk", Subtitle: payload.Subtitle, Type: "ranking_list", Payload: payload}
		return realtimePayloadResult{Results: rankingRowsForParser(view.Rankings), Selected: &selected, Message: fmt.Sprintf("已获取 %s 真实 sk 数据（%d 条）。", config.RegionLabel(region), len(view.Rankings))}, true
	case "churn-ranking":
		board, err := runtime.Ranking.GetChurn()
		if err != nil {
			return realtimePayloadResult{Message: "查房获取失败：" + err.Error(), Warnings: []string{"实时数据获取失败，将使用静态预览兜底。"}}, true
		}
		s.hydrateChurnBoardAvatars(runtime, board)
		entries := resolveRankingArgumentForParser(board.Rankings, argument, false)
		if len(entries) == 0 {
			entries = filterRankingEntriesForParser(board.Rankings, defaultRankingTiersForParser())
		}
		if len(entries) == 0 {
			entries = firstRankingEntriesForParser(board.Rankings, 10)
		}
		view := *board
		view.Rankings = entries
		payload := renderer.BuildChurnRankingListPayloadWithStore(view, store, resolver)
		payload.Title = "查房"
		payload.Subtitle = rankingSubtitleForParser(region, view)
		selected := EntityResult{ID: view.EventID, Title: "实时查房", Subtitle: payload.Subtitle, Type: "churn_ranking_list", Payload: payload}
		return realtimePayloadResult{Results: rankingRowsForParser(view.Rankings), Selected: &selected, Message: fmt.Sprintf("已获取 %s 真实查房数据（%d 条）。", config.RegionLabel(region), len(view.Rankings))}, true
	case "water-table":
		board, err := runtime.Ranking.GetChurn()
		if err != nil {
			return realtimePayloadResult{Message: "查水表获取失败：" + err.Error(), Warnings: []string{"实时数据获取失败，将使用静态预览兜底。"}}, true
		}
		s.hydrateChurnBoardAvatars(runtime, board)
		entries := resolveRankingArgumentForParser(board.Rankings, argument, true)
		if len(entries) == 0 {
			entries = firstRankingEntriesForParser(board.Rankings, 1)
		}
		if len(entries) == 0 {
			return realtimePayloadResult{Message: "没有可用查水表数据。", Warnings: []string{"实时数据为空，将使用静态预览兜底。"}}, true
		}
		payload := renderer.BuildWaterTablePayloadWithStore(*board, entries[0], store, resolver)
		payload.Subtitle = rankingSubtitleForParser(region, *board)
		selected := EntityResult{ID: entries[0].Rank, Title: "实时查水表", Subtitle: payload.Subtitle, Type: "water_table", Payload: payload}
		return realtimePayloadResult{Results: rankingRowsForParser(entries), Selected: &selected, Message: fmt.Sprintf("已获取 %s 真实查水表数据。", config.RegionLabel(region))}, true
	case "forecast-ranking":
		if region != config.RegionCN && region != config.RegionJP {
			return realtimePayloadResult{Message: "榜线预测仅支持国服/日服。", Warnings: []string{"当前区服不支持预测，将使用静态预览兜底。"}}, true
		}
		events, err := runtime.Ranking.GetForecastEvents()
		if err != nil {
			return realtimePayloadResult{Message: "预测活动列表获取失败：" + err.Error(), Warnings: []string{"实时数据获取失败，将使用静态预览兜底。"}}, true
		}
		event := selectForecastEventForParser(events, argument)
		if event.EventID == 0 {
			return realtimePayloadResult{Message: "没有可用预测活动。", Warnings: []string{"实时数据为空，将使用静态预览兜底。"}}, true
		}
		board, err := runtime.Ranking.GetForecastLatest(event.EventID)
		if err != nil {
			return realtimePayloadResult{Message: "预测榜线获取失败：" + err.Error(), Warnings: []string{"实时数据获取失败，将使用静态预览兜底。"}}, true
		}
		payload := renderer.BuildForecastRankingPayload(*board, event.Name, region, config.RegionLabel(region))
		selected := EntityResult{ID: event.EventID, Title: "真实榜线预测", Subtitle: event.Name, Type: "forecast_ranking_list", Payload: payload}
		return realtimePayloadResult{Results: forecastRowsForParser(board.Items), Selected: &selected, Message: fmt.Sprintf("已获取 %s 真实预测数据（Event #%d）。", config.RegionLabel(region), event.EventID)}, true
	}
	return realtimePayloadResult{}, false
}

func (s *Service) runtimeForRegion(region string) *servers.Runtime {
	if s.Servers == nil {
		return nil
	}
	return s.Servers.Get(region)
}

func (s *Service) hydrateChurnBoardAvatars(runtime *servers.Runtime, churn *ranking.Board) {
	if runtime == nil || runtime.Ranking == nil || churn == nil || len(churn.Rankings) == 0 {
		return
	}
	latest, err := runtime.Ranking.GetLatest()
	if err != nil || latest == nil || len(latest.Rankings) == 0 {
		return
	}
	hydrateRankingEntriesFromLatestForParser(churn.Rankings, latest.Rankings)
}

func hydrateRankingEntriesFromLatestForParser(churn []ranking.RankingEntry, latest []ranking.RankingEntry) {
	byUID := make(map[string]ranking.RankingEntry, len(latest))
	byRank := make(map[int]ranking.RankingEntry, len(latest))
	for _, entry := range latest {
		if uid := entry.UserID.String(); uid != "" {
			byUID[uid] = entry
		}
		byRank[entry.Rank] = entry
	}
	for i := range churn {
		if churn[i].LeaderCard != nil {
			continue
		}
		var src ranking.RankingEntry
		if uid := churn[i].UserID.String(); uid != "" {
			src = byUID[uid]
		}
		if src.LeaderCard == nil && churn[i].Rank > 0 {
			src = byRank[churn[i].Rank]
		}
		if src.LeaderCard == nil {
			continue
		}
		churn[i].LeaderCard = src.LeaderCard
		if churn[i].Name == "" {
			churn[i].Name = src.Name
		}
		if churn[i].Word == "" {
			churn[i].Word = src.Word
		}
	}
}

func isRankingDefinition(id string) bool {
	switch id {
	case "ranking-list", "ranking-target", "churn-ranking", "water-table", "forecast-ranking":
		return true
	default:
		return false
	}
}

func templateForPayload(fallback string, payload any) string {
	switch payload.(type) {
	case renderer.CardListPayload:
		return "card_list"
	case renderer.MusicListPayload:
		return "music_list"
	case renderer.EventListPayload:
		return "event_list"
	case renderer.GachaListPayload:
		return "gacha_list"
	case renderer.VirtualLiveListPayload:
		return "virtual_live_list"
	case renderer.WaterTablePayload:
		return "water_table"
	case renderer.ForecastRankingPayload:
		return "forecast_ranking_list"
	default:
		return fallback
	}
}

func defaultRankingTiersForParser() []int {
	return []int{10, 20, 30, 40, 50, 100, 200, 300, 400, 500, 1000, 2000, 3000, 4000, 5000, 10000, 20000, 30000, 40000, 50000, 100000}
}

func resolveRankingArgumentForParser(entries []ranking.RankingEntry, argument string, single bool) []ranking.RankingEntry {
	argument = strings.TrimSpace(argument)
	if argument == "" {
		return nil
	}
	var ranks []int
	var uid string
	for _, token := range strings.Fields(argument) {
		if strings.Contains(token, "-") {
			items, ok := parseRankingRangeForParser(token)
			if ok {
				ranks = append(ranks, items...)
				continue
			}
		}
		if rank, ok := parseRankingNumberForParser(token); ok {
			ranks = append(ranks, rank)
			continue
		}
		if looksLikeUIDForParser(token) {
			uid = token
		}
	}
	if uid != "" {
		for _, entry := range entries {
			if entry.UserID.String() == uid {
				return []ranking.RankingEntry{entry}
			}
		}
	}
	if single && len(ranks) > 1 {
		ranks = ranks[:1]
	}
	return filterRankingEntriesForParser(entries, ranks)
}

func filterRankingEntriesForParser(entries []ranking.RankingEntry, ranks []int) []ranking.RankingEntry {
	out := make([]ranking.RankingEntry, 0, len(ranks))
	for _, rank := range dedupeRankingNumbersForParser(ranks) {
		if rank <= 0 || len(entries) == 0 {
			continue
		}
		out = append(out, nearestRankingEntryForParser(entries, rank))
	}
	return out
}

func firstRankingEntriesForParser(entries []ranking.RankingEntry, limit int) []ranking.RankingEntry {
	if len(entries) == 0 || limit <= 0 {
		return nil
	}
	if len(entries) < limit {
		limit = len(entries)
	}
	return append([]ranking.RankingEntry(nil), entries[:limit]...)
}

func nearestRankingEntryForParser(entries []ranking.RankingEntry, rank int) ranking.RankingEntry {
	best := entries[0]
	bestDiff := absIntForParser(best.Rank - rank)
	for _, entry := range entries[1:] {
		diff := absIntForParser(entry.Rank - rank)
		if diff < bestDiff {
			best = entry
			bestDiff = diff
		}
	}
	return best
}

func parseRankingRangeForParser(token string) ([]int, bool) {
	parts := strings.Split(token, "-")
	if len(parts) != 2 {
		return nil, false
	}
	start, ok1 := parseRankingNumberForParser(parts[0])
	end, ok2 := parseRankingNumberForParser(parts[1])
	if !ok1 || !ok2 || start <= 0 || end <= 0 {
		return nil, false
	}
	if start > end {
		start, end = end, start
	}
	if end-start > 19 {
		end = start + 19
	}
	out := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		out = append(out, i)
	}
	return out, true
}

func parseRankingNumberForParser(token string) (int, bool) {
	value := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(token, "#")))
	if value == "" {
		return 0, false
	}
	multiplier := 1
	if strings.HasSuffix(value, "w") || strings.HasSuffix(value, "万") {
		multiplier = 10000
		value = strings.TrimSuffix(strings.TrimSuffix(value, "w"), "万")
	} else if strings.HasSuffix(value, "k") || strings.HasSuffix(value, "千") {
		multiplier = 1000
		value = strings.TrimSuffix(strings.TrimSuffix(value, "k"), "千")
	}
	number, err := strconv.Atoi(value)
	if err != nil || number <= 0 {
		return 0, false
	}
	return number * multiplier, true
}

func looksLikeUIDForParser(token string) bool {
	if len(token) < 8 {
		return false
	}
	for _, r := range token {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func dedupeRankingNumbersForParser(values []int) []int {
	seen := map[int]struct{}{}
	out := make([]int, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func rankingRowsForParser(entries []ranking.RankingEntry) []EntityResult {
	rows := make([]EntityResult, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, EntityResult{ID: entry.Rank, Title: fmt.Sprintf("#%d %s", entry.Rank, firstNonEmptyForParser(entry.Name, "Unknown")), Subtitle: fmt.Sprintf("%dP · UID %s", entry.Score, entry.UserID.String()), Type: "ranking"})
	}
	return rows
}

func forecastRowsForParser(items []ranking.ForecastItem) []EntityResult {
	rows := make([]EntityResult, 0, len(items))
	for _, item := range items {
		prediction, ok := item.PredictedScore()
		subtitle := fmt.Sprintf("当前 %dP", item.Score)
		if item.IsFinal {
			subtitle += " · 最终线"
		} else if ok {
			subtitle += fmt.Sprintf(" · 预测 %dP", prediction)
		}
		rows = append(rows, EntityResult{ID: item.Rank, Title: fmt.Sprintf("#%d", item.Rank), Subtitle: subtitle, Type: "forecast_ranking"})
	}
	return rows
}

func selectForecastEventForParser(events []ranking.ForecastEvent, argument string) ranking.ForecastEvent {
	if id, ok := parseRankingNumberForParser(argument); ok {
		for _, event := range events {
			if event.EventID == id {
				return event
			}
		}
		return ranking.ForecastEvent{EventID: id, Name: fmt.Sprintf("Event %d", id)}
	}
	for _, event := range events {
		if event.Status == "active" && event.HasRealtimeData {
			return event
		}
	}
	for _, event := range events {
		if event.HasRealtimeData || event.HasFinalizedData {
			return event
		}
	}
	if len(events) > 0 {
		return events[0]
	}
	return ranking.ForecastEvent{}
}

func rankingSubtitleForParser(region string, board ranking.Board) string {
	parts := []string{config.RegionLabel(region)}
	if board.EventID > 0 {
		parts = append(parts, fmt.Sprintf("Event #%d", board.EventID))
	}
	if board.BoardType == "worldlink" && board.TargetID > 0 {
		parts = append(parts, fmt.Sprintf("WL 角色 %d", board.TargetID))
	}
	if board.UpdatedAt > 0 {
		parts = append(parts, "更新 "+formatMillis(normalizeMillisForParser(board.UpdatedAt)))
	}
	return strings.Join(parts, " · ")
}

func normalizeMillisForParser(value int64) int64 {
	if value > 0 && value < 1_000_000_000_000 {
		return value * 1000
	}
	return value
}

func firstNonEmptyForParser(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func absIntForParser(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func searchAndBuild(def Definition, store *masterdata.Store, resolver *assets.Resolver, argument string) ([]EntityResult, *EntityResult) {
	argument = strings.TrimSpace(argument)
	switch def.SearchType {
	case SearchTypeCard:
		result := cardquery.Resolve(store, argument)
		cards := result.Cards
		rows := make([]EntityResult, 0, result.Total)
		allCards := cards
		if result.Mode == cardquery.ModeList && result.Total > len(cards) {
			all := cardquery.ResolveAll(store, argument)
			allCards = all.Cards
		}
		for _, card := range allCards {
			rows = append(rows, EntityResult{ID: card.ID, Title: card.Prefix, Subtitle: fmt.Sprintf("角色 #%d · %s", card.CharacterID, card.CardRarityType), Type: "card"})
		}
		if len(cards) == 0 {
			return rows, nil
		}
		if result.Mode == cardquery.ModeList {
			payload := renderer.BuildCardListPayloadWithAssets("卡牌查询", commandListSubtitle(argument), cards, store, resolver, result.Page, result.TotalPages, result.Total)
			selected := EntityResult{ID: rows[0].ID, Title: "卡牌列表", Subtitle: fmt.Sprintf("共 %d 张，展示前 %d 张", result.Total, len(cards)), Type: "card_list", Payload: payload}
			return rows, &selected
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
		if argument == "" || strings.EqualFold(argument, "当前") {
			argument = "当前"
		}
		events := store.SearchEvents(argument)
		if argument == "当前" {
			events = currentOrNextEventsForParser(store.AllEvents())
		}
		rows := make([]EntityResult, 0, len(events))
		for _, event := range events {
			rows = append(rows, EntityResult{ID: event.ID, Title: event.Name, Subtitle: fmt.Sprintf("%s · %s", event.EventType, event.Unit), Type: "event"})
		}
		if len(events) == 0 {
			return rows, nil
		}
		if len(events) > 1 {
			shown := events
			if len(shown) > 12 {
				shown = shown[:12]
			}
			payload := renderer.BuildEventListPayloadWithAssets("活动查询", commandListSubtitle(argument), shown, store, resolver, 1, (len(events)+11)/12, len(events))
			selected := EntityResult{ID: rows[0].ID, Title: "活动列表", Subtitle: fmt.Sprintf("共 %d 个，展示前 %d 个", len(events), len(shown)), Type: "event_list", Payload: payload}
			return rows, &selected
		}
		payload := renderer.BuildEventInfoPayloadWithAssets(store, events[0], resolver)
		selected := rows[0]
		selected.Payload = payload
		return rows, &selected
	case SearchTypeGacha:
		gachas := store.SearchGachas(argument)
		if argument == "" || strings.EqualFold(argument, "当前") {
			gachas = currentGachasForParser(store.AllGachas())
		}
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
	case SearchTypeVirtualLive:
		lives := searchVirtualLivesForParser(store, argument)
		rows := make([]EntityResult, 0, len(lives))
		for _, live := range lives {
			rows = append(rows, EntityResult{ID: live.ID, Title: live.Name, Subtitle: virtualLiveSubtitleForParser(live), Type: "virtual_live"})
		}
		if len(lives) == 0 {
			return rows, nil
		}
		payload := renderer.BuildVirtualLiveListPayloadWithAssets("虚拟 Live", "近期演唱会", lives, store, resolver, 1, 1, len(lives))
		selected := rows[0]
		selected.Payload = payload
		return rows, &selected
	default:
		return []EntityResult{}, nil
	}
}

func commandListSubtitle(argument string) string {
	argument = strings.TrimSpace(argument)
	if argument == "" {
		return "列表查询"
	}
	return "关键词：" + argument
}

func currentOrNextEventsForParser(events []masterdata.EventInfo) []masterdata.EventInfo {
	now := time.Now().UnixMilli()
	for _, event := range events {
		if event.StartAt <= now && now <= event.ClosedAt {
			return []masterdata.EventInfo{event}
		}
	}
	for _, event := range events {
		if event.StartAt > now {
			return []masterdata.EventInfo{event}
		}
	}
	return nil
}

func currentGachasForParser(gachas []masterdata.GachaInfo) []masterdata.GachaInfo {
	now := time.Now().UnixMilli()
	out := []masterdata.GachaInfo{}
	for _, gacha := range gachas {
		if gacha.StartAt <= now && (gacha.EndAt <= 0 || now <= gacha.EndAt) {
			out = append(out, gacha)
		}
	}
	if len(out) > 0 {
		return out
	}
	for _, gacha := range gachas {
		if gacha.StartAt <= now {
			out = append(out, gacha)
		}
	}
	if len(out) > 6 {
		return out[len(out)-6:]
	}
	return out
}

func searchVirtualLivesForParser(store *masterdata.Store, argument string) []masterdata.VirtualLive {
	lives := store.AllVirtualLives()
	now := time.Now().UnixMilli()
	argument = strings.TrimSpace(argument)
	out := []masterdata.VirtualLive{}
	for _, live := range lives {
		start, end := virtualLiveBoundsForParser(live)
		if argument == "" {
			if end > now && start-now < int64(7*24*time.Hour/time.Millisecond) {
				out = append(out, live)
			}
			continue
		}
		if fmt.Sprintf("%d", live.ID) == argument || strings.Contains(strings.ToLower(live.Name), strings.ToLower(argument)) || strings.Contains(strings.ToLower(live.AssetbundleName), strings.ToLower(argument)) {
			out = append(out, live)
		}
	}
	return out
}

func virtualLiveBoundsForParser(live masterdata.VirtualLive) (int64, int64) {
	start, end := live.StartAt, live.EndAt
	for i, schedule := range live.VirtualLiveSchedules {
		if i == 0 || schedule.StartAt < start || start == 0 {
			start = schedule.StartAt
		}
		if schedule.EndAt > end {
			end = schedule.EndAt
		}
	}
	return start, end
}

func virtualLiveSubtitleForParser(live masterdata.VirtualLive) string {
	start, end := virtualLiveBoundsForParser(live)
	return fmt.Sprintf("%s - %s", formatMillis(start), formatMillis(end))
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
