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
	"gorm.io/gorm"
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
}

func cardBoxFields() []string {
	return suite.Fields()
}

func RegisterSuiteCardBox(deps *Deps) {
	for _, cmd := range parserCommands(deps, "卡牌一览") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		zero.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, user := runtimeForCommand(deps, ctx, forcedRegion)
			if runtime == nil || !runtime.Enabled || runtime.Store == nil {
				ctx.SendChain(message.Text(runtimeUnavailableText(runtime)))
				return
			}
			if forcedRegion != "" {
				var err error
				user, err = deps.DB.GetUserByPlatformRegion("onebot", userIDFromCtx(ctx), runtime.Region)
				if err != nil && err != gorm.ErrRecordNotFound {
					ctx.SendChain(message.Text("数据库错误，请稍后重试"))
					return
				}
			}
			if user == nil || user.GameID == "" {
				ctx.SendChain(message.Text(fmt.Sprintf("你还没有绑定%s游戏账号~\n使用 /%s绑定 [游戏ID] 来绑定", runtime.Label, runtime.Region)))
				return
			}
			if runtime.Suite == nil || !runtime.Suite.Enabled() {
				ctx.SendChain(message.Text(fmt.Sprintf("暂不支持查询%s的抓包数据", runtime.Label)))
				return
			}
			setting := suiteSettingOrDefault(deps, userIDFromCtx(ctx), runtime.Region)
			if setting.Hidden {
				ctx.SendChain(message.Text(fmt.Sprintf("你已隐藏%s抓包信息，发送 /%s展示抓包 可重新展示", runtime.Label, runtime.Region)))
				return
			}

			options := parseCardBoxOptions(commandArgs(ctx))
			cards, msg := cardBoxCards(runtime.Store, options)
			if msg != "" {
				ctx.SendChain(message.Text(msg))
				return
			}

			var profile cardBoxProfile
			if err := runtime.Suite.GetUserData(user.GameID, "", cardBoxFields(), &profile); err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("获取你的%s Haruki Suite 公开数据失败\n%s", runtime.Label, err.Error())))
				return
			}
			owned := renderer.SuiteUserCardMap(profile.UserCards)
			deckSet := cardBoxDeckSet(profile)
			payload := renderer.BuildSuiteCardBoxPayload(
				suitePanelTitle(runtime, "卡牌一览"),
				cardBoxSubtitle(options, len(cards), len(owned)),
				runtime.Region,
				"",
				profile.BaseProfile,
				profile.UserGamedata,
				cards,
				owned,
				deckSet,
				runtime.Store,
				runtime.Assets,
				renderer.SuiteCardBoxOptions{ShowID: options.ShowID, OwnedOnly: options.OwnedOnly, UseBeforeTraining: options.UseBeforeTraining, ShowCreatedAt: options.ShowCreatedAt, SortBy: options.SortBy},
			)
			sendSuiteCardBoxOrText(ctx, deps, payload, formatCardBoxText(runtime.Region, profile, cards, owned, options))
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
			remaining = append(remaining, token)
		}
	}
	options.FilterText = strings.TrimSpace(strings.Join(remaining, " "))
	return options
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
		cards := append([]masterdata.CardInfo(nil), result.Cards...)
		sortCardBoxMasterCards(cards)
		return cards, ""
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

func cardBoxSubtitle(options cardBoxQueryOptions, total int, owned int) string {
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
	return strings.Join(parts, " · ")
}

func formatCardBoxText(region string, profile cardBoxProfile, cards []masterdata.CardInfo, owned map[int]suite.UserCard, options cardBoxQueryOptions) string {
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
