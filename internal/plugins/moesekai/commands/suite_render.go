package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/plugins/moesekai/servers"
	"moebot-next/internal/plugins/moesekai/suite"
	"moebot-next/internal/renderer"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"moebot-next/internal/plugins/moesekai/renderpayloads"
)

const (
	leaderCountProgressMax  = 50000
	challengeProgressMax    = 3000000
	musicRewardRankRewardID = 4
)

type suiteCommandProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata `json:"userGamedata"`
	UserDecks    []suite.UserDeck   `json:"userDecks"`
	UserCards    []suite.UserCard   `json:"userCards"`
}

type musicAchievementReward struct {
	Coin  int
	Jewel int
	Shard int
}

type musicRewardSummary struct {
	RankJewelRemain    int
	RankRemainCount    int
	ValidMusicCount    int
	AchievementTotal   int
	AchievedMusicCount int
	ComboJewelRemain   int
	ComboShardRemain   int
	TotalJewelRemain   int
	TotalShardRemain   int
	ComboRows          []renderpayloads.SuiteSectionRowPayload
	TopRows            []renderpayloads.SuiteSectionRowPayload
}

type musicProgressSummary struct {
	Rows       []renderpayloads.SuiteSectionRowPayload
	LevelRows  []renderpayloads.SuiteSectionRowPayload
	Stats      []renderpayloads.SuiteStatPayload
	TotalSongs int
	TotalClear int
	TotalFC    int
	TotalAP    int
}

var musicRankRewards = map[int]musicAchievementReward{
	1: {Jewel: 10},
	2: {Jewel: 20},
	3: {Jewel: 30},
	4: {Jewel: 50},
}

var musicComboRewards = map[string]map[int]musicAchievementReward{
	"easy":   {5: {Coin: 500}, 6: {Coin: 1000}, 7: {Coin: 2000}, 8: {Coin: 5000}},
	"normal": {9: {Coin: 1000}, 10: {Coin: 2000}, 11: {Coin: 4000}, 12: {Coin: 10000}},
	"hard":   {13: {Coin: 1500}, 14: {Coin: 3000}, 15: {Coin: 6000}, 16: {Jewel: 50}},
	"expert": {17: {Coin: 2000}, 18: {Coin: 4000}, 19: {Jewel: 20}, 20: {Jewel: 50}},
	"master": {21: {Coin: 3000}, 22: {Coin: 6000}, 23: {Jewel: 20}, 24: {Jewel: 50}},
	"append": {25: {Coin: 3000}, 26: {Coin: 6000}, 27: {Shard: 5}, 28: {Shard: 10}},
}

func sendSuitePanelOrText(ctx *zero.Ctx, deps *Deps, payload renderpayloads.SuitePanelPayload, fallback string) {
	if deps != nil && deps.Renderer != nil {
		png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "suite_panel", Data: payload})
		if err == nil {
			ctx.SendChain(message.ImageBytes(png))
			return
		}
	}
	ctx.SendChain(message.Text(fallback))
}

func sendSuiteCardBoxOrText(ctx *zero.Ctx, deps *Deps, payload renderpayloads.SuiteCardBoxPayload, fallback string) {
	if deps != nil && deps.Renderer != nil {
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
}) renderpayloads.SuitePanelPayload {
	common := profile.commonSuiteProfile()
	region := ""
	if runtime != nil {
		region = runtime.Region
	}
	payload := renderpayloads.SuitePanelPayload{
		Title:       title,
		Profile:     renderpayloads.BuildSuiteProfilePayload(region, mode, common.BaseProfile, common.UserGamedata),
		AssetSource: suiteAssetSource(runtime),
	}
	if runtime != nil {
		payload.DeckCards = renderpayloads.BuildSuiteDeckCards(common.UserDecks, common.UserCards, common.UserGamedata.Deck, runtime.Store, runtime.Assets)
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

func (p musicOverviewProfile) commonSuiteProfile() suiteCommandProfile {
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

func suiteBasicStats(profile suiteCommandProfile) []renderpayloads.SuiteStatPayload {
	stats := make([]renderpayloads.SuiteStatPayload, 0, 4)
	if profile.UserGamedata.Rank > 0 {
		stats = append(stats, renderpayloads.SuiteStatPayload{Label: "Rank", Value: formatInt(profile.UserGamedata.Rank)})
	}
	if profile.UserGamedata.Coin > 0 {
		stats = append(stats, renderpayloads.SuiteStatPayload{Label: "金币", Value: formatInt64(profile.UserGamedata.Coin)})
	}
	if len(profile.UserCards) > 0 {
		stats = append(stats, renderpayloads.SuiteStatPayload{Label: "持有卡牌", Value: formatInt(len(profile.UserCards))})
	}
	return stats
}

func rowsFromGachaHistory(profile gachaHistoryProfile, store *masterdata.Store, limit int) ([]renderpayloads.SuiteSectionRowPayload, []renderpayloads.SuiteStatPayload) {
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
	rows := make([]renderpayloads.SuiteSectionRowPayload, 0, limit)
	for i := 0; i < limit; i++ {
		record := records[i]
		rows = append(rows, renderpayloads.SuiteSectionRowPayload{Rank: i + 1, Label: gachaHistoryName(store, record.GachaID), Value: fmt.Sprintf("%d抽", record.Count)})
	}
	return rows, []renderpayloads.SuiteStatPayload{{Label: "总抽数", Value: formatInt(total)}, {Label: "卡池数", Value: formatInt(len(records))}}
}

func rowsFromBonds(profile bondProfile, limit int) ([]renderpayloads.SuiteSectionRowPayload, []renderpayloads.SuiteStatPayload) {
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
	rows := make([]renderpayloads.SuiteSectionRowPayload, 0, limit)
	maxRank := 0
	for i, bond := range bonds {
		if bond.Rank > maxRank {
			maxRank = bond.Rank
		}
		if i >= limit {
			continue
		}
		cid1, cid2 := bondCharacterIDs(bond)
		name1, name2 := characterDisplayName(cid1), characterDisplayName(cid2)
		rows = append(rows, renderpayloads.SuiteSectionRowPayload{
			Rank:  i + 1,
			Label: name1 + " × " + name2,
			Value: fmt.Sprintf("Lv.%d", bond.Rank),
			Meta:  fmt.Sprintf("EXP %d", bond.Exp),
			Extra: map[string]interface{}{
				"characterId1":   cid1,
				"characterId2":   cid2,
				"characterName1": name1,
				"characterName2": name2,
				"rankLevel":      bond.Rank,
				"exp":            bond.Exp,
			},
		})
	}
	return rows, []renderpayloads.SuiteStatPayload{{Label: "羁绊组数", Value: formatInt(len(bonds))}, {Label: "最高羁绊", Value: fmt.Sprintf("Lv.%d", maxRank)}}
}

func rowsFromMusicProgress(profile musicProgressProfile) ([]renderpayloads.SuiteSectionRowPayload, []renderpayloads.SuiteStatPayload) {
	summary := musicProgressSummaryFromProfile(profile, nil)
	return summary.Rows, summary.Stats
}

func musicProgressSummaryFromProfile(profile musicOverviewProfile, store *masterdata.Store) musicProgressSummary {
	counts := musicProgressCounts(profile)
	levelCounts := musicProgressLevelCounts(profile, store)
	rows := make([]renderpayloads.SuiteSectionRowPayload, 0, len(counts))
	totalPlayed, totalClear, totalFC, totalAP := 0, 0, 0, 0
	for _, diff := range musicDifficultyOrder() {
		count := counts[diff]
		if count == nil {
			continue
		}
		totalPlayed += count.Played
		totalClear += count.Clear
		totalFC += count.FullCombo
		totalAP += count.AllPerfect
		rows = append(rows, renderpayloads.SuiteSectionRowPayload{
			Label: strings.ToUpper(diff),
			Value: fmt.Sprintf("Clear %d / FC %d / AP %d", count.Clear, count.FullCombo, count.AllPerfect),
			Meta:  fmt.Sprintf("游玩 %d", count.Played),
			Color: suiteDifficultyColor(diff),
			Extra: map[string]interface{}{"diff": diff, "played": count.Played, "clear": count.Clear, "fc": count.FullCombo, "ap": count.AllPerfect},
		})
	}
	levelRows := musicProgressLevelRows(levelCounts)
	return musicProgressSummary{
		Rows:       rows,
		LevelRows:  levelRows,
		Stats:      []renderpayloads.SuiteStatPayload{{Label: "游玩", Value: formatInt(totalPlayed)}, {Label: "Clear", Value: formatInt(totalClear)}, {Label: "FC", Value: formatInt(totalFC)}, {Label: "AP", Value: formatInt(totalAP)}},
		TotalSongs: totalPlayed,
		TotalClear: totalClear,
		TotalFC:    totalFC,
		TotalAP:    totalAP,
	}
}

func musicProgressCounts(profile musicOverviewProfile) map[string]*musicProgressCount {
	counts := map[string]*musicProgressCount{}
	best := bestMusicResults(profile.UserMusicResults)
	for _, result := range best {
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

func rowsFromMaterials(profile materialProfile, limit int) ([]renderpayloads.SuiteSectionRowPayload, []renderpayloads.SuiteStatPayload) {
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
	rows := make([]renderpayloads.SuiteSectionRowPayload, 0, limit)
	for i := 0; i < limit; i++ {
		material := materials[i]
		rows = append(rows, renderpayloads.SuiteSectionRowPayload{Rank: i + 1, Label: fmt.Sprintf("材料 #%d", material.MaterialID), Value: formatInt64(material.Quantity)})
	}
	return rows, []renderpayloads.SuiteStatPayload{{Label: "金币", Value: formatInt64(profile.UserGamedata.Coin)}, {Label: "材料种类", Value: formatInt(len(materials))}}
}

func rowsFromChallenge(profile challengeProfile, store *masterdata.Store, limit int) ([]renderpayloads.SuiteSectionRowPayload, []renderpayloads.SuiteStatPayload, map[string]interface{}) {
	rowsByCharacter := challengeRows(profile, store)
	activeCount := 0
	maxScore := 0
	completed := challengeCompletedRewardIDs(profile.Rewards, store)
	for cid, row := range rowsByCharacter {
		row.RemainJewel, row.RemainFragment = challengeRemainRewards(store, cid, completed[cid])
	}
	maxMasterScore := challengeMasterMaxScore(store)
	for _, row := range rowsByCharacter {
		if row.HighScore > 0 || row.Rank > 0 || row.RewardCount > 0 {
			activeCount++
		}
		if row.HighScore > maxScore {
			maxScore = row.HighScore
		}
	}
	displayMax := max(max(maxScore, maxMasterScore), challengeProgressMax)
	out := make([]renderpayloads.SuiteSectionRowPayload, 0, 26)
	totalRemainJewel, totalRemainFragment, totalRemainRewards := 0, 0, 0
	rankCounts := map[int]int{}
	for cid := 1; cid <= 26; cid++ {
		row := rowsByCharacter[cid]
		if row == nil {
			row = &challengeSummaryRow{CharacterID: cid}
			row.RemainJewel, row.RemainFragment = challengeRemainRewards(store, cid, completed[cid])
		}
		rewardRemain := challengeRemainRewardCount(store, cid, completed[cid])
		totalRemainJewel += row.RemainJewel
		totalRemainFragment += row.RemainFragment
		totalRemainRewards += rewardRemain
		rankCounts[row.Rank]++
		value := "-"
		if row.HighScore > 0 {
			value = formatInt(row.HighScore)
		}
		rankText := "-"
		if row.Rank > 0 {
			rankText = fmt.Sprintf("Lv.%d", row.Rank)
		}
		out = append(out, renderpayloads.SuiteSectionRowPayload{ID: cid, Rank: cid, Label: characterDisplayName(cid), Value: value, Meta: fmt.Sprintf("%s · 水晶 %d · 碎片 %d · 剩余档 %d", rankText, row.RemainJewel, row.RemainFragment, rewardRemain), CharacterID: cid, Progress: float64(row.HighScore), ProgressMax: float64(displayMax), ProgressLabel: fmt.Sprintf("%s / %s", formatInt(row.HighScore), formatInt(displayMax)), Extra: map[string]interface{}{"rankLevel": row.Rank, "rewardCount": row.RewardCount, "rewardRemain": rewardRemain, "remainJewel": row.RemainJewel, "remainFragment": row.RemainFragment, "highScore": row.HighScore, "jewel": row.RemainJewel, "shard": row.RemainFragment}})
	}
	extra := map[string]interface{}{"totalRemainJewel": totalRemainJewel, "totalRemainFragment": totalRemainFragment, "totalRemainRewards": totalRemainRewards, "totalRemainRewardSlots": totalRemainRewards, "rankDistribution": challengeRankDistribution(rankCounts)}
	stats := []renderpayloads.SuiteStatPayload{{Label: "角色数", Value: formatInt(activeCount)}, {Label: "最高分", Value: formatInt(maxScore)}, {Label: "剩余水晶", Value: formatInt(totalRemainJewel)}, {Label: "剩余碎片", Value: formatInt(totalRemainFragment)}, {Label: "剩余奖励档", Value: formatInt(totalRemainRewards)}}
	return out, stats, extra
}

func challengeRows(profile challengeProfile, store *masterdata.Store) map[int]*challengeSummaryRow {
	rowsByCharacter := map[int]*challengeSummaryRow{}
	rewardByID := challengeRewardMasterByID(store)
	for _, result := range profile.Results {
		if result.CharacterID <= 0 {
			continue
		}
		row := challengeRow(rowsByCharacter, result.CharacterID)
		row.HighScore = max(row.HighScore, result.HighScore)
	}
	for _, stage := range profile.Stages {
		if stage.CharacterID <= 0 {
			continue
		}
		row := challengeRow(rowsByCharacter, stage.CharacterID)
		row.Rank = max(row.Rank, stage.Rank)
	}
	for _, reward := range profile.Rewards {
		cid := challengeRewardCharacterID(reward)
		if cid <= 0 {
			if masterReward, ok := rewardByID[challengeRewardID(reward)]; ok {
				cid = masterReward.CharacterID
			}
		}
		if cid <= 0 {
			continue
		}
		row := challengeRow(rowsByCharacter, cid)
		row.RewardCount++
	}
	return rowsByCharacter
}

func challengeCompletedRewardIDs(rewards []challengeReward, store *masterdata.Store) map[int]map[int]struct{} {
	out := map[int]map[int]struct{}{}
	rewardByID := challengeRewardMasterByID(store)
	for _, reward := range rewards {
		rid := challengeRewardID(reward)
		if rid <= 0 {
			continue
		}
		cid := challengeRewardCharacterID(reward)
		if cid <= 0 {
			if masterReward, ok := rewardByID[rid]; ok {
				cid = masterReward.CharacterID
			}
		}
		if cid <= 0 {
			continue
		}
		ids := out[cid]
		if ids == nil {
			ids = map[int]struct{}{}
			out[cid] = ids
		}
		ids[rid] = struct{}{}
	}
	return out
}

func challengeRewardMasterByID(store *masterdata.Store) map[int]masterdata.ChallengeLiveHighScoreReward {
	out := map[int]masterdata.ChallengeLiveHighScoreReward{}
	if store == nil {
		return out
	}
	for _, reward := range store.AllChallengeLiveHighScoreRewards() {
		out[reward.ID] = reward
	}
	return out
}

func challengeRewardID(reward challengeReward) int {
	if reward.RewardID > 0 {
		return reward.RewardID
	}
	if reward.ChallengeLiveSoloHighScoreRewardID > 0 {
		return reward.ChallengeLiveSoloHighScoreRewardID
	}
	return reward.RewardIDAlias
}

func challengeRewardCharacterID(reward challengeReward) int {
	if reward.CharacterID > 0 {
		return reward.CharacterID
	}
	return reward.GameCharacterID
}

func challengeRemainRewards(store *masterdata.Store, characterID int, completed map[int]struct{}) (int, int) {
	if store == nil || characterID <= 0 {
		return 0, 0
	}
	jewel, fragment := 0, 0
	for _, reward := range store.GetChallengeLiveHighScoreRewards(characterID) {
		if _, ok := completed[reward.ID]; ok {
			continue
		}
		amount := collectChallengeResourceBox(store, reward.ResourceBoxID)
		jewel += amount.Jewel
		fragment += amount.Fragment
	}
	return jewel, fragment
}

type challengeRewardAmount struct {
	Jewel    int
	Fragment int
}

func collectChallengeResourceBox(store *masterdata.Store, rootBoxID int) challengeRewardAmount {
	return collectChallengeResourceBoxWithVisited(store, rootBoxID, map[int]struct{}{})
}

func collectChallengeResourceBoxWithVisited(store *masterdata.Store, boxID int, visited map[int]struct{}) challengeRewardAmount {
	if store == nil || boxID <= 0 {
		return challengeRewardAmount{}
	}
	if _, ok := visited[boxID]; ok {
		return challengeRewardAmount{}
	}
	visited[boxID] = struct{}{}
	details := challengeResourceBoxDetails(store, challengeLiveHighScorePurpose, boxID)
	amount := challengeRewardAmount{}
	for _, detail := range details {
		quantity := detail.ResourceQuantity
		if quantity <= 0 {
			continue
		}
		resourceType := strings.ToLower(detail.ResourceType)
		switch {
		case strings.Contains(resourceType, "jewel"):
			amount.Jewel += quantity
		case resourceType == "material" && detail.ResourceID == 15:
			amount.Fragment += quantity
		case strings.Contains(resourceType, "box"):
			nested := collectChallengeResourceBoxWithVisited(store, detail.ResourceID, visited)
			amount.Jewel += nested.Jewel
			amount.Fragment += nested.Fragment
		}
	}
	return amount
}

func challengeResourceBoxDetails(store *masterdata.Store, purpose string, boxID int) []masterdata.ResourceBoxDetail {
	if box := store.GetResourceBox(purpose, boxID); box != nil && len(box.Details) > 0 {
		return box.Details
	}
	return store.GetResourceBoxDetails(purpose, boxID)
}

func challengeRemainRewardCount(store *masterdata.Store, characterID int, completed map[int]struct{}) int {
	if store == nil || characterID <= 0 {
		return 0
	}
	count := 0
	for _, reward := range store.GetChallengeLiveHighScoreRewards(characterID) {
		if _, ok := completed[reward.ID]; ok {
			continue
		}
		count++
	}
	return count
}

func challengeRankDistribution(counts map[int]int) []map[string]interface{} {
	levels := make([]int, 0, len(counts))
	for level := range counts {
		levels = append(levels, level)
	}
	sort.SliceStable(levels, func(i, j int) bool { return levels[i] > levels[j] })
	out := make([]map[string]interface{}, 0, len(levels))
	for _, level := range levels {
		out = append(out, map[string]interface{}{"level": level, "count": counts[level], "label": challengeRankLabel(level)})
	}
	return out
}

func challengeRankLabel(level int) string {
	if level <= 0 {
		return "Lv.0"
	}
	return fmt.Sprintf("Lv.%d", level)
}

func challengeMasterMaxScore(store *masterdata.Store) int {
	if store == nil {
		return 0
	}
	maxScore := 0
	for _, reward := range store.AllChallengeLiveHighScoreRewards() {
		if reward.HighScore > maxScore {
			maxScore = reward.HighScore
		}
	}
	return maxScore
}

func rowsFromEventRecord(profile eventRecordProfile, store *masterdata.Store, resolver interface{ GetEventBannerURL(string) string }, limit int) ([]renderpayloads.SuiteSectionPayload, []renderpayloads.SuiteStatPayload) {
	events := append([]userEventRecord(nil), profile.UserEvents...)
	sortEventRecords(events)
	blooms := append([]userWorldBloomRecord(nil), profile.UserWorldBlooms...)
	sortWorldBloomRecords(blooms)
	if limit <= 0 {
		limit = max(len(events), len(blooms))
	}
	sections := make([]renderpayloads.SuiteSectionPayload, 0, 2)
	if len(events) > 0 {
		rows := make([]renderpayloads.SuiteSectionRowPayload, 0, min(limit, len(events)))
		for i := 0; i < min(limit, len(events)); i++ {
			event := events[i]
			rows = append(rows, eventRecordRow(store, resolver, event.EventID, event.EventPoint, event.Rank, 0, 0, i+1))
		}
		sections = append(sections, renderpayloads.SuiteSectionPayload{Title: "活动PT", Kind: "event_record", Note: "每次抓包仅包含最近活动记录；上传时增量更新，未上传过的记录可能缺失。", Rows: rows})
	}
	if len(blooms) > 0 {
		rows := make([]renderpayloads.SuiteSectionRowPayload, 0, min(limit, len(blooms)))
		for i := 0; i < min(limit, len(blooms)); i++ {
			bloom := blooms[i]
			rows = append(rows, eventRecordRow(store, resolver, bloom.EventID, worldBloomPoint(bloom), bloom.WorldBloomChapterRank, bloom.GameCharacterID, bloom.WorldBloomChapterNumber, i+1))
		}
		sections = append(sections, renderpayloads.SuiteSectionPayload{Title: "WL章节", Kind: "event_record_wl", Note: "WL 章节记录按章节 PT 排序，角色头像来自本地 assets/characters。", Rows: rows})
	}
	return sections, []renderpayloads.SuiteStatPayload{{Label: "活动记录", Value: formatInt(len(events))}, {Label: "WL记录", Value: formatInt(len(blooms))}}
}

func rowsFromLeaderCount(profile leaderCountProfile, store *masterdata.Store, limit int) ([]renderpayloads.SuiteSectionRowPayload, []renderpayloads.SuiteStatPayload, map[string]interface{}) {
	rowsByCharacter := leaderRows(profile)
	exLevels := leaderExLevels(profile.Statuses)
	exTotals := leaderExTotals(rowsByCharacter, exLevels, store)
	progressMax := leaderProgressMax(store)
	groups := leaderNormalGroups(store)
	maxLevel := len(groups)
	out := make([]renderpayloads.SuiteSectionRowPayload, 0, 26)
	total := 0
	totalRemain := 0
	totalMissionLevel := 0
	totalMissionRemain := 0
	totalEx := 0
	activeCount := 0
	for cid := 1; cid <= 26; cid++ {
		row := rowsByCharacter[cid]
		if row == nil {
			row = &leaderCountRow{CharacterID: cid}
		}
		total += row.PlayLive
		exTotal := exTotals[cid]
		totalEx += exTotal
		missionLevel := leaderMissionLevel(groups, row.PlayLive)
		missionRemain := max(maxLevel-missionLevel, 0)
		nextNeed := leaderNextNeed(groups, row.PlayLive)
		playLiveRemain := max(progressMax-row.PlayLive, 0)
		totalRemain += playLiveRemain
		totalMissionLevel += missionLevel
		totalMissionRemain += missionRemain
		if row.PlayLive > 0 || exTotal > 0 || exLevels[cid] > 0 {
			activeCount++
		}
		value := "-"
		if row.PlayLive > 0 {
			value = formatInt(row.PlayLive)
		}
		meta := fmt.Sprintf("剩余 %s · 档位 %d/%d · EX等级 x%d · EX次数 %s", formatInt(playLiveRemain), missionLevel, maxLevel, exLevels[cid], dashInt(exTotal))
		out = append(out, renderpayloads.SuiteSectionRowPayload{ID: cid, Rank: cid, Label: characterDisplayName(cid), Value: value, Meta: meta, CharacterID: cid, Progress: float64(row.PlayLive), ProgressMax: float64(progressMax), ProgressLabel: fmt.Sprintf("%s / %s", formatInt(row.PlayLive), formatInt(progressMax)), Extra: map[string]interface{}{"playLive": row.PlayLive, "playLiveRemain": playLiveRemain, "playLiveEx": exTotal, "playLiveExRaw": row.PlayLiveEx, "exLevel": exLevels[cid], "missionLevel": missionLevel, "missionLevelMax": maxLevel, "missionLevelRemain": missionRemain, "nextNeed": nextNeed, "progressRate": progressRate(float64(row.PlayLive), float64(progressMax))}})
	}
	totalMissionMax := maxLevel * 26
	extra := map[string]interface{}{"totalPlayLive": total, "totalRemain": totalRemain, "totalMissionLevel": totalMissionLevel, "totalMissionMax": totalMissionMax, "totalMissionRemain": totalMissionRemain, "totalEx": totalEx, "progressMax": progressMax}
	stats := []renderpayloads.SuiteStatPayload{{Label: "总队长次数", Value: formatInt(total)}, {Label: "剩余总次数", Value: formatInt(totalRemain)}, {Label: "普通档位", Value: fmt.Sprintf("%d/%d", totalMissionLevel, totalMissionMax)}, {Label: "剩余档位", Value: formatInt(totalMissionRemain)}, {Label: "EX总次数", Value: formatInt(totalEx)}, {Label: "角色数", Value: formatInt(activeCount)}}
	return out, stats, extra
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

func rowsFromMusicReward(profile musicRewardProfile, store *masterdata.Store, limit int) ([]renderpayloads.SuiteSectionRowPayload, []renderpayloads.SuiteStatPayload) {
	summary := musicRewardSummaryFromProfile(profile, store, limit)
	return summary.TopRows, []renderpayloads.SuiteStatPayload{{Label: "S评级剩余", Value: formatInt(summary.RankJewelRemain)}, {Label: "连击剩余", Value: formatRewardTotal(summary.ComboRows)}, {Label: "涉及歌曲", Value: formatInt(summary.AchievedMusicCount)}}
}

func sectionsFromMusicReward(profile musicRewardProfile, store *masterdata.Store, limit int) ([]renderpayloads.SuiteSectionPayload, []renderpayloads.SuiteStatPayload) {
	reward := musicRewardSummaryFromProfile(profile, store, limit)
	sections := musicRewardSections(reward)
	return sections, []renderpayloads.SuiteStatPayload{{Label: "S评级剩余", Value: formatInt(reward.RankJewelRemain)}, {Label: "剩余连击奖励", Value: formatRewardTotal(reward.ComboRows)}, {Label: "有效歌曲", Value: formatInt(reward.ValidMusicCount)}}
}

func sectionsFromMusicOverview(profile musicOverviewProfile, store *masterdata.Store, limit int) ([]renderpayloads.SuiteSectionPayload, []renderpayloads.SuiteStatPayload) {
	progress := musicProgressSummaryFromProfile(profile, store)
	reward := musicRewardSummaryFromProfile(profile, store, limit)
	sections := []renderpayloads.SuiteSectionPayload{
		{Title: "打歌进度", Kind: "music_progress_summary", Note: "按歌曲+难度去重后统计；同一谱面取最佳 Clear/FC/AP 状态。", Rows: progress.Rows, Extra: map[string]interface{}{"totalPlayed": progress.TotalSongs, "totalClear": progress.TotalClear, "totalFC": progress.TotalFC, "totalAP": progress.TotalAP}},
	}
	if len(progress.LevelRows) > 0 {
		sections = append(sections, renderpayloads.SuiteSectionPayload{Title: "等级数量", Kind: "music_progress_level", Rows: progress.LevelRows})
	}
	sections = append(sections, musicRewardSections(reward)...)
	stats := append([]renderpayloads.SuiteStatPayload{}, progress.Stats...)
	stats = append(stats,
		renderpayloads.SuiteStatPayload{Label: "S评级剩余", Value: formatInt(reward.RankJewelRemain)},
		renderpayloads.SuiteStatPayload{Label: "连击剩余", Value: formatRewardTotal(reward.ComboRows)},
	)
	return sections, stats
}

func musicRewardSections(summary musicRewardSummary) []renderpayloads.SuiteSectionPayload {
	sections := []renderpayloads.SuiteSectionPayload{{
		Title: "歌曲评级奖励(S)",
		Kind:  "music_reward_summary",
		Note:  "统计尚未获得的 S 评级水晶奖励；连击奖励按谱面等级汇总剩余值。",
		Rows: []renderpayloads.SuiteSectionRowPayload{
			{Label: "S评级剩余水晶", Value: formatInt(summary.RankJewelRemain), Meta: fmt.Sprintf("%d首未达成 / 共%d首", summary.RankRemainCount, summary.ValidMusicCount), Extra: map[string]interface{}{"rewardType": "jewel", "amount": summary.RankJewelRemain, "remainCount": summary.RankRemainCount, "validMusicCount": summary.ValidMusicCount}},
			{Label: "已达成奖励", Value: formatInt(summary.AchievementTotal), Meta: fmt.Sprintf("涉及%d首歌曲", summary.AchievedMusicCount), Extra: map[string]interface{}{"achievementTotal": summary.AchievementTotal, "achievedMusicCount": summary.AchievedMusicCount}},
		},
		Extra: musicRewardExtra(summary),
	}}
	if len(summary.ComboRows) > 0 {
		sections = append(sections, renderpayloads.SuiteSectionPayload{Title: "连击奖励剩余", Kind: "music_reward_combo", Rows: summary.ComboRows, Extra: musicRewardComboExtra(summary)})
	}
	if len(summary.TopRows) > 0 {
		sections = append(sections, renderpayloads.SuiteSectionPayload{Title: "已达成奖励 TOP", Kind: "music_reward_achieved", Rows: summary.TopRows})
	}
	return sections
}

func musicRewardSummaryFromProfile(profile musicRewardProfile, store *masterdata.Store, limit int) musicRewardSummary {
	achievements := musicAchievementsByMusic(profile.Achievements)
	validMusics := validMusicRewardMusics(store, achievements)
	summary := musicRewardSummary{ValidMusicCount: len(validMusics), AchievedMusicCount: len(achievements)}
	for _, ids := range achievements {
		summary.AchievementTotal += len(ids)
	}
	for _, music := range validMusics {
		ids := achievements[music.ID]
		if !ids[musicRewardRankRewardID] {
			summary.RankJewelRemain += musicRankRewards[musicRewardRankRewardID].Jewel
			summary.RankRemainCount++
		}
	}
	summary.ComboRows = musicRewardComboRows(store, validMusics, achievements)
	summary.ComboJewelRemain, summary.ComboShardRemain = rewardTotals(summary.ComboRows)
	summary.TotalJewelRemain = summary.RankJewelRemain + summary.ComboJewelRemain
	summary.TotalShardRemain = summary.ComboShardRemain
	summary.TopRows = musicRewardTopRows(store, achievements, limit)
	return summary
}

func musicRewardExtra(summary musicRewardSummary) map[string]interface{} {
	return map[string]interface{}{
		"rankJewelRemain":    summary.RankJewelRemain,
		"rankRemainCount":    summary.RankRemainCount,
		"validMusicCount":    summary.ValidMusicCount,
		"achievementTotal":   summary.AchievementTotal,
		"achievedMusicCount": summary.AchievedMusicCount,
		"comboJewelRemain":   summary.ComboJewelRemain,
		"comboShardRemain":   summary.ComboShardRemain,
		"totalJewelRemain":   summary.TotalJewelRemain,
		"totalShardRemain":   summary.TotalShardRemain,
	}
}

func musicRewardComboExtra(summary musicRewardSummary) map[string]interface{} {
	extra := musicRewardExtra(summary)
	extra["total"] = formatRewardTotal(summary.ComboRows)
	return extra
}

func musicAchievementsByMusic(achievements []musicAchievement) map[int]map[int]bool {
	out := map[int]map[int]bool{}
	for _, achievement := range achievements {
		if achievement.MusicID <= 0 || achievement.MusicAchievementID <= 0 {
			continue
		}
		ids := out[achievement.MusicID]
		if ids == nil {
			ids = map[int]bool{}
			out[achievement.MusicID] = ids
		}
		ids[achievement.MusicAchievementID] = true
	}
	return out
}

func validMusicRewardMusics(store *masterdata.Store, achievements map[int]map[int]bool) []masterdata.MusicInfo {
	if store != nil && store.IsLoaded() {
		musics := store.AllMusics()
		now := time.Now().UnixMilli()
		out := make([]masterdata.MusicInfo, 0, len(musics))
		for _, music := range musics {
			if music.ID <= 0 || (music.PublishedAt > 0 && music.PublishedAt > now) {
				continue
			}
			out = append(out, music)
		}
		sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
		return out
	}
	ids := make([]int, 0, len(achievements))
	for mid := range achievements {
		ids = append(ids, mid)
	}
	sort.Ints(ids)
	out := make([]masterdata.MusicInfo, 0, len(ids))
	for _, mid := range ids {
		out = append(out, masterdata.MusicInfo{ID: mid, Title: fmt.Sprintf("歌曲 #%d", mid)})
	}
	return out
}

func musicRewardTopRows(store *masterdata.Store, achievements map[int]map[int]bool, limit int) []renderpayloads.SuiteSectionRowPayload {
	rows := make([]musicRewardRow, 0, len(achievements))
	for mid, ids := range achievements {
		rows = append(rows, musicRewardRow{MusicID: mid, Count: len(ids)})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Count == rows[j].Count {
			return rows[i].MusicID < rows[j].MusicID
		}
		return rows[i].Count > rows[j].Count
	})
	limit = clampLimit(limit, len(rows))
	out := make([]renderpayloads.SuiteSectionRowPayload, 0, limit)
	for i := 0; i < limit; i++ {
		row := rows[i]
		out = append(out, renderpayloads.SuiteSectionRowPayload{Rank: i + 1, Label: musicDisplayName(store, row.MusicID), Value: formatInt(row.Count), Meta: "已达成奖励", MusicID: row.MusicID})
	}
	return out
}

func musicRewardComboRows(store *masterdata.Store, musics []masterdata.MusicInfo, achievements map[int]map[int]bool) []renderpayloads.SuiteSectionRowPayload {
	type rewardBucket struct {
		Diff       string
		Level      int
		Amount     int
		RewardType string
		Count      int
	}
	buckets := map[string]*rewardBucket{}
	for _, music := range musics {
		ids := achievements[music.ID]
		for _, diff := range []string{"hard", "expert", "master", "append"} {
			level := musicDifficultyLevel(store, music.ID, diff)
			if level <= 0 {
				continue
			}
			amount := 0
			rewardType := "jewel"
			for achievementID, reward := range musicComboRewards[diff] {
				if ids[achievementID] {
					continue
				}
				if diff == "append" {
					amount += reward.Shard
					rewardType = "shard"
				} else {
					amount += reward.Jewel
				}
			}
			if amount <= 0 {
				continue
			}
			key := fmt.Sprintf("%s:%d", diff, level)
			bucket := buckets[key]
			if bucket == nil {
				bucket = &rewardBucket{Diff: diff, Level: level, RewardType: rewardType}
				buckets[key] = bucket
			}
			bucket.Amount += amount
			bucket.Count++
		}
	}
	ordered := make([]*rewardBucket, 0, len(buckets))
	for _, bucket := range buckets {
		ordered = append(ordered, bucket)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if diffOrder(ordered[i].Diff) != diffOrder(ordered[j].Diff) {
			return diffOrder(ordered[i].Diff) < diffOrder(ordered[j].Diff)
		}
		return ordered[i].Level < ordered[j].Level
	})
	accByDiff := map[string]int{}
	rows := make([]renderpayloads.SuiteSectionRowPayload, 0, len(ordered))
	for _, bucket := range ordered {
		accByDiff[bucket.Diff] += bucket.Amount
		rows = append(rows, renderpayloads.SuiteSectionRowPayload{Label: bucket.Diff, Value: formatInt(bucket.Amount), Meta: fmt.Sprintf("Lv.%d · 累计 %d · %d谱面", bucket.Level, accByDiff[bucket.Diff], bucket.Count), Color: suiteDifficultyColor(bucket.Diff), Extra: map[string]interface{}{"diff": bucket.Diff, "level": bucket.Level, "amount": bucket.Amount, "accumulate": accByDiff[bucket.Diff], "rewardType": bucket.RewardType, "count": bucket.Count}})
	}
	return rows
}

func bestMusicResults(results []userMusicResult) map[string]userMusicResult {
	best := map[string]userMusicResult{}
	for _, result := range results {
		diff := musicResultDifficulty(result)
		if result.MusicID <= 0 || diff == "" {
			continue
		}
		key := fmt.Sprintf("%d:%s", result.MusicID, diff)
		if prev, ok := best[key]; !ok || musicResultRank(result) > musicResultRank(prev) {
			best[key] = result
		}
	}
	return best
}

func musicResultRank(result userMusicResult) int {
	if musicResultAllPerfect(result) {
		return 3
	}
	if musicResultFullCombo(result) {
		return 2
	}
	if musicResultCleared(result) {
		return 1
	}
	return 0
}

func musicProgressLevelCounts(profile musicOverviewProfile, store *masterdata.Store) map[string]map[int]*musicProgressCount {
	counts := map[string]map[int]*musicProgressCount{}
	if store != nil && store.IsLoaded() {
		now := time.Now().UnixMilli()
		for _, music := range store.AllMusics() {
			if music.ID <= 0 || (music.PublishedAt > 0 && music.PublishedAt > now) {
				continue
			}
			for _, diffInfo := range store.GetMusicDifficulties(music.ID) {
				diff := strings.ToLower(diffInfo.MusicDifficulty)
				if diff == "" || diffInfo.PlayLevel <= 0 {
					continue
				}
				musicProgressLevelCount(counts, diff, diffInfo.PlayLevel).Total++
			}
		}
	}
	for _, result := range bestMusicResults(profile.UserMusicResults) {
		diff := musicResultDifficulty(result)
		level := musicDifficultyLevel(store, result.MusicID, diff)
		if diff == "" || level <= 0 {
			continue
		}
		count := musicProgressLevelCount(counts, diff, level)
		if count.Total <= 0 {
			count.Total = 1
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

func musicProgressLevelCount(counts map[string]map[int]*musicProgressCount, diff string, level int) *musicProgressCount {
	byLevel := counts[diff]
	if byLevel == nil {
		byLevel = map[int]*musicProgressCount{}
		counts[diff] = byLevel
	}
	count := byLevel[level]
	if count == nil {
		count = &musicProgressCount{}
		byLevel[level] = count
	}
	return count
}

func musicProgressLevelRows(counts map[string]map[int]*musicProgressCount) []renderpayloads.SuiteSectionRowPayload {
	rows := []renderpayloads.SuiteSectionRowPayload{}
	for _, diff := range musicDifficultyOrder() {
		byLevel := counts[diff]
		if len(byLevel) == 0 {
			continue
		}
		levels := make([]int, 0, len(byLevel))
		for level := range byLevel {
			levels = append(levels, level)
		}
		sort.Ints(levels)
		for _, level := range levels {
			count := byLevel[level]
			total := count.Total
			if total < count.Played {
				total = count.Played
			}
			notPlayed := max(total-count.Played, 0)
			clearOnly := max(count.Clear-count.FullCombo, 0)
			fcOnly := max(count.FullCombo-count.AllPerfect, 0)
			rows = append(rows, renderpayloads.SuiteSectionRowPayload{
				Label: strings.ToUpper(diff),
				Value: fmt.Sprintf("Clear %d / FC %d / AP %d", count.Clear, count.FullCombo, count.AllPerfect),
				Meta:  fmt.Sprintf("Lv.%d · 游玩 %d/%d", level, count.Played, total),
				Color: suiteDifficultyColor(diff),
				Extra: map[string]interface{}{"diff": diff, "level": level, "total": total, "played": count.Played, "clear": count.Clear, "fc": count.FullCombo, "ap": count.AllPerfect, "notPlayed": notPlayed, "clearOnly": clearOnly, "fcOnly": fcOnly, "apOnly": count.AllPerfect},
			})
		}
	}
	return rows
}

func musicDifficultyLevel(store *masterdata.Store, musicID int, diff string) int {
	if store == nil {
		return 0
	}
	for _, item := range store.GetMusicDifficulties(musicID) {
		if strings.EqualFold(item.MusicDifficulty, diff) {
			return item.PlayLevel
		}
	}
	return 0
}

func musicDisplayName(store *masterdata.Store, musicID int) string {
	if store != nil {
		if music := store.GetMusic(musicID); music != nil && strings.TrimSpace(music.Title) != "" {
			return fmt.Sprintf("#%d %s", musicID, music.Title)
		}
	}
	return fmt.Sprintf("歌曲 #%d", musicID)
}

func eventRecordRow(store *masterdata.Store, resolver interface{ GetEventBannerURL(string) string }, eventID int, point int, rank int, characterID int, chapterNo int, order int) renderpayloads.SuiteSectionRowPayload {
	label := eventName(store, eventID)
	var startAt, endAt int64
	bannerURL := ""
	dateText := ""
	if store != nil {
		if event := store.GetEvent(eventID); event != nil {
			label = event.Name
			if strings.TrimSpace(label) == "" {
				label = fmt.Sprintf("活动 #%d", eventID)
			}
			startAt = event.StartAt
			endAt = event.AggregateAt
			if endAt <= 0 {
				endAt = event.ClosedAt
			}
			dateText = eventDateRange(startAt, endAt)
			if resolver != nil && event.AssetbundleName != "" {
				bannerURL = resolver.GetEventBannerURL(event.AssetbundleName)
			}
		}
	}
	metaParts := []string{}
	if rank > 0 {
		metaParts = append(metaParts, formatRank(rank))
	}
	if characterID > 0 {
		metaParts = append(metaParts, characterDisplayName(characterID))
	}
	if chapterNo > 0 {
		metaParts = append(metaParts, fmt.Sprintf("第%d章", chapterNo))
	}
	return renderpayloads.SuiteSectionRowPayload{ID: eventID, Rank: order, Label: label, Value: fmt.Sprintf("%dpt", point), Meta: strings.Join(metaParts, " · "), EventID: eventID, CharacterID: characterID, BannerURL: bannerURL, DateText: dateText, StartAt: startAt, EndAt: endAt, Extra: map[string]interface{}{"point": point, "rank": rank, "chapterNo": chapterNo}}
}

func sortEventRecords(events []userEventRecord) {
	hasRank := false
	for _, event := range events {
		if event.Rank > 0 {
			hasRank = true
			break
		}
	}
	sort.SliceStable(events, func(i, j int) bool {
		if hasRank {
			ir, jr := normalizedRank(events[i].Rank), normalizedRank(events[j].Rank)
			if ir != jr {
				return ir < jr
			}
		}
		return events[i].EventPoint > events[j].EventPoint
	})
}

func sortWorldBloomRecords(blooms []userWorldBloomRecord) {
	hasRank := false
	for _, bloom := range blooms {
		if bloom.WorldBloomChapterRank > 0 {
			hasRank = true
			break
		}
	}
	sort.SliceStable(blooms, func(i, j int) bool {
		if hasRank {
			ir, jr := normalizedRank(blooms[i].WorldBloomChapterRank), normalizedRank(blooms[j].WorldBloomChapterRank)
			if ir != jr {
				return ir < jr
			}
		}
		return worldBloomPoint(blooms[i]) > worldBloomPoint(blooms[j])
	})
}

func leaderExLevels(statuses []characterMissionV2Status) map[int]int {
	out := map[int]int{}
	for _, status := range statuses {
		if status.CharacterID <= 0 || status.ParameterGroupID != 101 {
			continue
		}
		out[status.CharacterID] = max(out[status.CharacterID], status.Seq)
	}
	return out
}

func leaderExTotals(rows map[int]*leaderCountRow, exLevels map[int]int, store *masterdata.Store) map[int]int {
	out := map[int]int{}
	for cid := 1; cid <= 26; cid++ {
		progressRaw := 0
		if row := rows[cid]; row != nil {
			progressRaw = row.PlayLiveEx
		}
		clearedTotal := leaderExClearedTotal(store, exLevels[cid])
		currentTotal := progressRaw
		if progressRaw < clearedTotal {
			currentTotal = clearedTotal + progressRaw
		} else if progressRaw == 0 {
			currentTotal = clearedTotal
		}
		out[cid] = currentTotal
	}
	return out
}

func leaderExClearedTotal(store *masterdata.Store, seq int) int {
	if store == nil || seq <= 0 {
		return 0
	}
	groups := store.GetCharacterMissionV2ParameterGroups(101)
	if len(groups) == 0 {
		return 0
	}
	total := 0
	for round := 1; round <= seq; round++ {
		total += leaderRequirementForRound(groups, round)
	}
	return total
}

func leaderRequirementForRound(groups []masterdata.CharacterMissionV2ParameterGroup, round int) int {
	req := 0
	for _, group := range groups {
		if group.Seq > round {
			break
		}
		req = group.Requirement
	}
	return req
}

func leaderProgressMax(store *masterdata.Store) int {
	if store == nil {
		return leaderCountProgressMax
	}
	maxReq := 0
	for _, group := range store.GetCharacterMissionV2ParameterGroups(1) {
		if group.Requirement > maxReq {
			maxReq = group.Requirement
		}
	}
	if maxReq <= 0 {
		return leaderCountProgressMax
	}
	return maxReq
}

func leaderNormalGroups(store *masterdata.Store) []masterdata.CharacterMissionV2ParameterGroup {
	if store == nil {
		return nil
	}
	return store.GetCharacterMissionV2ParameterGroups(1)
}

func leaderMissionLevel(groups []masterdata.CharacterMissionV2ParameterGroup, playLive int) int {
	level := 0
	for _, group := range groups {
		if group.Requirement <= playLive {
			level++
		}
	}
	return level
}

func leaderNextNeed(groups []masterdata.CharacterMissionV2ParameterGroup, playLive int) int {
	for _, group := range groups {
		if group.Requirement > playLive {
			return group.Requirement - playLive
		}
	}
	return 0
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

func eventDateRange(startAt int64, endAt int64) string {
	if startAt <= 0 && endAt <= 0 {
		return ""
	}
	if startAt <= 0 {
		return formatEventDate(endAt)
	}
	if endAt <= 0 {
		return formatEventDate(startAt)
	}
	return fmt.Sprintf("%s - %s", formatEventDate(startAt), formatEventDate(endAt))
}

func formatEventDate(value int64) string {
	if value <= 0 {
		return "-"
	}
	return time.UnixMilli(normalizeSuiteMillis(value)).Format("2006-01-02")
}

func normalizeSuiteMillis(value int64) int64 {
	if value > 0 && value < 100000000000 {
		return value * 1000
	}
	return value
}

func normalizedRank(rank int) int {
	if rank <= 0 {
		return 1 << 30
	}
	return rank
}

func progressRate(value float64, maxValue float64) float64 {
	if maxValue <= 0 || value <= 0 {
		return 0
	}
	if value >= maxValue {
		return 1
	}
	return value / maxValue
}

func dashInt(value int) string {
	if value <= 0 {
		return "-"
	}
	return formatInt(value)
}

func musicDifficultyOrder() []string {
	return []string{"easy", "normal", "hard", "expert", "master", "append"}
}

func diffOrder(diff string) int {
	switch strings.ToLower(diff) {
	case "easy":
		return 1
	case "normal":
		return 2
	case "hard":
		return 3
	case "expert":
		return 4
	case "master":
		return 5
	case "append":
		return 6
	default:
		return 99
	}
}

func rewardTotals(rows []renderpayloads.SuiteSectionRowPayload) (int, int) {
	jewel, shard := 0, 0
	for _, row := range rows {
		if row.Extra == nil {
			continue
		}
		amount, _ := row.Extra["amount"].(int)
		rewardType, _ := row.Extra["rewardType"].(string)
		if rewardType == "shard" {
			shard += amount
		} else {
			jewel += amount
		}
	}
	return jewel, shard
}

func formatRewardTotal(rows []renderpayloads.SuiteSectionRowPayload) string {
	jewel, shard := rewardTotals(rows)
	parts := []string{}
	if jewel > 0 {
		parts = append(parts, fmt.Sprintf("%d水晶", jewel))
	}
	if shard > 0 {
		parts = append(parts, fmt.Sprintf("%d碎片", shard))
	}
	if len(parts) == 0 {
		return "0"
	}
	return strings.Join(parts, " / ")
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
