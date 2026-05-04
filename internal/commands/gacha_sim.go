package commands

import (
	"math/rand"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/renderer"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type gachaResultPayload struct {
	PullType    string                  `json:"pullType"`
	AssetSource string                  `json:"assetSource,omitempty"`
	Results     []gachaResultCardSample `json:"results"`
}

type gachaResultCardSample struct {
	CardID              int    `json:"cardId"`
	CharacterName       string `json:"characterName"`
	Rarity              string `json:"rarity"`
	Attr                string `json:"attr"`
	AssetbundleName     string `json:"assetbundleName,omitempty"`
	ThumbnailURL        string `json:"thumbnailUrl,omitempty"`
	TrainedThumbnailURL string `json:"trainedThumbnailUrl,omitempty"`
	IsTrained           bool   `json:"isTrained,omitempty"`
	IsNew               bool   `json:"isNew,omitempty"`
}

// RegisterGachaSimulation registers /抽卡模拟 and its aliases.
func RegisterGachaSimulation(deps *Deps) {
	for _, cmd := range parserCommands(deps, "抽卡模拟") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		zero.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, _ := runtimeForCommand(deps, ctx, forcedRegion)
			if runtime == nil || runtime.Store == nil || !runtime.Enabled {
				ctx.SendChain(message.Text(runtimeUnavailableText(runtime)))
				return
			}

			payload := buildGachaSimulationPayload(runtime.Store, runtime.Assets)
			if deps.Renderer != nil && deps.Renderer.Health() {
				png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "gacha_result", Data: payload})
				if err == nil {
					ctx.SendChain(message.ImageBytes(png))
					bot.RecordCommandRegion(deps.DB, "抽卡模拟", runtime.Region, ctx, start)
					return
				}
			}
			ctx.SendChain(message.Text("抽卡模拟结果已生成，但图片渲染失败，请确认 Renderer 可用。"))
			bot.RecordCommandRegion(deps.DB, "抽卡模拟", runtime.Region, ctx, start)
		})
	}
}

func buildGachaSimulationPayload(store *masterdata.Store, resolver interface {
	RendererAssetSource() string
	GetCardThumbnailURL(string, bool) string
}) gachaResultPayload {
	cards := store.AllCards()
	if len(cards) == 0 {
		return fallbackGachaSimulationPayload()
	}
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	results := make([]gachaResultCardSample, 0, 10)
	for i := 0; i < 10; i++ {
		card := cards[seed.Intn(len(cards))]
		if i == 9 {
			card = preferHighRarity(cards, seed)
		}
		item := gachaResultCardSample{
			CardID:          card.ID,
			CharacterName:   characterDisplayName(card.CharacterID),
			Rarity:          card.CardRarityType,
			Attr:            card.Attr,
			AssetbundleName: card.AssetbundleName,
			IsTrained:       card.CardRarityType == "rarity_3" || card.CardRarityType == "rarity_4" || card.CardRarityType == "rarity_birthday",
			IsNew:           i == 2 || i == 8,
		}
		if resolver != nil && card.AssetbundleName != "" {
			item.ThumbnailURL = resolver.GetCardThumbnailURL(card.AssetbundleName, false)
			item.TrainedThumbnailURL = resolver.GetCardThumbnailURL(card.AssetbundleName, true)
		}
		results = append(results, item)
	}
	assetSource := ""
	if resolver != nil {
		assetSource = resolver.RendererAssetSource()
	}
	return gachaResultPayload{PullType: "multi", AssetSource: assetSource, Results: results}
}

func preferHighRarity(cards []masterdata.CardInfo, seed *rand.Rand) masterdata.CardInfo {
	candidates := make([]masterdata.CardInfo, 0)
	for _, card := range cards {
		if card.CardRarityType == "rarity_3" || card.CardRarityType == "rarity_4" || card.CardRarityType == "rarity_birthday" {
			candidates = append(candidates, card)
		}
	}
	if len(candidates) == 0 {
		return cards[seed.Intn(len(cards))]
	}
	return candidates[seed.Intn(len(candidates))]
}

func characterDisplayName(characterID int) string {
	names := map[int]string{
		1: "一歌", 2: "咲希", 3: "穗波", 4: "志步", 5: "实乃理", 6: "遥", 7: "爱莉", 8: "雫",
		9: "心羽", 10: "杏", 11: "彰人", 12: "冬弥", 13: "司", 14: "笑梦", 15: "宁宁", 16: "类",
		17: "奏", 18: "真冬", 19: "绘名", 20: "瑞希", 21: "初音未来", 22: "镜音铃", 23: "镜音连", 24: "巡音流歌", 25: "MEIKO", 26: "KAITO",
	}
	if name := names[characterID]; name != "" {
		return name
	}
	return "角色"
}

func fallbackGachaSimulationPayload() gachaResultPayload {
	return gachaResultPayload{
		PullType:    "multi",
		AssetSource: "main-jp",
		Results: []gachaResultCardSample{
			{CardID: 3001, CharacterName: "初音未来", Rarity: "rarity_4", Attr: "cute", AssetbundleName: "res001_no003", IsTrained: true, IsNew: true},
			{CardID: 3002, CharacterName: "镜音铃", Rarity: "rarity_3", Attr: "happy", AssetbundleName: "res002_no003", IsTrained: true},
			{CardID: 3003, CharacterName: "镜音连", Rarity: "rarity_2", Attr: "cool", AssetbundleName: "res003_no003"},
		},
	}
}
