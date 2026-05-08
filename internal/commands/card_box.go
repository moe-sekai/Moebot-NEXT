package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/cardquery"
	"moebot-next/internal/config"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/renderer"
	"moebot-next/internal/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type cardBoxProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata `json:"userGamedata"`
	UserDecks    []suite.UserDeck   `json:"userDecks"`
	UserCards    []suite.UserCard   `json:"userCards"`
}

type cardBoxQueryOptions struct {
	ShowID            bool
	OwnedOnly         bool
	UseBeforeTraining bool
	ShowCreatedAt     bool
	SortBy            string
	FilterText        string
	Page              int
}

const cardBoxPageSize = 100

func cardBoxFields() []string {
	return suite.Fields()
}

func RegisterSuiteCardBox(deps *Deps) {
	for _, cmd := range parserCommands(deps, "卡牌一览") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		zero.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, inferredUser, ok := requireRuntimeWithStore(deps, ctx, forcedRegion)
			if !ok {
				return
			}
			user, ok := requireBoundUser(deps, ctx, runtime, forcedRegion, inferredUser)
			if !ok {
				return
			}
			if !requireSuite(ctx, runtime, "卡牌一览") {
				return
			}
			if _, ok := requireSuiteVisible(deps, ctx, runtime); !ok {
				return
			}

			options := parseCardBoxOptions(commandArgs(ctx))
			cards, msg := cardBoxCards(runtime.Store, options)
			if msg != "" {
				ctx.SendChain(message.Text(msg))
				return
			}

			var profile cardBoxProfile
			if !fetchSuiteUserData(ctx, runtime, user.GameID, "卡牌一览", cardBoxFields(), &profile) {
				return
			}
			owned := renderer.SuiteUserCardMap(profile.UserCards)
			deckSet := cardBoxDeckSet(profile)
			pagedCards, page, totalPages := paginateCardBox(cards, options, owned)
			payload := renderer.BuildSuiteCardBoxPayload(
				suitePanelTitle(runtime, "卡牌一览"),
				cardBoxSubtitle(options, len(cards), len(owned), page, totalPages),
				runtime.Region,
				"",
				profile.BaseProfile,
				profile.UserGamedata,
				pagedCards,
				owned,
				deckSet,
				runtime.Store,
				runtime.Assets,
				renderer.SuiteCardBoxOptions{ShowID: options.ShowID, OwnedOnly: options.OwnedOnly, UseBeforeTraining: options.UseBeforeTraining, ShowCreatedAt: options.ShowCreatedAt, SortBy: options.SortBy},
			)
			payload.Page = page
			payload.TotalPages = totalPages
			payload.PageSize = cardBoxPageSize
			payload.TotalAll = len(cards)
			sendSuiteCardBoxOrText(ctx, deps, payload, formatCardBoxText(runtime.Region, profile, cards, owned, options, page, totalPages))
			bot.RecordCommandRegion(deps.DB, "卡牌一览", runtime.Region, ctx, start)
		})
	}
}

func parseCardBoxOptions(raw string) cardBoxQueryOptions {
	options := cardBoxQueryOptions{FilterText: strings.TrimSpace(raw)}
	tokens := strings.Fields(raw)
	remaining := make([]string, 0, len(tokens))
	for _, token := range tokens {
		lower := strings.ToLower(strings.TrimSpace(token))
		switch lower {
		case "box", "owned", "持有", "已持有":
			options.OwnedOnly = true
		case "id", "ids", "编号", "显示id":
			options.ShowID = true
			options.SortBy = "id"
		case "before", "normal", "花前", "特训前":
			options.UseBeforeTraining = true
		case "time", "created", "createdat", "获取时间", "入手时间", "时间排序":
			options.SortBy = "time"
			options.ShowCreatedAt = true
		case "mr", "master", "masterrank", "专精", "专精排序", "rank":
			options.SortBy = "mr"
		case "sl", "skill", "skilllevel", "技能等级", "技能等级排序":
			options.SortBy = "sl"
		default:
			if page, ok := parseCardBoxPageToken(lower); ok {
				options.Page = page
				continue
			}
			remaining = append(remaining, token)
		}
	}
	options.FilterText = strings.TrimSpace(strings.Join(remaining, " "))
	return options
}

func parseCardBoxPageToken(token string) (int, bool) {
	if token == "" {
		return 0, false
	}
	switch {
	case strings.HasPrefix(token, "@"):
		return parsePositivePage(strings.TrimPrefix(token, "@"))
	case strings.HasPrefix(token, "p"):
		return parsePositivePage(strings.TrimPrefix(token, "p"))
	case strings.HasSuffix(token, "页"):
		return parsePositivePage(strings.TrimSuffix(token, "页"))
	}
	return 0, false
}

func parsePositivePage(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	n := 0
	for _, r := range value {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
		if n > 100000 {
			return 0, false
		}
	}
	if n <= 0 {
		return 0, false
	}
	return n, true
}

func cardBoxCards(store *masterdata.Store, options cardBoxQueryOptions) ([]masterdata.CardInfo, string) {
	if store == nil {
		return nil, "卡牌数据不可用"
	}
	filter := strings.TrimSpace(options.FilterText)
	if filter != "" {
		result := cardquery.ResolveAll(store, filter)
		if result.Message != "" {
			return nil, result.Message
		}
		return append([]masterdata.CardInfo(nil), result.Cards...), ""
	}
	cards := store.AllCards()
	sortCardBoxMasterCards(cards)
	return cards, ""
}

func sortCardBoxMasterCards(cards []masterdata.CardInfo) {
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].CharacterID != cards[j].CharacterID {
			return cards[i].CharacterID < cards[j].CharacterID
		}
		if cards[i].ReleaseAt != cards[j].ReleaseAt {
			return cards[i].ReleaseAt < cards[j].ReleaseAt
		}
		return cards[i].ID < cards[j].ID
	})
}

func cardBoxDeckSet(profile cardBoxProfile) map[int]struct{} {
	common := suiteCommandProfile{BaseProfile: profile.BaseProfile, UserGamedata: profile.UserGamedata, UserDecks: profile.UserDecks, UserCards: profile.UserCards}
	deckCards := renderer.BuildSuiteDeckCards(common.UserDecks, common.UserCards, common.UserGamedata.Deck, nil, nil)
	out := make(map[int]struct{}, len(deckCards))
	for _, card := range deckCards {
		if card.CardID > 0 {
			out[card.CardID] = struct{}{}
		}
	}
	return out
}

func cardBoxSubtitle(options cardBoxQueryOptions, total int, owned int, page int, totalPages int) string {
	parts := []string{fmt.Sprintf("筛选 %d 张", total), fmt.Sprintf("已持有 %d 张", owned)}
	if options.FilterText != "" {
		parts = append(parts, "条件: "+options.FilterText)
	}
	if options.OwnedOnly {
		parts = append(parts, "仅显示持有")
	}
	if options.SortBy != "" {
		parts = append(parts, "排序: "+options.SortBy)
	}
	if totalPages > 1 {
		parts = append(parts, fmt.Sprintf("第 %d/%d 页", page, totalPages))
	}
	return strings.Join(parts, " · ")
}

func paginateCardBox(cards []masterdata.CardInfo, options cardBoxQueryOptions, owned map[int]suite.UserCard) ([]masterdata.CardInfo, int, int) {
	filtered := cards
	if options.OwnedOnly {
		filtered = make([]masterdata.CardInfo, 0, len(cards))
		for _, card := range cards {
			if _, ok := owned[card.ID]; ok {
				filtered = append(filtered, card)
			}
		}
	}
	total := len(filtered)
	totalPages := 1
	if total > 0 {
		totalPages = (total + cardBoxPageSize - 1) / cardBoxPageSize
	}
	page := options.Page
	if page <= 0 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}
	start := (page - 1) * cardBoxPageSize
	if start < 0 {
		start = 0
	}
	if start > total {
		start = total
	}
	end := start + cardBoxPageSize
	if end > total {
		end = total
	}
	if options.OwnedOnly {
		return filtered[start:end], page, totalPages
	}
	// When OwnedOnly is false we paginate the unfiltered list directly to keep
	// rendering predictable; the renderer will hide untyped cards via OwnedOnly elsewhere.
	if start > len(cards) {
		start = len(cards)
	}
	if end > len(cards) {
		end = len(cards)
	}
	return cards[start:end], page, totalPages
}

func formatCardBoxText(region string, profile cardBoxProfile, cards []masterdata.CardInfo, owned map[int]suite.UserCard, options cardBoxQueryOptions, page int, totalPages int) string {
	cards = sortedCardBoxTextCards(cards, owned, options)
	name := profile.UserGamedata.Name
	if name == "" {
		name = "未知玩家"
	}
	shown := 0
	for _, card := range cards {
		_, has := owned[card.ID]
		if options.OwnedOnly && !has {
			continue
		}
		shown++
	}
	lines := []string{
		fmt.Sprintf("%s 卡牌一览", strings.ToUpper(config.NormalizeRegion(region))),
		fmt.Sprintf("玩家: %s", name),
		fmt.Sprintf("显示卡牌: %d", shown),
		fmt.Sprintf("已持有: %d", len(owned)),
		fmt.Sprintf("更新时间: %s", suiteUpdateText(profile.UploadTime)),
		fmt.Sprintf("数据来源: %s", suiteSourceText(profile.BaseProfile)),
	}
	if totalPages > 1 {
		lines = append(lines, fmt.Sprintf("第 %d/%d 页（输入 @页码 切换，例如 @2）", page, totalPages))
	}
	limit := 30
	count := 0
	for _, card := range cards {
		userCard, has := owned[card.ID]
		if options.OwnedOnly && !has {
			continue
		}
		if count == 0 {
			lines = append(lines, "---")
		}
		if count >= limit {
			lines = append(lines, fmt.Sprintf("还有 %d 张未显示，请使用图片渲染查看完整列表", shown-limit))
			break
		}
		status := "未持有"
		if has {
			status = fmt.Sprintf("Lv.%d MR%d SL%d", userCard.Level, userCard.MasterRank, userCard.SkillLevel)
		}
		lines = append(lines, fmt.Sprintf("#%d %s %s · %s", card.ID, characterDisplayName(card.CharacterID), card.Prefix, status))
		count++
	}
	if shown == 0 {
		lines = append(lines, "暂无符合条件的卡牌")
	}
	return strings.Join(lines, "\n")
}

func sortedCardBoxTextCards(cards []masterdata.CardInfo, owned map[int]suite.UserCard, options cardBoxQueryOptions) []masterdata.CardInfo {
	out := append([]masterdata.CardInfo(nil), cards...)
	if options.SortBy == "" {
		return out
	}
	sort.SliceStable(out, func(i, j int) bool {
		left, leftOwned := owned[out[i].ID]
		right, rightOwned := owned[out[j].ID]
		switch options.SortBy {
		case "mr":
			if left.MasterRank != right.MasterRank {
				return left.MasterRank > right.MasterRank
			}
			if leftOwned != rightOwned {
				return leftOwned
			}
		case "sl":
			if left.SkillLevel != right.SkillLevel {
				return left.SkillLevel > right.SkillLevel
			}
			if leftOwned != rightOwned {
				return leftOwned
			}
		case "time":
			if left.CreatedAt != right.CreatedAt {
				return left.CreatedAt > right.CreatedAt
			}
			if leftOwned != rightOwned {
				return leftOwned
			}
		case "id":
			if out[i].ID != out[j].ID {
				return out[i].ID < out[j].ID
			}
		}
		if out[i].CharacterID != out[j].CharacterID {
			return out[i].CharacterID < out[j].CharacterID
		}
		if out[i].ReleaseAt != out[j].ReleaseAt {
			return out[i].ReleaseAt < out[j].ReleaseAt
		}
		return out[i].ID < out[j].ID
	})
	return out
}
