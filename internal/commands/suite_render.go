package commands

import (
	"fmt"
	"sort"
	"strings"

	"moebot-next/internal/config"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/renderer"
	"moebot-next/internal/servers"
	"moebot-next/internal/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type suiteCommandProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata `json:"userGamedata"`
	UserDecks    []suite.UserDeck   `json:"userDecks"`
	UserCards    []suite.UserCard   `json:"userCards"`
}

func sendSuitePanelOrText(ctx *zero.Ctx, deps *Deps, payload renderer.SuitePanelPayload, fallback string) {
	if deps != nil && deps.Renderer != nil && deps.Renderer.Health() {
		png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "suite_panel", Data: payload})
		if err == nil {
			ctx.SendChain(message.ImageBytes(png))
			return
		}
	}
	ctx.SendChain(message.Text(fallback))
}

func sendSuiteCardBoxOrText(ctx *zero.Ctx, deps *Deps, payload renderer.SuiteCardBoxPayload, fallback string) {
	if deps != nil && deps.Renderer != nil && deps.Renderer.Health() {
		png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "suite_card_box", Data: payload, Precision: 1})
		if err == nil {
			ctx.SendChain(message.ImageBytes(png))
			return
		}
	}
	ctx.SendChain(message.Text(fallback))
}

func buildSuitePanel(runtime *servers.Runtime, title string, mode string, profile interface {
	commonSuiteProfile() suiteCommandProfile
}) renderer.SuitePanelPayload {
	common := profile.commonSuiteProfile()
	region := ""
	if runtime != nil {
		region = runtime.Region
	}
	payload := renderer.SuitePanelPayload{
		Title:       title,
		Profile:     renderer.BuildSuiteProfilePayload(region, mode, common.BaseProfile, common.UserGamedata),
		AssetSource: suiteAssetSource(runtime),
	}
	if runtime != nil {
		payload.DeckCards = renderer.BuildSuiteDeckCards(common.UserDecks, common.UserCards, common.UserGamedata.Deck, runtime.Store, runtime.Assets)
	}
	return payload
}

func (p suiteCommandProfile) commonSuiteProfile() suiteCommandProfile { return p }

func (p gachaHistoryProfile) commonSuiteProfile() suiteCommandProfile {
	return suiteCommandProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata}
}

func (p bondProfile) commonSuiteProfile() suiteCommandProfile {
	return suiteCommandProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func (p musicProgressProfile) commonSuiteProfile() suiteCommandProfile {
	return suiteCommandProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func (p materialProfile) commonSuiteProfile() suiteCommandProfile {
	return suiteCommandProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func (p challengeProfile) commonSuiteProfile() suiteCommandProfile {
	return suiteCommandProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func (p eventRecordProfile) commonSuiteProfile() suiteCommandProfile {
	return suiteCommandProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func (p leaderCountProfile) commonSuiteProfile() suiteCommandProfile {
	return suiteCommandProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func (p musicRewardProfile) commonSuiteProfile() suiteCommandProfile {
	return suiteCommandProfile{BaseProfile: p.BaseProfile, UserGamedata: p.UserGamedata, UserDecks: p.UserDecks, UserCards: p.UserCards}
}

func suiteAssetSource(runtime *servers.Runtime) string {
	if runtime == nil || runtime.Assets == nil {
		return ""
	}
	return runtime.Assets.RendererAssetSource()
}

func suitePanelTitle(runtime *servers.Runtime, name string) string {
	region := ""
	if runtime != nil {
		region = runtime.Region
	}
	return fmt.Sprintf("%s %s", strings.ToUpper(config.NormalizeRegion(region)), name)
}

func suitePanelSubtitle(profile suite.BaseProfile) string {
	return fmt.Sprintf("更新时间: %s · 数据来源: %s", suiteUpdateText(profile.UploadTime), suiteSourceText(profile))
}

func suiteBasicStats(profile suiteCommandProfile) []renderer.SuiteStatPayload {
	stats := make([]renderer.SuiteStatPayload, 0, 4)
	if profile.UserGamedata.Rank > 0 {
		stats = append(stats, renderer.SuiteStatPayload{Label: "Rank", Value: formatInt(profile.UserGamedata.Rank)})
	}
	if profile.UserGamedata.Coin > 0 {
		stats = append(stats, renderer.SuiteStatPayload{Label: "金币", Value: formatInt64(profile.UserGamedata.Coin)})
	}
	if len(profile.UserCards) > 0 {
		stats = append(stats, renderer.SuiteStatPayload{Label: "持有卡牌", Value: formatInt(len(profile.UserCards))})
	}
	return stats
}

func rowsFromGachaHistory(profile gachaHistoryProfile, store *masterdata.Store, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	records := make([]userGachaRecord, 0, len(profile.UserGachas))
	total := 0
	for _, record := range profile.UserGachas {
		if record.Count <= 0 {
			continue
		}
		records = append(records, record)
		total += record.Count
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].Count == records[j].Count {
			return records[i].GachaID > records[j].GachaID
		}
		return records[i].Count > records[j].Count
	})
	limit = clampLimit(limit, len(records))
	rows := make([]renderer.SuiteSectionRowPayload, 0, limit)
	for i := 0; i < limit; i++ {
		record := records[i]
		rows = append(rows, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: gachaHistoryName(store, record.GachaID), Value: fmt.Sprintf("%d抽", record.Count)})
	}
	return rows, []renderer.SuiteStatPayload{{Label: "总抽数", Value: formatInt(total)}, {Label: "卡池数", Value: formatInt(len(records))}}
}

func rowsFromBonds(profile bondProfile, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	bonds := make([]userBond, 0, len(profile.UserBonds))
	for _, bond := range profile.UserBonds {
		cid1, cid2 := bondCharacterIDs(bond)
		if cid1 <= 0 || cid2 <= 0 {
			continue
		}
		bonds = append(bonds, bond)
	}
	sort.SliceStable(bonds, func(i, j int) bool {
		if bonds[i].Rank == bonds[j].Rank {
			return bonds[i].Exp > bonds[j].Exp
		}
		return bonds[i].Rank > bonds[j].Rank
	})
	limit = clampLimit(limit, len(bonds))
	rows := make([]renderer.SuiteSectionRowPayload, 0, limit)
	maxRank := 0
	for i, bond := range bonds {
		if bond.Rank > maxRank {
			maxRank = bond.Rank
		}
		if i >= limit {
			continue
		}
		cid1, cid2 := bondCharacterIDs(bond)
		rows = append(rows, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: characterDisplayName(cid1) + " × " + characterDisplayName(cid2), Value: fmt.Sprintf("Lv.%d", bond.Rank), Meta: fmt.Sprintf("EXP %d", bond.Exp)})
	}
	return rows, []renderer.SuiteStatPayload{{Label: "羁绊组数", Value: formatInt(len(bonds))}, {Label: "最高羁绊", Value: fmt.Sprintf("Lv.%d", maxRank)}}
}

func rowsFromMusicProgress(profile musicProgressProfile) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	counts := musicProgressCounts(profile)
	rows := make([]renderer.SuiteSectionRowPayload, 0, len(counts))
	totalPlayed, totalClear, totalFC, totalAP := 0, 0, 0, 0
	for _, diff := range []string{"easy", "normal", "hard", "expert", "master", "append"} {
		count := counts[diff]
		if count == nil {
			continue
		}
		totalPlayed += count.Played
		totalClear += count.Clear
		totalFC += count.FullCombo
		totalAP += count.AllPerfect
		rows = append(rows, renderer.SuiteSectionRowPayload{Label: strings.ToUpper(diff), Value: fmt.Sprintf("Clear %d / FC %d / AP %d", count.Clear, count.FullCombo, count.AllPerfect), Meta: fmt.Sprintf("游玩 %d", count.Played), Color: suiteDifficultyColor(diff)})
	}
	return rows, []renderer.SuiteStatPayload{{Label: "游玩", Value: formatInt(totalPlayed)}, {Label: "Clear", Value: formatInt(totalClear)}, {Label: "FC", Value: formatInt(totalFC)}, {Label: "AP", Value: formatInt(totalAP)}}
}

func musicProgressCounts(profile musicProgressProfile) map[string]*musicProgressCount {
	counts := map[string]*musicProgressCount{}
	for _, result := range profile.UserMusicResults {
		diff := musicResultDifficulty(result)
		if diff == "" {
			continue
		}
		count := counts[diff]
		if count == nil {
			count = &musicProgressCount{}
			counts[diff] = count
		}
		count.Played++
		if musicResultCleared(result) {
			count.Clear++
		}
		if musicResultFullCombo(result) {
			count.FullCombo++
		}
		if musicResultAllPerfect(result) {
			count.AllPerfect++
		}
	}
	return counts
}

func rowsFromMaterials(profile materialProfile, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	materials := make([]userMaterial, 0, len(profile.UserMaterials))
	for _, material := range profile.UserMaterials {
		if material.Quantity <= 0 {
			continue
		}
		materials = append(materials, material)
	}
	sort.SliceStable(materials, func(i, j int) bool {
		if materials[i].Quantity == materials[j].Quantity {
			return materials[i].MaterialID < materials[j].MaterialID
		}
		return materials[i].Quantity > materials[j].Quantity
	})
	limit = clampLimit(limit, len(materials))
	rows := make([]renderer.SuiteSectionRowPayload, 0, limit)
	for i := 0; i < limit; i++ {
		material := materials[i]
		rows = append(rows, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: fmt.Sprintf("材料 #%d", material.MaterialID), Value: formatInt64(material.Quantity)})
	}
	return rows, []renderer.SuiteStatPayload{{Label: "金币", Value: formatInt64(profile.UserGamedata.Coin)}, {Label: "材料种类", Value: formatInt(len(materials))}}
}

func rowsFromChallenge(profile challengeProfile, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	rowsByCharacter := challengeRows(profile)
	rows := make([]challengeSummaryRow, 0, len(rowsByCharacter))
	for _, row := range rowsByCharacter {
		if row.HighScore == 0 && row.Rank == 0 && row.RewardCount == 0 {
			continue
		}
		rows = append(rows, *row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].HighScore == rows[j].HighScore {
			return rows[i].Rank > rows[j].Rank
		}
		return rows[i].HighScore > rows[j].HighScore
	})
	limit = clampLimit(limit, len(rows))
	out := make([]renderer.SuiteSectionRowPayload, 0, limit)
	maxScore := 0
	for i, row := range rows {
		if row.HighScore > maxScore {
			maxScore = row.HighScore
		}
		if i >= limit {
			continue
		}
		out = append(out, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: characterDisplayName(row.CharacterID), Value: formatInt(row.HighScore), Meta: fmt.Sprintf("Lv.%d · 奖励 %d", row.Rank, row.RewardCount)})
	}
	return out, []renderer.SuiteStatPayload{{Label: "角色数", Value: formatInt(len(rows))}, {Label: "最高分", Value: formatInt(maxScore)}}
}

func challengeRows(profile challengeProfile) map[int]*challengeSummaryRow {
	rowsByCharacter := map[int]*challengeSummaryRow{}
	for _, result := range profile.Results {
		row := challengeRow(rowsByCharacter, result.CharacterID)
		row.HighScore = max(row.HighScore, result.HighScore)
	}
	for _, stage := range profile.Stages {
		row := challengeRow(rowsByCharacter, stage.CharacterID)
		row.Rank = max(row.Rank, stage.Rank)
	}
	for _, reward := range profile.Rewards {
		row := challengeRow(rowsByCharacter, reward.CharacterID)
		row.RewardCount++
	}
	return rowsByCharacter
}

func rowsFromEventRecord(profile eventRecordProfile, store *masterdata.Store, limit int) ([]renderer.SuiteSectionPayload, []renderer.SuiteStatPayload) {
	events := append([]userEventRecord(nil), profile.UserEvents...)
	sort.SliceStable(events, func(i, j int) bool { return events[i].EventPoint > events[j].EventPoint })
	blooms := append([]userWorldBloomRecord(nil), profile.UserWorldBlooms...)
	sort.SliceStable(blooms, func(i, j int) bool { return worldBloomPoint(blooms[i]) > worldBloomPoint(blooms[j]) })
	if limit <= 0 {
		limit = max(len(events), len(blooms))
	}
	sections := make([]renderer.SuiteSectionPayload, 0, 2)
	if len(events) > 0 {
		rows := make([]renderer.SuiteSectionRowPayload, 0, min(limit, len(events)))
		for i := 0; i < min(limit, len(events)); i++ {
			event := events[i]
			rows = append(rows, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: eventName(store, event.EventID), Value: fmt.Sprintf("%dpt", event.EventPoint), Meta: formatRank(event.Rank)})
		}
		sections = append(sections, renderer.SuiteSectionPayload{Title: "活动PT", Rows: rows})
	}
	if len(blooms) > 0 {
		rows := make([]renderer.SuiteSectionRowPayload, 0, min(limit, len(blooms)))
		for i := 0; i < min(limit, len(blooms)); i++ {
			bloom := blooms[i]
			rows = append(rows, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: eventName(store, bloom.EventID), Value: fmt.Sprintf("%dpt", worldBloomPoint(bloom)), Meta: characterDisplayName(bloom.GameCharacterID) + " " + formatRank(bloom.WorldBloomChapterRank)})
		}
		sections = append(sections, renderer.SuiteSectionPayload{Title: "WL章节", Rows: rows})
	}
	return sections, []renderer.SuiteStatPayload{{Label: "活动记录", Value: formatInt(len(events))}, {Label: "WL记录", Value: formatInt(len(blooms))}}
}

func rowsFromLeaderCount(profile leaderCountProfile, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	rowsByCharacter := leaderRows(profile)
	rows := make([]leaderCountRow, 0, len(rowsByCharacter))
	for _, row := range rowsByCharacter {
		if row.PlayLive == 0 && row.PlayLiveEx == 0 {
			continue
		}
		rows = append(rows, *row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].PlayLive == rows[j].PlayLive {
			return rows[i].PlayLiveEx > rows[j].PlayLiveEx
		}
		return rows[i].PlayLive > rows[j].PlayLive
	})
	limit = clampLimit(limit, len(rows))
	out := make([]renderer.SuiteSectionRowPayload, 0, limit)
	total := 0
	for i, row := range rows {
		total += row.PlayLive
		if i >= limit {
			continue
		}
		out = append(out, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: characterDisplayName(row.CharacterID), Value: formatInt(row.PlayLive), Meta: fmt.Sprintf("EX %d", row.PlayLiveEx)})
	}
	return out, []renderer.SuiteStatPayload{{Label: "总队长次数", Value: formatInt(total)}, {Label: "角色数", Value: formatInt(len(rows))}}
}

func leaderRows(profile leaderCountProfile) map[int]*leaderCountRow {
	rowsByCharacter := map[int]*leaderCountRow{}
	for _, mission := range profile.Missions {
		if mission.CharacterID <= 0 {
			continue
		}
		row := rowsByCharacter[mission.CharacterID]
		if row == nil {
			row = &leaderCountRow{CharacterID: mission.CharacterID}
			rowsByCharacter[mission.CharacterID] = row
		}
		switch mission.CharacterMissionType {
		case "play_live":
			row.PlayLive = max(row.PlayLive, mission.Progress)
		case "play_live_ex":
			row.PlayLiveEx = max(row.PlayLiveEx, mission.Progress)
		}
	}
	return rowsByCharacter
}

func rowsFromMusicReward(profile musicRewardProfile, store *masterdata.Store, limit int) ([]renderer.SuiteSectionRowPayload, []renderer.SuiteStatPayload) {
	counts := map[int]int{}
	for _, achievement := range profile.Achievements {
		if achievement.MusicID <= 0 {
			continue
		}
		counts[achievement.MusicID]++
	}
	rows := make([]musicRewardRow, 0, len(counts))
	total := 0
	for mid, count := range counts {
		rows = append(rows, musicRewardRow{MusicID: mid, Count: count})
		total += count
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Count == rows[j].Count {
			return rows[i].MusicID < rows[j].MusicID
		}
		return rows[i].Count > rows[j].Count
	})
	limit = clampLimit(limit, len(rows))
	out := make([]renderer.SuiteSectionRowPayload, 0, limit)
	for i := 0; i < limit; i++ {
		row := rows[i]
		out = append(out, renderer.SuiteSectionRowPayload{Rank: i + 1, Label: musicDisplayName(store, row.MusicID), Value: formatInt(row.Count), Meta: "已达成奖励"})
	}
	return out, []renderer.SuiteStatPayload{{Label: "已达成奖励", Value: formatInt(total)}, {Label: "涉及歌曲", Value: formatInt(len(rows))}}
}

func musicDisplayName(store *masterdata.Store, musicID int) string {
	if store != nil {
		if music := store.GetMusic(musicID); music != nil && strings.TrimSpace(music.Title) != "" {
			return fmt.Sprintf("#%d %s", musicID, music.Title)
		}
	}
	return fmt.Sprintf("歌曲 #%d", musicID)
}

func suiteDifficultyColor(diff string) string {
	switch strings.ToLower(diff) {
	case "easy":
		return "#33ccbb"
	case "normal":
		return "#88dd44"
	case "hard":
		return "#ffb000"
	case "expert":
		return "#ff6699"
	case "master":
		return "#a863e8"
	case "append":
		return "#172033"
	default:
		return "#33ccbb"
	}
}

func formatRank(rank int) string {
	if rank <= 0 {
		return ""
	}
	return fmt.Sprintf("Rank %d", rank)
}

func clampLimit(limit int, total int) int {
	if limit <= 0 || limit > total {
		return total
	}
	return limit
}

func formatInt(value int) string { return fmt.Sprintf("%d", value) }
