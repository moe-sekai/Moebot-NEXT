package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/models"
	"moebot-next/internal/plugins/moesekai/ranking"
	"moebot-next/internal/plugins/moesekai/servers"
	"moebot-next/internal/renderer"

	"moebot-next/internal/plugins/moesekai/renderpayloads"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

func RegisterRanking(deps *Deps) {
	registerRankingBoardCommands(deps)
	registerRankingTargetCommands(deps)
	registerRankingChurnCommands(deps)
	registerWaterTableCommands(deps)
	registerForecastCommands(deps)
}

func registerRankingBoardCommands(deps *Deps) {
	for _, cmd := range withWorldLinkCommands(parserCommands(deps, "榜线")) {
		cmd := cmd
		Engine.OnCommand(cmd.Name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, _ := runtimeForCommand(deps, ctx, cmd.Region)
			if !rankingRuntimeOK(ctx, runtime, "榜线") {
				return
			}
			board, err := fetchRankingBoard(runtime.Ranking, cmd.WorldLink, commandArgs(ctx))
			if err != nil {
				ctx.SendChain(message.Text(rankingErrorText(err, cmd.WorldLink, "榜线")))
				return
			}
			view := *board
			view.Rankings = filterRankingsByRanks(board.Rankings, defaultBorderRanks())
			if len(view.Rankings) == 0 {
				view.Rankings = topRankings(board.Rankings, 10)
			}
			title := rankingTitle("活动榜线", cmd.WorldLink, view.TargetID)
			payload := renderpayloads.BuildRankingListPayloadWithStore(title, view, runtime.Store, runtime.Assets)
			payload.Subtitle = rankingSubtitle(runtime.Region, view)
			if sendRankingImage(ctx, deps.Renderer, payload) {
				bot.RecordCommandRegion(deps.DB, "榜线", runtime.Region, ctx, start)
				return
			}
			ctx.SendChain(message.Text(formatRankingText(payload)))
			bot.RecordCommandRegion(deps.DB, "榜线", runtime.Region, ctx, start)
		})
	}
}

func registerRankingTargetCommands(deps *Deps) {
	for _, cmd := range withWorldLinkCommands(parserCommands(deps, "sk")) {
		cmd := cmd
		Engine.On("message", rankingCommandRule(cmd.Name, true)).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, inferredUser := runtimeForCommand(deps, ctx, cmd.Region)
			if !rankingRuntimeOK(ctx, runtime, "sk") {
				return
			}
			board, err := fetchRankingBoard(runtime.Ranking, cmd.WorldLink, commandArgs(ctx))
			if err != nil {
				ctx.SendChain(message.Text(rankingErrorText(err, cmd.WorldLink, "sk")))
				return
			}
			query, err := parseRankingQuery(commandArgs(ctx), inferredUser, deps, ctx, runtime.Region, cmd.WorldLink, false)
			if err != nil {
				ctx.SendChain(message.Text(err.Error()))
				return
			}
			entries := selectRankingEntries(board.Rankings, query)
			if len(entries) == 0 {
				ctx.SendChain(message.Text("没有找到对应玩家/排名的榜线数据"))
				return
			}
			view := *board
			view.Rankings = entries
			title := rankingTitle("sk", cmd.WorldLink, view.TargetID)
			payload := renderpayloads.BuildRankingListPayloadWithStore(title, view, runtime.Store, runtime.Assets)
			payload.Subtitle = rankingSubtitle(runtime.Region, view)
			if sendRankingImage(ctx, deps.Renderer, payload) {
				bot.RecordCommandRegion(deps.DB, "sk", runtime.Region, ctx, start)
				return
			}
			ctx.SendChain(message.Text(formatRankingText(payload)))
			bot.RecordCommandRegion(deps.DB, "sk", runtime.Region, ctx, start)
		})
	}
}

func registerRankingChurnCommands(deps *Deps) {
	for _, cmd := range withWorldLinkCommands(parserCommands(deps, "查房")) {
		cmd := cmd
		Engine.On("message", rankingCommandRule(cmd.Name, true)).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, inferredUser := runtimeForCommand(deps, ctx, cmd.Region)
			if !rankingRuntimeOK(ctx, runtime, "查房") {
				return
			}
			board, err := fetchChurnBoard(runtime.Ranking, cmd.WorldLink, commandArgs(ctx))
			if err != nil {
				ctx.SendChain(message.Text(rankingErrorText(err, cmd.WorldLink, "查房")))
				return
			}
			hydrateChurnBoardAvatars(runtime.Ranking, board, cmd.WorldLink, commandArgs(ctx))
			query, err := parseRankingQuery(commandArgs(ctx), inferredUser, deps, ctx, runtime.Region, cmd.WorldLink, false)
			if err != nil {
				query = rankingQuery{Ranks: defaultBorderRanks()}
				if inferredUser != nil && inferredUser.GameID != "" {
					query.UserID = inferredUser.GameID
				}
			}
			entries := selectRankingEntries(board.Rankings, query)
			if len(entries) == 0 {
				entries = filterRankingsByRanks(board.Rankings, defaultBorderRanks())
			}
			if len(entries) == 0 {
				entries = topRankings(board.Rankings, 10)
			}
			view := *board
			view.Rankings = entries
			payload := renderpayloads.BuildChurnRankingListPayloadWithStore(view, runtime.Store, runtime.Assets)
			payload.Title = rankingTitle("查房", cmd.WorldLink, view.TargetID)
			payload.Subtitle = rankingSubtitle(runtime.Region, view)
			if sendRankingImageWithTemplate(ctx, deps.Renderer, "churn_ranking_list", payload) {
				bot.RecordCommandRegion(deps.DB, "查房", runtime.Region, ctx, start)
				return
			}
			ctx.SendChain(message.Text(formatRankingText(payload)))
			bot.RecordCommandRegion(deps.DB, "查房", runtime.Region, ctx, start)
		})
	}
}

func registerWaterTableCommands(deps *Deps) {
	for _, cmd := range withWorldLinkCommands(parserCommands(deps, "查水表")) {
		cmd := cmd
		Engine.On("message", rankingCommandRule(cmd.Name, true)).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, inferredUser := runtimeForCommand(deps, ctx, cmd.Region)
			if !rankingRuntimeOK(ctx, runtime, "查水表") {
				return
			}
			board, err := fetchChurnBoard(runtime.Ranking, cmd.WorldLink, commandArgs(ctx))
			if err != nil {
				ctx.SendChain(message.Text(rankingErrorText(err, cmd.WorldLink, "查水表")))
				return
			}
			hydrateChurnBoardAvatars(runtime.Ranking, board, cmd.WorldLink, commandArgs(ctx))
			query, err := parseRankingQuery(commandArgs(ctx), inferredUser, deps, ctx, runtime.Region, cmd.WorldLink, true)
			if err != nil {
				ctx.SendChain(message.Text(err.Error()))
				return
			}
			entries := selectRankingEntries(board.Rankings, query)
			if len(entries) == 0 {
				ctx.SendChain(message.Text("没有找到对应玩家/排名的查水表数据"))
				return
			}
			payload := renderpayloads.BuildWaterTablePayloadWithStore(*board, entries[0], runtime.Store, runtime.Assets)
			payload.Title = rankingTitle("查水表", cmd.WorldLink, board.TargetID)
			payload.Subtitle = rankingSubtitle(runtime.Region, *board)
			if sendGenericImage(ctx, deps.Renderer, "water_table", payload) {
				bot.RecordCommandRegion(deps.DB, "查水表", runtime.Region, ctx, start)
				return
			}
			ctx.SendChain(message.Text(formatWaterTableText(payload)))
			bot.RecordCommandRegion(deps.DB, "查水表", runtime.Region, ctx, start)
		})
	}
}

func registerForecastCommands(deps *Deps) {
	for _, cmd := range parserCommands(deps, "榜线预测") {
		cmd := cmd
		Engine.OnCommand(cmd.Name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, _ := runtimeForCommand(deps, ctx, cmd.Region)
			if !rankingRuntimeOK(ctx, runtime, "榜线预测") {
				return
			}
			if runtime.Region != config.RegionCN && runtime.Region != config.RegionJP {
				ctx.SendChain(message.Text("榜线预测仅支持国服/日服"))
				return
			}
			event, board, err := fetchForecast(runtime.Ranking, commandArgs(ctx))
			if err != nil {
				ctx.SendChain(message.Text(rankingErrorText(err, false, "榜线预测")))
				return
			}
			payload := renderpayloads.BuildForecastRankingPayload(*board, event.Name, runtime.Region, regionLabel(runtime.Region))
			if sendGenericImage(ctx, deps.Renderer, "forecast_ranking_list", payload) {
				bot.RecordCommandRegion(deps.DB, "榜线预测", runtime.Region, ctx, start)
				return
			}
			ctx.SendChain(message.Text(formatForecastText(payload)))
			bot.RecordCommandRegion(deps.DB, "榜线预测", runtime.Region, ctx, start)
		})
	}
}

type rankingQuery struct {
	Ranks  []int
	UserID string
}

func rankingCommandRule(command string, allowInlineArgument bool) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		if len(ctx.Event.Message) == 0 || ctx.Event.Message[0].Type != "text" {
			return false
		}
		first := ctx.Event.Message[0]
		firstMessage := first.Data["text"]
		prefix := zero.BotConfig.CommandPrefix
		if !strings.HasPrefix(firstMessage, prefix) {
			return false
		}
		cmdMessage := firstMessage[len(prefix):]
		args, ok := splitRankingCommandArgs(cmdMessage, command, allowInlineArgument)
		if !ok {
			return false
		}
		if len(ctx.Event.Message) > 1 {
			args += ctx.Event.Message[1:].ExtractPlainText()
		}
		ctx.State["command"] = command
		ctx.State["args"] = strings.TrimSpace(args)
		return true
	}
}

func splitRankingCommandArgs(cmdMessage string, command string, allowInlineArgument bool) (string, bool) {
	if len(cmdMessage) < len(command) || !strings.EqualFold(cmdMessage[:len(command)], command) {
		return "", false
	}
	rest := cmdMessage[len(command):]
	if rest == "" {
		return "", true
	}
	if strings.TrimSpace(rest) == "" || startsWithSpace(rest) {
		return strings.TrimSpace(rest), true
	}
	if allowInlineArgument && startsWithRankingInlineArgument(rest) {
		return strings.TrimSpace(rest), true
	}
	return "", false
}

func startsWithSpace(value string) bool {
	if value == "" {
		return false
	}
	first, _ := utf8.DecodeRuneInString(value)
	return unicode.IsSpace(first)
}

func startsWithRankingInlineArgument(value string) bool {
	if value == "" {
		return false
	}
	first, _ := utf8.DecodeRuneInString(value)
	return first == '#' || first == '+' || first == '-' || first == '＋' || first == '－' || (first >= '0' && first <= '9')
}

func withWorldLinkCommands(base []regionalCommand) []regionalCommand {
	out := make([]regionalCommand, 0, len(base)*2)
	seen := map[string]struct{}{}
	add := func(cmd regionalCommand) {
		key := strings.ToLower(cmd.Name)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, cmd)
	}
	for _, cmd := range base {
		add(cmd)
		wl := cmd
		wl.WorldLink = true
		if cmd.Region != "" && strings.HasPrefix(cmd.Name, cmd.Region) {
			wl.Name = cmd.Region + "wl" + strings.TrimPrefix(cmd.Name, cmd.Region)
		} else {
			wl.Name = "wl" + cmd.Name
		}
		wl.Base = "wl" + cmd.Base
		add(wl)
	}
	return out
}

func rankingRuntimeOK(ctx *zero.Ctx, runtime *servers.Runtime, command string) bool {
	if runtime == nil || !runtime.Enabled || runtime.Ranking == nil {
		ctx.SendChain(message.Text(fmt.Sprintf("%s服务未配置", command)))
		return false
	}
	return true
}

func fetchRankingBoard(client *ranking.Client, worldLink bool, rawArgs string) (*ranking.Board, error) {
	if client == nil {
		return nil, fmt.Errorf("ranking client is nil")
	}
	if !worldLink {
		return client.GetLatest()
	}
	wl, err := client.GetWorldLinkLatest()
	if err != nil {
		return nil, err
	}
	group, ok := selectWorldLinkGroup(wl.Groups, rawArgs)
	if !ok {
		return nil, ranking.ErrNoWorldLinkData
	}
	board := group.Board()
	if board.EventID == 0 {
		board.EventID = wl.EventID
	}
	if board.Region == "" {
		board.Region = wl.Region
	}
	if board.UpdatedAt == 0 {
		board.UpdatedAt = wl.UpdatedAt
	}
	return &board, nil
}

func fetchChurnBoard(client *ranking.Client, worldLink bool, rawArgs string) (*ranking.Board, error) {
	if client == nil {
		return nil, fmt.Errorf("ranking client is nil")
	}
	if !worldLink {
		return client.GetChurn()
	}
	wl, err := client.GetWorldLinkLatest()
	if err != nil {
		return nil, err
	}
	group, ok := selectWorldLinkGroup(wl.Groups, rawArgs)
	if !ok {
		return nil, ranking.ErrNoWorldLinkData
	}
	return client.GetWorldLinkChurn(group.GameCharacterID)
}

func hydrateChurnBoardAvatars(client *ranking.Client, churn *ranking.Board, worldLink bool, rawArgs string) {
	if client == nil || churn == nil || len(churn.Rankings) == 0 {
		return
	}
	latest, err := fetchRankingBoard(client, worldLink, rawArgs)
	if err != nil || latest == nil || len(latest.Rankings) == 0 {
		return
	}
	hydrateRankingEntriesFromLatest(churn.Rankings, latest.Rankings)
}

func hydrateRankingEntriesFromLatest(churn []ranking.RankingEntry, latest []ranking.RankingEntry) {
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

func selectWorldLinkGroup(groups []ranking.WorldLinkGroup, rawArgs string) (ranking.WorldLinkGroup, bool) {
	if len(groups) == 0 {
		return ranking.WorldLinkGroup{}, false
	}
	selector := firstNumericToken(rawArgs)
	if selector > 0 {
		for _, group := range groups {
			if group.GameCharacterID == selector {
				return group, true
			}
		}
		if selector <= len(groups) {
			return groups[selector-1], true
		}
	}
	best := groups[0]
	for _, group := range groups[1:] {
		if group.UpdatedAt > best.UpdatedAt {
			best = group
		}
	}
	return best, true
}

func firstNumericToken(raw string) int {
	tokens := strings.Fields(raw)
	// 只有当至少有两个 token 时，首个数字 token 才作为 WorldLink 角色组选择符；
	// 单个 token 视为排名/UID，角色组使用默认值。
	if len(tokens) < 2 {
		return 0
	}
	for _, token := range tokens {
		value, ok := parseRankToken(token)
		if ok && value > 0 && value <= 26 {
			return value
		}
	}
	return 0
}

func parseRankingQuery(raw string, inferredUser *models.User, deps *Deps, ctx *zero.Ctx, region string, worldLink bool, single bool) (rankingQuery, error) {
	args := strings.TrimSpace(raw)
	query := rankingQuery{}
	if args == "" {
		user, gameID := boundGameID(deps, ctx, region, inferredUser)
		_ = user
		if gameID == "" {
			return query, fmt.Errorf("请提供排名/UID，或先使用 /%s绑定 [游戏ID] 绑定账号", region)
		}
		query.UserID = gameID
		return query, nil
	}
	tokens := strings.Fields(args)
	for index, token := range tokens {
		// 仅在 token 数 >= 2 时，首个 selector token 才被视为 WorldLink 角色组选择符
		if worldLink && index == 0 && len(tokens) >= 2 && isWorldLinkSelectorToken(token) {
			continue
		}
		if strings.Contains(token, "-") {
			items, ok := parseRankRange(token)
			if ok {
				query.Ranks = append(query.Ranks, items...)
				continue
			}
		}
		if rank, ok := parseRankToken(token); ok {
			query.Ranks = append(query.Ranks, rank)
			continue
		}
		if looksLikeUID(token) {
			query.UserID = strings.TrimSpace(token)
		}
	}
	query.Ranks = dedupeInts(query.Ranks)
	if single && len(query.Ranks) > 1 {
		return query, fmt.Errorf("查水表不支持同时查询多个玩家")
	}
	if query.UserID == "" && len(query.Ranks) == 0 {
		return query, fmt.Errorf("没有识别到排名或 UID")
	}
	return query, nil
}

func boundGameID(deps *Deps, ctx *zero.Ctx, region string, inferred *models.User) (*models.User, string) {
	if inferred != nil && inferred.GameID != "" && (region == "" || inferred.ServerRegion == region) {
		return inferred, inferred.GameID
	}
	if deps == nil || deps.DB == nil {
		return nil, ""
	}
	user, err := deps.DB.GetUserByPlatformRegion("onebot", userIDFromCtx(ctx), region)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, ""
	}
	if user == nil {
		return nil, ""
	}
	return user, user.GameID
}

func isWorldLinkSelectorToken(token string) bool {
	value, ok := parseRankToken(token)
	return ok && value > 0 && value <= 26
}

func parseRankRange(token string) ([]int, bool) {
	parts := strings.Split(token, "-")
	if len(parts) != 2 {
		return nil, false
	}
	start, ok1 := parseRankToken(parts[0])
	end, ok2 := parseRankToken(parts[1])
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

func parseRankToken(token string) (int, bool) {
	value := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(token, "#")))
	if value == "" {
		return 0, false
	}
	multiplier := 1
	switch {
	case strings.HasSuffix(value, "w") || strings.HasSuffix(value, "万"):
		multiplier = 10000
		value = strings.TrimSuffix(strings.TrimSuffix(value, "w"), "万")
	case strings.HasSuffix(value, "k") || strings.HasSuffix(value, "千"):
		multiplier = 1000
		value = strings.TrimSuffix(strings.TrimSuffix(value, "k"), "千")
	}
	number, err := strconv.Atoi(value)
	if err != nil || number <= 0 {
		return 0, false
	}
	return number * multiplier, true
}

func looksLikeUID(token string) bool {
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

func dedupeInts(values []int) []int {
	seen := map[int]struct{}{}
	out := make([]int, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func defaultBorderRanks() []int {
	return []int{10, 20, 30, 40, 50, 100, 200, 300, 400, 500, 1000, 2000, 3000, 4000, 5000, 10000, 20000, 30000, 40000, 50000, 100000}
}

func selectRankingEntries(entries []ranking.RankingEntry, query rankingQuery) []ranking.RankingEntry {
	if query.UserID != "" {
		for _, entry := range entries {
			if entry.UserID.String() == query.UserID {
				return []ranking.RankingEntry{entry}
			}
		}
	}
	if query.UserID != "" && len(query.Ranks) > 0 {
		// 若绑定用户不在当前 churn 响应（例如不在 TOP100），仍显示默认 TOP 档位查房数据。
		// 若在 TOP100，则上面的 UID 命中会优先返回该用户。
	}
	if len(query.Ranks) == 0 {
		return nil
	}
	return filterRankingsByRanks(entries, query.Ranks)
}

func filterRankingsByRanks(entries []ranking.RankingEntry, ranks []int) []ranking.RankingEntry {
	out := make([]ranking.RankingEntry, 0, len(ranks))
	for _, rank := range ranks {
		if rank <= 0 || len(entries) == 0 {
			continue
		}
		out = append(out, nearestRanking(entries, rank))
	}
	return out
}

func topRankings(entries []ranking.RankingEntry, limit int) []ranking.RankingEntry {
	if limit <= 0 || len(entries) == 0 {
		return nil
	}
	if len(entries) < limit {
		limit = len(entries)
	}
	return append([]ranking.RankingEntry(nil), entries[:limit]...)
}

func nearestRanking(entries []ranking.RankingEntry, rank int) ranking.RankingEntry {
	best := entries[0]
	bestDiff := absInt(entries[0].Rank - rank)
	for _, entry := range entries[1:] {
		diff := absInt(entry.Rank - rank)
		if diff < bestDiff {
			best = entry
			bestDiff = diff
		}
	}
	return best
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func fetchForecast(client *ranking.Client, rawArgs string) (ranking.ForecastEvent, *ranking.ForecastBoard, error) {
	events, err := client.GetForecastEvents()
	if err != nil {
		return ranking.ForecastEvent{}, nil, err
	}
	eventID := 0
	if rank, ok := parseRankToken(strings.TrimSpace(rawArgs)); ok {
		eventID = rank
	}
	var selected ranking.ForecastEvent
	if eventID > 0 {
		for _, event := range events {
			if event.EventID == eventID {
				selected = event
				break
			}
		}
		if selected.EventID == 0 {
			selected = ranking.ForecastEvent{EventID: eventID, Name: fmt.Sprintf("Event %d", eventID)}
		}
	} else {
		selected = selectForecastEvent(events)
	}
	if selected.EventID == 0 {
		return ranking.ForecastEvent{}, nil, fmt.Errorf("没有可用活动")
	}
	board, err := client.GetForecastLatest(selected.EventID)
	if err != nil {
		return selected, nil, err
	}
	return selected, board, nil
}

func selectForecastEvent(events []ranking.ForecastEvent) ranking.ForecastEvent {
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

func rankingTitle(base string, worldLink bool, targetID int) string {
	if worldLink {
		if targetID > 0 {
			return fmt.Sprintf("WL %s · 角色 %d", base, targetID)
		}
		return "WL " + base
	}
	return base
}

func rankingSubtitle(region string, board ranking.Board) string {
	parts := []string{fmt.Sprintf("%s · Event #%d", regionLabel(region), board.EventID)}
	if board.BoardType == "worldlink" && board.TargetID > 0 {
		parts = append(parts, fmt.Sprintf("角色 %d", board.TargetID))
	}
	if board.UpdatedAt > 0 {
		parts = append(parts, "更新 "+formatTimestamp(board.UpdatedAt))
	}
	return strings.Join(parts, " · ")
}

func rankingErrorText(err error, worldLink bool, action string) string {
	if errors.Is(err, ranking.ErrNoWorldLinkData) {
		return "当前没有可用 WL 单人榜数据"
	}
	if worldLink {
		return fmt.Sprintf("WL %s获取失败，请稍后重试", action)
	}
	return fmt.Sprintf("%s获取失败，请稍后重试", action)
}

func sendRankingImage(ctx *zero.Ctx, client *renderer.Client, payload renderpayloads.RankingListPayload) bool {
	return sendRankingImageWithTemplate(ctx, client, "ranking_list", payload)
}

func sendRankingImageWithTemplate(ctx *zero.Ctx, client *renderer.Client, template string, payload renderpayloads.RankingListPayload) bool {
	return sendGenericImage(ctx, client, template, payload)
}

func sendGenericImage(ctx *zero.Ctx, client *renderer.Client, template string, payload any) bool {
	if client == nil || !client.Health() {
		return false
	}
	png, err := client.Render(renderer.RenderRequest{Template: template, Data: payload})
	if err != nil {
		return false
	}
	ctx.SendChain(message.ImageBytes(png))
	return true
}

func formatRankingText(payload renderpayloads.RankingListPayload) string {
	lines := []string{payload.Title}
	if payload.Subtitle != "" {
		lines = append(lines, payload.Subtitle)
	}
	for _, entry := range payload.Rankings {
		name := entry.Name
		if name == "" {
			name = entry.DisplayName
		}
		if name == "" {
			name = "Unknown"
		}
		line := fmt.Sprintf("#%d %s %sP", entry.Rank, name, formatInt64(entry.Score))
		if entry.ScoreDelta != 0 {
			line += fmt.Sprintf(" Δ%s", formatSignedInt64(entry.ScoreDelta))
		}
		if entry.UserID != "" {
			line += " · UID " + entry.UserID
		}
		if entry.Signature != "" {
			line += " · " + entry.Signature
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func formatWaterTableText(payload renderpayloads.WaterTablePayload) string {
	entry := payload.Entry
	name := firstNonEmptyString(entry.Name, entry.DisplayName, "Unknown")
	lines := []string{payload.Title}
	if payload.Subtitle != "" {
		lines = append(lines, payload.Subtitle)
	}
	lines = append(lines, fmt.Sprintf("#%d %s %sP", entry.Rank, name, formatInt64(entry.Score)))
	lines = append(lines, fmt.Sprintf("48H周回：%d，1H增长：%s", entry.Churn48h, formatInt64(entry.Growth1h)))
	if len(payload.HourlyChurn) > 0 {
		last := payload.HourlyChurn[len(payload.HourlyChurn)-1]
		lines = append(lines, fmt.Sprintf("最近小时：%s · %d 次", last.Hour, last.Count))
	}
	if len(payload.Parking) > 0 {
		period := payload.Parking[len(payload.Parking)-1]
		lines = append(lines, fmt.Sprintf("最近停车：%s", formatDurationSeconds(period.DurationS)))
	}
	return strings.Join(lines, "\n")
}

func formatForecastText(payload renderpayloads.ForecastRankingPayload) string {
	lines := []string{payload.Title}
	if payload.Subtitle != "" {
		lines = append(lines, payload.Subtitle)
	}
	if payload.UpdatedAt > 0 {
		lines = append(lines, "更新 "+formatTimestamp(payload.UpdatedAt))
	}
	for _, item := range payload.Items {
		line := fmt.Sprintf("#%d 当前 %sP", item.Rank, formatInt64(item.Score))
		if item.IsFinal {
			line += " · 最终线"
		} else if item.HasPrediction {
			line += fmt.Sprintf(" · 预测 %sP", formatInt64(item.Prediction))
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func formatInt64(value int64) string {
	return formatNumber(int(value))
}

func formatSignedInt64(value int64) string {
	if value > 0 {
		return "+" + formatInt64(value)
	}
	return formatInt64(value)
}

func formatTimestamp(value int64) string {
	if value <= 0 {
		return "-"
	}
	if value < 1_000_000_000_000 {
		value *= 1000
	}
	return time.UnixMilli(value).Format("01-02 15:04")
}

func formatDurationSeconds(value int64) string {
	if value <= 0 {
		return "进行中"
	}
	d := time.Duration(value) * time.Second
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
