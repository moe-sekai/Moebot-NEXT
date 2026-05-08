package commands

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/renderer"
	"moebot-next/internal/plugins/moesekai/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"moebot-next/internal/plugins/moesekai/renderpayloads"
)

type characterRankMissionProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata `json:"userGamedata"`
	Characters   []struct {
		CharacterID   int `json:"characterId"`
		CharacterRank int `json:"characterRank"`
	} `json:"userCharacters"`
	Missions []characterMissionV2       `json:"userCharacterMissionV2s"`
	Statuses []characterMissionV2Status `json:"userCharacterMissionV2Statuses"`
}

type characterRankMissionOptions struct {
	CharacterID int
	ShowAll     bool
	MissionType string
}

type characterRankMissionRow struct {
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

type characterRankMissionAllRow struct {
	Seq            int  `json:"seq"`
	Requirement    int  `json:"requirement"`
	AccRequirement int  `json:"accRequirement"`
	Exp            int  `json:"exp"`
	AccExp         int  `json:"accExp"`
	Reached        bool `json:"reached"`
}

type characterRankMissionPayload struct {
	Title       string                       `json:"title"`
	Subtitle    string                       `json:"subtitle,omitempty"`
	Profile     renderpayloads.SuiteProfilePayload `json:"profile"`
	CharacterID int                          `json:"characterId"`
	Character   string                       `json:"character"`
	Rows        []characterRankMissionRow    `json:"rows,omitempty"`
	AllRows     []characterRankMissionAllRow `json:"allRows,omitempty"`
	MissionType string                       `json:"missionType,omitempty"`
	Mode        string                       `json:"mode"`
	AssetSource string                       `json:"assetSource,omitempty"`
}

func characterRankMissionFields() []string {
	return suite.Fields(suite.FieldUserCharacterMissionV2s, suite.FieldUserCharacterMissionV2Statuses, suite.FieldUserCharacters)
}

func RegisterCharacterRankMission(deps *Deps) {
	for _, cmd := range parserCommands(deps, "CR任务") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		Engine.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, inferredUser, ok := requireRuntimeWithStore(deps, ctx, forcedRegion)
			if !ok {
				return
			}
			user, ok := requireBoundUser(deps, ctx, runtime, forcedRegion, inferredUser)
			if !ok {
				return
			}
			if !requireSuite(ctx, runtime, "CR任务") {
				return
			}
			if _, ok := requireSuiteVisible(deps, ctx, runtime); !ok {
				return
			}
			options, err := parseCharacterRankMissionArgs(commandArgs(ctx))
			if err != nil {
				ctx.SendChain(message.Text(err.Error()))
				return
			}
			var profile characterRankMissionProfile
			if !fetchSuiteUserData(ctx, runtime, user.GameID, "CR任务", characterRankMissionFields(), &profile) {
				return
			}
			payload, fallback, err := buildCharacterRankMissionPayload(runtime.Region, profile, runtime.Store, runtime.Assets, options)
			if err != nil {
				ctx.SendChain(message.Text(err.Error()))
				return
			}
			sendCharacterRankMissionOrText(ctx, deps, payload, fallback)
			bot.RecordCommandRegion(deps.DB, "CR任务", runtime.Region, ctx, start)
		})
	}
}

func parseCharacterRankMissionArgs(raw string) (characterRankMissionOptions, error) {
	fields := strings.Fields(strings.TrimSpace(raw))
	if len(fields) == 0 {
		return characterRankMissionOptions{}, errors.New("使用方式: /cr任务 角色名 或 /cr任务 角色名 all 任务名")
	}
	cid := characterIDByAlias(fields[0])
	if cid <= 0 {
		return characterRankMissionOptions{}, fmt.Errorf("角色名无效: %s", fields[0])
	}
	opts := characterRankMissionOptions{CharacterID: cid}
	if len(fields) == 1 {
		return opts, nil
	}
	if !isCharacterRankAllKeyword(fields[1]) {
		return characterRankMissionOptions{}, fmt.Errorf("参数无法解析: %s", strings.Join(fields[1:], " "))
	}
	if len(fields) < 3 {
		return characterRankMissionOptions{}, errors.New("请在 all 后输入任务名")
	}
	missionType := characterRankMissionTypeByAlias(strings.Join(fields[2:], ""))
	if missionType == "" {
		return characterRankMissionOptions{}, fmt.Errorf("未识别到角色等级任务名: %s", strings.Join(fields[2:], " "))
	}
	opts.ShowAll = true
	opts.MissionType = missionType
	return opts, nil
}

func isCharacterRankAllKeyword(value string) bool {
	switch strings.ToLower(value) {
	case "all", "全部", "全量", "总表", "表格":
		return true
	default:
		return false
	}
}

func buildCharacterRankMissionPayload(region string, profile characterRankMissionProfile, store *masterdata.Store, resolver interface{ RendererAssetSource() string }, options characterRankMissionOptions) (characterRankMissionPayload, string, error) {
	payload := characterRankMissionPayload{
		Title:       fmt.Sprintf("%s CR任务", characterDisplayName(options.CharacterID)),
		Profile:     renderpayloads.BuildSuiteProfilePayload(region, "cr任务", profile.BaseProfile, profile.UserGamedata),
		CharacterID: options.CharacterID,
		Character:   characterDisplayName(options.CharacterID),
		Mode:        "overview",
	}
	if resolver != nil {
		payload.AssetSource = resolver.RendererAssetSource()
	}
	if options.ShowAll {
		rows, err := buildCharacterRankMissionAllRows(store, profile, options.CharacterID, options.MissionType)
		if err != nil {
			return payload, "", err
		}
		payload.Mode = "all"
		payload.MissionType = options.MissionType
		payload.Title = fmt.Sprintf("%s %s 档位表", characterDisplayName(options.CharacterID), characterRankMissionTitle(options.MissionType))
		payload.Subtitle = fmt.Sprintf("当前进度 %d", missionProgress(profile, options.CharacterID, options.MissionType))
		payload.AllRows = rows
		return payload, formatCharacterRankMissionAllText(region, profile, options.CharacterID, options.MissionType, rows), nil
	}
	rows := buildCharacterRankMissionOverviewRows(store, profile, options.CharacterID)
	payload.Rows = rows
	payload.Subtitle = fmt.Sprintf("共 %d 项任务", len(rows))
	return payload, formatCharacterRankMissionOverviewText(region, profile, options.CharacterID, rows), nil
}

func buildCharacterRankMissionOverviewRows(store *masterdata.Store, profile characterRankMissionProfile, cid int) []characterRankMissionRow {
	rows := make([]characterRankMissionRow, 0, len(characterRankMissionOrder))
	exLevels := leaderExLevels(profile.Statuses)
	for _, missionType := range characterRankMissionOrder {
		if missionType == "play_live_ex" {
			row := buildCharacterRankMissionRow(store, profile, cid, missionType, exLevels[cid])
			if row.Current > 0 || row.Level > 0 {
				rows = append(rows, row)
			}
			continue
		}
		row := buildCharacterRankMissionRow(store, profile, cid, missionType, 0)
		if row.Current > 0 || row.Level > 0 || missionType == "play_live" {
			rows = append(rows, row)
		}
	}
	return rows
}

func buildCharacterRankMissionRow(store *masterdata.Store, profile characterRankMissionProfile, cid int, missionType string, exSeq int) characterRankMissionRow {
	progress := missionProgress(profile, cid, missionType)
	pgid := characterRankMissionParameterGroupID(missionType)
	groups := sortedMissionGroups(store, pgid)
	current := progress
	if missionType == "play_live_ex" {
		cleared := missionExClearedTotal(groups, exSeq)
		if progress < cleared {
			current = cleared + progress
		} else if progress == 0 {
			current = cleared
		}
	}
	level, nextNeed, nextExp := missionLevelAndNext(groups, current)
	upper := maxMissionRequirement(groups)
	if missionType == "play_live_ex" {
		level, nextNeed, nextExp = missionExLevelAndNext(groups, current)
		upper = missionExLimit(groups, 30)
	}
	progressRate := 0.0
	if upper > 0 {
		progressRate = float64(min(current, upper)) / float64(upper)
	}
	return characterRankMissionRow{MissionType: missionType, Title: characterRankMissionTitle(missionType), Current: current, Upper: upper, Level: level, LevelMax: len(groups), NextNeed: nextNeed, NextExp: nextExp, Progress: progressRate, IsEX: missionType == "play_live_ex"}
}

func buildCharacterRankMissionAllRows(store *masterdata.Store, profile characterRankMissionProfile, cid int, missionType string) ([]characterRankMissionAllRow, error) {
	pgid := characterRankMissionParameterGroupID(missionType)
	groups := sortedMissionGroups(store, pgid)
	if len(groups) == 0 {
		return nil, fmt.Errorf("找不到任务档位数据: %s", missionType)
	}
	current := missionProgress(profile, cid, missionType)
	if missionType == "play_live_ex" {
		current = buildCharacterRankMissionRow(store, profile, cid, missionType, leaderExLevels(profile.Statuses)[cid]).Current
	}
	rows := make([]characterRankMissionAllRow, 0, len(groups))
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
		rows = append(rows, characterRankMissionAllRow{Seq: group.Seq, Requirement: requirement, AccRequirement: accRequirement, Exp: group.Exp, AccExp: accExp, Reached: current >= accRequirement})
	}
	return rows, nil
}

func missionProgress(profile characterRankMissionProfile, cid int, missionType string) int {
	progress := 0
	for _, mission := range profile.Missions {
		if mission.CharacterID == cid && mission.CharacterMissionType == missionType {
			progress = max(progress, mission.Progress)
		}
	}
	return progress
}

func missionLevelAndNext(groups []masterdata.CharacterMissionV2ParameterGroup, current int) (level int, nextNeed int, nextExp int) {
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

func missionExLevelAndNext(groups []masterdata.CharacterMissionV2ParameterGroup, current int) (level int, nextNeed int, nextExp int) {
	acc := 0
	for round := 1; round <= 30; round++ {
		req := leaderRequirementForRound(groups, round)
		if req <= 0 {
			break
		}
		acc += req
		exp := missionExpForRound(groups, round)
		if current >= acc {
			level = round
			continue
		}
		return level, acc, exp
	}
	return level, 0, 0
}

func missionExpForRound(groups []masterdata.CharacterMissionV2ParameterGroup, round int) int {
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

func maxMissionRequirement(groups []masterdata.CharacterMissionV2ParameterGroup) int {
	maxReq := 0
	for _, group := range groups {
		maxReq = max(maxReq, group.Requirement)
	}
	return maxReq
}

func missionExClearedTotal(groups []masterdata.CharacterMissionV2ParameterGroup, seq int) int {
	total := 0
	for round := 1; round <= seq; round++ {
		total += leaderRequirementForRound(groups, round)
	}
	return total
}

func missionExLimit(groups []masterdata.CharacterMissionV2ParameterGroup, rounds int) int {
	total := 0
	for round := 1; round <= rounds; round++ {
		total += leaderRequirementForRound(groups, round)
	}
	return total
}

func sendCharacterRankMissionOrText(ctx *zero.Ctx, deps *Deps, payload characterRankMissionPayload, fallback string) {
	if deps.Renderer != nil && deps.Renderer.Health() {
		if png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "character_rank_mission", Data: payload}); err == nil {
			ctx.SendChain(message.ImageBytes(png))
			return
		}
	}
	ctx.SendChain(message.Text(fallback))
}

func formatCharacterRankMissionOverviewText(region string, profile characterRankMissionProfile, cid int, rows []characterRankMissionRow) string {
	lines := []string{fmt.Sprintf("%s %s CR任务", strings.ToUpper(region), characterDisplayName(cid)), fmt.Sprintf("玩家: %s", firstNonEmpty(profile.UserGamedata.Name, "未知玩家"))}
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("%s: %d | 档位 %d/%d | 下一档 %d", row.Title, row.Current, row.Level, row.LevelMax, row.NextNeed))
	}
	return strings.Join(lines, "\n")
}

func formatCharacterRankMissionAllText(region string, profile characterRankMissionProfile, cid int, missionType string, rows []characterRankMissionAllRow) string {
	lines := []string{fmt.Sprintf("%s %s %s 档位表", strings.ToUpper(region), characterDisplayName(cid), characterRankMissionTitle(missionType)), fmt.Sprintf("玩家: %s", firstNonEmpty(profile.UserGamedata.Name, "未知玩家"))}
	for _, row := range rows {
		status := "未达成"
		if row.Reached {
			status = "已达成"
		}
		lines = append(lines, fmt.Sprintf("%d. %d EXP+%d %s", row.Seq, row.AccRequirement, row.Exp, status))
	}
	return strings.Join(lines, "\n")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

var characterRankMissionOrder = []string{
	"play_live", "play_live_ex", "waiting_room", "waiting_room_ex", "collect_costume_3d", "collect_stamp", "read_area_talk", "read_card_episode_first", "read_card_episode_second", "collect_another_vocal", "area_item_level_up_character", "area_item_level_up_unit", "area_item_level_up_reality_world", "collect_member", "skill_level_up_rare", "skill_level_up_standard", "master_rank_up_rare", "master_rank_up_standard", "collect_character_archive_voice", "collect_mysekai_fixture", "collect_mysekai_canvas", "read_mysekai_fixture_unique_character_talk",
}

func characterRankMissionTitle(missionType string) string {
	if title := characterRankMissionTitles[missionType]; title != "" {
		return title
	}
	return missionType
}

var characterRankMissionTitles = map[string]string{
	"play_live": "队长次数", "play_live_ex": "队长次数(EX)", "waiting_room": "休息室次数", "waiting_room_ex": "休息室次数(EX)", "collect_costume_3d": "服装", "collect_stamp": "表情", "read_area_talk": "区域对话", "read_card_episode_first": "卡面剧情前篇", "read_card_episode_second": "卡面剧情后篇", "collect_another_vocal": "Another Vocal", "area_item_level_up_character": "单人家具升级次数", "area_item_level_up_unit": "团家具升级次数", "area_item_level_up_reality_world": "属性道具升级次数", "collect_member": "卡面", "skill_level_up_rare": "技能等级升级次数（★4&生日卡）", "skill_level_up_standard": "技能等级升级次数（★1~★3）", "master_rank_up_rare": "专精等级升级次数（★4&生日卡）", "master_rank_up_standard": "专精等级升级次数（★1~★3）", "collect_character_archive_voice": "台词", "collect_mysekai_fixture": "MySekai家具数量", "collect_mysekai_canvas": "MySekai画布数量", "read_mysekai_fixture_unique_character_talk": "MySekai对话",
}

func characterRankMissionTypeByAlias(raw string) string {
	key := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(raw), "（", "("), "）", ")"))
	aliases := map[string]string{
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
	return aliases[key]
}

func characterRankMissionParameterGroupID(missionType string) int {
	switch missionType {
	case "play_live":
		return 1
	case "play_live_ex":
		return 101
	default:
		// MasterData 中不同角色任务的 parameterGroupId 会随类型变化；当前先使用
		// 与任务类型稳定对应的轻量映射，不存在时由无档位展示兜底。
		return map[string]int{
			"waiting_room": 2, "waiting_room_ex": 102, "collect_costume_3d": 3, "collect_stamp": 4, "read_area_talk": 5, "read_card_episode_first": 6, "read_card_episode_second": 7, "collect_another_vocal": 8, "area_item_level_up_character": 9, "area_item_level_up_unit": 10, "area_item_level_up_reality_world": 11, "collect_member": 12, "skill_level_up_rare": 13, "skill_level_up_standard": 14, "master_rank_up_rare": 15, "master_rank_up_standard": 16, "collect_character_archive_voice": 17, "collect_mysekai_fixture": 18, "collect_mysekai_canvas": 19, "read_mysekai_fixture_unique_character_talk": 20,
		}[missionType]
	}
}
