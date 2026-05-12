package crmission

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"moebot-next/internal/plugins/moesekai/assets"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/plugins/moesekai/renderpayloads"
	"moebot-next/internal/plugins/moesekai/suite"
)

const DefaultAllPageSize = 80

// Profile is the Suite payload shape needed by the CR mission command.
type Profile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata `json:"userGamedata"`
	Characters   []struct {
		CharacterID   int `json:"characterId"`
		CharacterRank int `json:"characterRank"`
	} `json:"userCharacters"`
	Missions []MissionV2       `json:"userCharacterMissionV2s"`
	Statuses []MissionV2Status `json:"userCharacterMissionV2Statuses"`
}

type MissionV2 struct {
	CharacterID          int    `json:"characterId"`
	CharacterMissionType string `json:"characterMissionType"`
	Progress             int    `json:"progress"`
}

type MissionV2Status struct {
	CharacterID      int    `json:"characterId"`
	ParameterGroupID int    `json:"parameterGroupId"`
	Seq              int    `json:"seq"`
	MissionStatus    string `json:"missionStatus"`
}

type Options struct {
	CharacterID int
	ShowAll     bool
	MissionType string
	Page        int
	PageSize    int
}

type Row struct {
	MissionType string  `json:"missionType"`
	Title       string  `json:"title"`
	Current     int     `json:"current"`
	Upper       int     `json:"upper,omitempty"`
	Level       int     `json:"level"`
	LevelMax    int     `json:"levelMax"`
	NextNeed    int     `json:"nextNeed,omitempty"`
	NextExp     int     `json:"nextExp,omitempty"`
	Progress    float64 `json:"progress"`
	IsEX        bool    `json:"isEx,omitempty"`
}

type AllRow struct {
	Seq            int  `json:"seq"`
	Requirement    int  `json:"requirement"`
	AccRequirement int  `json:"accRequirement"`
	Exp            int  `json:"exp"`
	AccExp         int  `json:"accExp"`
	Reached        bool `json:"reached"`
}

type Payload struct {
	Title        string                             `json:"title"`
	Subtitle     string                             `json:"subtitle,omitempty"`
	Profile      renderpayloads.SuiteProfilePayload `json:"profile"`
	CharacterID  int                                `json:"characterId"`
	Character    string                             `json:"character"`
	Rows         []Row                              `json:"rows,omitempty"`
	AllRows      []AllRow                           `json:"allRows,omitempty"`
	MissionType  string                             `json:"missionType,omitempty"`
	Mode         string                             `json:"mode"`
	AssetSource  string                             `json:"assetSource,omitempty"`
	Current      int                                `json:"current,omitempty"`
	AllRowsTotal int                                `json:"allRowsTotal,omitempty"`
	ShownFrom    int                                `json:"shownFrom,omitempty"`
	ShownTo      int                                `json:"shownTo,omitempty"`
	Page         int                                `json:"page,omitempty"`
	PageSize     int                                `json:"pageSize,omitempty"`
	TotalPages   int                                `json:"totalPages,omitempty"`
	Notice       string                             `json:"notice,omitempty"`
}

func Fields() []string {
	return suite.Fields(suite.FieldUserCharacterMissionV2s, suite.FieldUserCharacterMissionV2Statuses, suite.FieldUserCharacters)
}

func ParseArgs(raw string) (Options, error) {
	fields := strings.Fields(strings.TrimSpace(raw))
	if len(fields) == 0 {
		return Options{}, errors.New("使用方式: /cr任务 角色名 或 /cr任务 角色名 all 任务名")
	}
	cid := CharacterIDByAlias(fields[0])
	if cid <= 0 {
		return Options{}, fmt.Errorf("角色名无效: %s", fields[0])
	}
	opts := Options{CharacterID: cid, PageSize: DefaultAllPageSize}
	if len(fields) == 1 {
		return opts, nil
	}
	if !IsAllKeyword(fields[1]) {
		return Options{}, fmt.Errorf("参数无法解析: %s", strings.Join(fields[1:], " "))
	}
	missionTokens := make([]string, 0, len(fields)-2)
	for _, token := range fields[2:] {
		if page, ok := parsePageToken(token); ok {
			opts.Page = page
			continue
		}
		missionTokens = append(missionTokens, token)
	}
	if len(missionTokens) == 0 {
		return Options{}, errors.New("请在 all 后输入任务名")
	}
	missionType := TypeByAlias(strings.Join(missionTokens, ""))
	if missionType == "" {
		return Options{}, fmt.Errorf("未识别到角色等级任务名: %s", strings.Join(missionTokens, " "))
	}
	opts.ShowAll = true
	opts.MissionType = missionType
	return opts, nil
}

func IsAllKeyword(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "all", "全部", "全量", "总表", "表格":
		return true
	default:
		return false
	}
}

func parsePageToken(token string) (int, bool) {
	key := strings.ToLower(strings.TrimSpace(token))
	key = strings.Trim(key, "，,。")
	if key == "" {
		return 0, false
	}
	key = strings.TrimPrefix(key, "page")
	key = strings.TrimPrefix(key, "p")
	key = strings.TrimPrefix(key, "第")
	key = strings.TrimSuffix(key, "页")
	if key == strings.ToLower(strings.TrimSpace(token)) {
		return 0, false
	}
	page, err := strconv.Atoi(key)
	if err != nil || page <= 0 {
		return 0, false
	}
	return page, true
}

func BuildPayload(region string, profile Profile, store *masterdata.Store, resolver interface{ RendererAssetSource() string }, options Options) (Payload, string, error) {
	payload := Payload{
		Title:       fmt.Sprintf("%s CR任务", CharacterDisplayName(options.CharacterID)),
		Profile:     renderpayloads.BuildSuiteProfilePayload(region, "cr任务", profile.BaseProfile, profile.UserGamedata),
		CharacterID: options.CharacterID,
		Character:   CharacterDisplayName(options.CharacterID),
		Mode:        "overview",
	}
	if resolver != nil {
		payload.AssetSource = resolver.RendererAssetSource()
	}
	if options.ShowAll {
		rows, current, err := BuildAllRows(store, profile, options.CharacterID, options.MissionType)
		if err != nil {
			return payload, "", err
		}
		pageSize := options.PageSize
		if pageSize <= 0 {
			pageSize = DefaultAllPageSize
		}
		page := options.Page
		if page <= 0 {
			page = pageNearCurrent(rows, pageSize)
		}
		shown, page, totalPages, from, to := paginateRows(rows, page, pageSize)
		payload.Mode = "all"
		payload.MissionType = options.MissionType
		payload.Title = fmt.Sprintf("%s %s 档位表", CharacterDisplayName(options.CharacterID), Title(options.MissionType))
		payload.Current = current
		payload.AllRows = shown
		payload.AllRowsTotal = len(rows)
		payload.ShownFrom = from
		payload.ShownTo = to
		payload.Page = page
		payload.PageSize = pageSize
		payload.TotalPages = totalPages
		payload.Subtitle = fmt.Sprintf("当前进度 %d · 显示 %d-%d / %d · 第 %d/%d 页", current, from, to, len(rows), page, totalPages)
		if len(rows) > len(shown) {
			payload.Notice = fmt.Sprintf("档位过多，已分页渲染；追加 p%d 查看下一页。", nextPage(page, totalPages))
		}
		return payload, FormatAllText(region, profile, options.CharacterID, options.MissionType, rows, payload), nil
	}
	rows := BuildOverviewRows(store, profile, options.CharacterID)
	payload.Rows = rows
	payload.Subtitle = fmt.Sprintf("共 %d 项任务", len(rows))
	return payload, FormatOverviewText(region, profile, options.CharacterID, rows), nil
}

func BuildOverviewRows(store *masterdata.Store, profile Profile, cid int) []Row {
	rows := make([]Row, 0, len(Order))
	exLevels := exLevels(profile.Statuses)
	for _, missionType := range Order {
		if missionType == "play_live_ex" {
			row := BuildRow(store, profile, cid, missionType, exLevels[cid])
			if row.Current > 0 || row.Level > 0 {
				rows = append(rows, row)
			}
			continue
		}
		row := BuildRow(store, profile, cid, missionType, 0)
		if row.Current > 0 || row.Level > 0 || missionType == "play_live" {
			rows = append(rows, row)
		}
	}
	return rows
}

func BuildRow(store *masterdata.Store, profile Profile, cid int, missionType string, exSeq int) Row {
	progress := MissionProgress(profile, cid, missionType)
	pgid := ParameterGroupID(missionType)
	groups := sortedMissionGroups(store, pgid)
	current := progress
	if missionType == "play_live_ex" {
		cleared := exClearedTotal(groups, exSeq)
		if progress < cleared {
			current = cleared + progress
		} else if progress == 0 {
			current = cleared
		}
	}
	level, nextNeed, nextExp := levelAndNext(groups, current)
	upper := maxRequirement(groups)
	if missionType == "play_live_ex" {
		level, nextNeed, nextExp = exLevelAndNext(groups, current)
		upper = exLimit(groups, 30)
	}
	progressRate := 0.0
	if upper > 0 {
		progressRate = float64(minInt(current, upper)) / float64(upper)
	}
	return Row{MissionType: missionType, Title: Title(missionType), Current: current, Upper: upper, Level: level, LevelMax: len(groups), NextNeed: nextNeed, NextExp: nextExp, Progress: progressRate, IsEX: missionType == "play_live_ex"}
}

func BuildAllRows(store *masterdata.Store, profile Profile, cid int, missionType string) ([]AllRow, int, error) {
	pgid := ParameterGroupID(missionType)
	groups := sortedMissionGroups(store, pgid)
	if len(groups) == 0 {
		return nil, 0, fmt.Errorf("找不到任务档位数据: %s", missionType)
	}
	current := MissionProgress(profile, cid, missionType)
	if missionType == "play_live_ex" {
		current = BuildRow(store, profile, cid, missionType, exLevels(profile.Statuses)[cid]).Current
	}
	rows := make([]AllRow, 0, len(groups))
	accExp := 0
	accReq := 0
	for _, group := range groups {
		accExp += group.Exp
		requirement := group.Requirement
		accRequirement := group.Requirement
		if missionType == "play_live_ex" {
			accReq += group.Requirement
			requirement = group.Requirement
			accRequirement = accReq
		}
		rows = append(rows, AllRow{Seq: group.Seq, Requirement: requirement, AccRequirement: accRequirement, Exp: group.Exp, AccExp: accExp, Reached: current >= accRequirement})
	}
	return rows, current, nil
}

func MissionProgress(profile Profile, cid int, missionType string) int {
	progress := 0
	for _, mission := range profile.Missions {
		if mission.CharacterID == cid && mission.CharacterMissionType == missionType {
			progress = maxInt(progress, mission.Progress)
		}
	}
	return progress
}

func pageNearCurrent(rows []AllRow, pageSize int) int {
	if pageSize <= 0 {
		pageSize = DefaultAllPageSize
	}
	if len(rows) == 0 {
		return 1
	}
	idx := len(rows) - 1
	for i, row := range rows {
		if !row.Reached {
			idx = i
			break
		}
	}
	return idx/pageSize + 1
}

func paginateRows(rows []AllRow, page int, pageSize int) ([]AllRow, int, int, int, int) {
	if pageSize <= 0 {
		pageSize = DefaultAllPageSize
	}
	totalPages := maxInt(1, (len(rows)+pageSize-1)/pageSize)
	if page <= 0 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}
	start := (page - 1) * pageSize
	if start > len(rows) {
		start = len(rows)
	}
	end := minInt(len(rows), start+pageSize)
	from, to := 0, 0
	if len(rows) > 0 && end > start {
		from = start + 1
		to = end
	}
	return rows[start:end], page, totalPages, from, to
}

func nextPage(page int, totalPages int) int {
	if totalPages <= 0 {
		return 1
	}
	if page >= totalPages {
		return totalPages
	}
	return page + 1
}

func FormatOverviewText(region string, profile Profile, cid int, rows []Row) string {
	lines := []string{fmt.Sprintf("%s %s CR任务", strings.ToUpper(region), CharacterDisplayName(cid)), fmt.Sprintf("玩家: %s", firstNonEmpty(profile.UserGamedata.Name, "未知玩家"))}
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("%s: %d | 档位 %d/%d | 下一档 %d", row.Title, row.Current, row.Level, row.LevelMax, row.NextNeed))
	}
	return strings.Join(lines, "\n")
}

func FormatAllText(region string, profile Profile, cid int, missionType string, rows []AllRow, payload Payload) string {
	lines := []string{fmt.Sprintf("%s %s %s 档位表", strings.ToUpper(region), CharacterDisplayName(cid), Title(missionType)), fmt.Sprintf("玩家: %s", firstNonEmpty(profile.UserGamedata.Name, "未知玩家"))}
	if payload.AllRowsTotal > 0 {
		lines = append(lines, fmt.Sprintf("当前进度: %d", payload.Current))
		lines = append(lines, fmt.Sprintf("显示: %d-%d / %d（第 %d/%d 页）", payload.ShownFrom, payload.ShownTo, payload.AllRowsTotal, payload.Page, payload.TotalPages))
	}
	shown := payload.AllRows
	if len(shown) == 0 && len(rows) > 0 {
		shown = rows
	}
	for _, row := range shown {
		status := "未达成"
		if row.Reached {
			status = "已达成"
		}
		lines = append(lines, fmt.Sprintf("%d. %d EXP+%d %s", row.Seq, row.AccRequirement, row.Exp, status))
	}
	if payload.Notice != "" {
		lines = append(lines, payload.Notice)
	}
	return strings.Join(lines, "\n")
}

func CharacterIDByAlias(raw string) int {
	query := assets.NormalizeAlias(raw)
	if query == "" {
		return 0
	}
	for _, entry := range assets.CharacterAliasEntries() {
		if entry.Normalized == query {
			return entry.CharacterID
		}
	}
	return 0
}

func CharacterDisplayName(id int) string {
	if ch := assets.GetCharacterByID(id); ch != nil {
		if ch.NameCN != "" {
			return ch.NameCN
		}
		if ch.NameJP != "" {
			return ch.NameJP
		}
		if ch.NameEN != "" {
			return ch.NameEN
		}
	}
	if id > 0 {
		return fmt.Sprintf("角色 %d", id)
	}
	return "未知角色"
}

func Title(missionType string) string {
	if title := Titles[missionType]; title != "" {
		return title
	}
	return missionType
}

func TypeByAlias(raw string) string {
	key := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(raw), "（", "("), "）", ")"))
	return Aliases[key]
}

func ParameterGroupID(missionType string) int {
	switch missionType {
	case "play_live":
		return 1
	case "play_live_ex":
		return 101
	default:
		return parameterGroupIDs[missionType]
	}
}

func levelAndNext(groups []masterdata.CharacterMissionV2ParameterGroup, current int) (level int, nextNeed int, nextExp int) {
	for _, group := range groups {
		if current >= group.Requirement {
			level = group.Seq
			continue
		}
		if nextNeed == 0 {
			nextNeed = group.Requirement
			nextExp = group.Exp
		}
	}
	return level, nextNeed, nextExp
}

func exLevelAndNext(groups []masterdata.CharacterMissionV2ParameterGroup, current int) (level int, nextNeed int, nextExp int) {
	acc := 0
	for round := 1; round <= 30; round++ {
		req := requirementForRound(groups, round)
		if req <= 0 {
			break
		}
		acc += req
		exp := expForRound(groups, round)
		if current >= acc {
			level = round
			continue
		}
		return level, acc, exp
	}
	return level, 0, 0
}

func expForRound(groups []masterdata.CharacterMissionV2ParameterGroup, round int) int {
	exp := 0
	for _, group := range groups {
		if group.Seq > round {
			break
		}
		exp = group.Exp
	}
	return exp
}

func sortedMissionGroups(store *masterdata.Store, pgid int) []masterdata.CharacterMissionV2ParameterGroup {
	if store == nil || pgid <= 0 {
		return nil
	}
	groups := store.GetCharacterMissionV2ParameterGroups(pgid)
	sort.SliceStable(groups, func(i, j int) bool { return groups[i].Seq < groups[j].Seq })
	return groups
}

func maxRequirement(groups []masterdata.CharacterMissionV2ParameterGroup) int {
	maxReq := 0
	for _, group := range groups {
		maxReq = maxInt(maxReq, group.Requirement)
	}
	return maxReq
}

func exClearedTotal(groups []masterdata.CharacterMissionV2ParameterGroup, seq int) int {
	total := 0
	for round := 1; round <= seq; round++ {
		total += requirementForRound(groups, round)
	}
	return total
}

func exLimit(groups []masterdata.CharacterMissionV2ParameterGroup, rounds int) int {
	total := 0
	for round := 1; round <= rounds; round++ {
		total += requirementForRound(groups, round)
	}
	return total
}

func requirementForRound(groups []masterdata.CharacterMissionV2ParameterGroup, round int) int {
	req := 0
	for _, group := range groups {
		if group.Seq > round {
			break
		}
		req = group.Requirement
	}
	return req
}

func exLevels(statuses []MissionV2Status) map[int]int {
	out := map[int]int{}
	for _, status := range statuses {
		if status.CharacterID <= 0 || status.ParameterGroupID != 101 {
			continue
		}
		out[status.CharacterID] = maxInt(out[status.CharacterID], status.Seq)
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

var Order = []string{
	"play_live", "play_live_ex", "waiting_room", "waiting_room_ex", "collect_costume_3d", "collect_stamp", "read_area_talk", "read_card_episode_first", "read_card_episode_second", "collect_another_vocal", "area_item_level_up_character", "area_item_level_up_unit", "area_item_level_up_reality_world", "collect_member", "skill_level_up_rare", "skill_level_up_standard", "master_rank_up_rare", "master_rank_up_standard", "collect_character_archive_voice", "collect_mysekai_fixture", "collect_mysekai_canvas", "read_mysekai_fixture_unique_character_talk",
}

var Titles = map[string]string{
	"play_live": "队长次数", "play_live_ex": "队长次数(EX)", "waiting_room": "休息室次数", "waiting_room_ex": "休息室次数(EX)", "collect_costume_3d": "服装", "collect_stamp": "表情", "read_area_talk": "区域对话", "read_card_episode_first": "卡面剧情前篇", "read_card_episode_second": "卡面剧情后篇", "collect_another_vocal": "Another Vocal", "area_item_level_up_character": "单人家具升级次数", "area_item_level_up_unit": "团家具升级次数", "area_item_level_up_reality_world": "属性道具升级次数", "collect_member": "卡面", "skill_level_up_rare": "技能等级升级次数（★4&生日卡）", "skill_level_up_standard": "技能等级升级次数（★1~★3）", "master_rank_up_rare": "专精等级升级次数（★4&生日卡）", "master_rank_up_standard": "专精等级升级次数（★1~★3）", "collect_character_archive_voice": "台词", "collect_mysekai_fixture": "MySekai家具数量", "collect_mysekai_canvas": "MySekai画布数量", "read_mysekai_fixture_unique_character_talk": "MySekai对话",
}

var Aliases = map[string]string{
	"队长次数": "play_live", "队长": "play_live", "角色次数": "play_live", "play_live": "play_live",
	"队长次数ex": "play_live_ex", "队长次数(ex)": "play_live_ex", "队长ex": "play_live_ex", "队长(ex)": "play_live_ex", "play_live_ex": "play_live_ex",
	"休息室次数": "waiting_room", "休息室": "waiting_room", "控制室": "waiting_room", "waiting_room": "waiting_room",
	"服装": "collect_costume_3d", "衣装": "collect_costume_3d",
	"表情": "collect_stamp", "贴纸": "collect_stamp",
	"区域对话":   "read_area_talk",
	"卡面剧情前篇": "read_card_episode_first", "前篇": "read_card_episode_first", "前编": "read_card_episode_first",
	"卡面剧情后篇": "read_card_episode_second", "后篇": "read_card_episode_second", "后编": "read_card_episode_second",
	"anvo": "collect_another_vocal", "another vocal": "collect_another_vocal", "another_vocal": "collect_another_vocal",
	"单人家具": "area_item_level_up_character", "单人道具": "area_item_level_up_character",
	"团家具": "area_item_level_up_unit",
	"树花":  "area_item_level_up_reality_world", "属性家具": "area_item_level_up_reality_world", "属性道具": "area_item_level_up_reality_world", "植物": "area_item_level_up_reality_world",
	"卡面": "collect_member", "图鉴": "collect_member", "成员": "collect_member",
	"4星技能": "skill_level_up_rare", "四星技能": "skill_level_up_rare", "四星slv": "skill_level_up_rare", "4星slv": "skill_level_up_rare",
	"低星技能": "skill_level_up_standard", "低星slv": "skill_level_up_standard",
	"4星专精": "master_rank_up_rare", "四星专精": "master_rank_up_rare", "四星突破": "master_rank_up_rare", "4星突破": "master_rank_up_rare", "4星mr": "master_rank_up_rare", "四星mr": "master_rank_up_rare",
	"低星专精": "master_rank_up_standard", "低星突破": "master_rank_up_standard", "低星mr": "master_rank_up_standard",
	"台词": "collect_character_archive_voice", "语音": "collect_character_archive_voice",
	"ms家具": "collect_mysekai_fixture", "烤森家具": "collect_mysekai_fixture", "mysekai家具数量": "collect_mysekai_fixture",
	"ms画布": "collect_mysekai_canvas", "烤森画布": "collect_mysekai_canvas", "mysekai画布数量": "collect_mysekai_canvas",
	"ms对话": "read_mysekai_fixture_unique_character_talk", "烤森对话": "read_mysekai_fixture_unique_character_talk", "mysekai对话": "read_mysekai_fixture_unique_character_talk",
}

var parameterGroupIDs = map[string]int{
	"waiting_room": 2, "waiting_room_ex": 102, "collect_costume_3d": 3, "collect_stamp": 4, "read_area_talk": 5, "read_card_episode_first": 6, "read_card_episode_second": 7, "collect_another_vocal": 8, "area_item_level_up_character": 9, "area_item_level_up_unit": 10, "area_item_level_up_reality_world": 11, "collect_member": 12, "skill_level_up_rare": 13, "skill_level_up_standard": 14, "master_rank_up_rare": 15, "master_rank_up_standard": 16, "collect_character_archive_voice": 17, "collect_mysekai_fixture": 18, "collect_mysekai_canvas": 19, "read_mysekai_fixture_unique_character_talk": 20,
}
