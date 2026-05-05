package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/renderer"
	"moebot-next/internal/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

const materialDefaultLimit = 20

type materialProfile struct {
	suite.BaseProfile
	UserGamedata  suite.UserGamedata `json:"userGamedata"`
	UserDecks     []suite.UserDeck   `json:"userDecks"`
	UserCards     []suite.UserCard   `json:"userCards"`
	UserMaterials []userMaterial     `json:"userMaterials"`
}

type userMaterial struct {
	MaterialID int   `json:"materialId"`
	Quantity   int64 `json:"quantity"`
}

func materialFields() []string {
	return suite.Fields(suite.FieldUserMaterials)
}

func RegisterMaterial(deps *Deps) {
	for _, cmd := range parserCommands(deps, "材料信息") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		zero.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, user := runtimeForCommand(deps, ctx, forcedRegion)
			if runtime == nil || !runtime.Enabled {
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
			var profile materialProfile
			if err := runtime.Suite.GetUserData(user.GameID, "", materialFields(), &profile); err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("获取你的%s Haruki Suite 公开数据失败\n%s", runtime.Label, err.Error())))
				return
			}
			payload := buildSuitePanel(runtime, suitePanelTitle(runtime, "材料信息"), "", profile)
			payload.Subtitle = suitePanelSubtitle(profile.BaseProfile)
			rows, stats := rowsFromMaterials(profile, materialDefaultLimit)
			payload.Stats = append(suiteBasicStats(profile.commonSuiteProfile()), stats...)
			payload.Sections = []renderer.SuiteSectionPayload{{Title: "持有材料", Rows: rows}}
			sendSuitePanelOrText(ctx, deps, payload, formatMaterialText(runtime.Region, profile, materialDefaultLimit))
			bot.RecordCommandRegion(deps.DB, "材料信息", runtime.Region, ctx, start)
		})
	}
}

func formatMaterialText(region string, profile materialProfile, limit int) string {
	name := profile.UserGamedata.Name
	if name == "" {
		name = "未知玩家"
	}
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
	if limit <= 0 || limit > len(materials) {
		limit = len(materials)
	}
	lines := []string{
		fmt.Sprintf("%s 材料信息", strings.ToUpper(config.NormalizeRegion(region))),
		fmt.Sprintf("玩家: %s", name),
		fmt.Sprintf("金币: %d", profile.UserGamedata.Coin),
		fmt.Sprintf("更新时间: %s", suiteUpdateText(profile.UploadTime)),
		fmt.Sprintf("数据来源: %s", suiteSourceText(profile.BaseProfile)),
	}
	if len(materials) == 0 {
		lines = append(lines, "暂无材料数据")
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "---")
	for i := 0; i < limit; i++ {
		material := materials[i]
		lines = append(lines, fmt.Sprintf("%d. 材料 #%d: %d", i+1, material.MaterialID, material.Quantity))
	}
	return strings.Join(lines, "\n")
}
