package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"moebot-next/internal/plugins/moesekai/renderpayloads"
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
			runtime, inferredUser, ok := requireRuntime(deps, ctx, forcedRegion)
			if !ok {
				return
			}
			user, ok := requireBoundUser(deps, ctx, runtime, forcedRegion, inferredUser)
			if !ok {
				return
			}
			if !requireSuite(ctx, runtime, "材料信息") {
				return
			}
			if _, ok := requireSuiteVisible(deps, ctx, runtime); !ok {
				return
			}
			var profile materialProfile
			if !fetchSuiteUserData(ctx, runtime, user.GameID, "材料信息", materialFields(), &profile) {
				return
			}
			payload := buildSuitePanel(runtime, suitePanelTitle(runtime, "材料信息"), "", profile)
			payload.Subtitle = suitePanelSubtitle(profile.BaseProfile)
			rows, stats := rowsFromMaterials(profile, materialDefaultLimit)
			payload.Stats = append(suiteBasicStats(profile.commonSuiteProfile()), stats...)
			payload.Sections = []renderpayloads.SuiteSectionPayload{{Title: "持有材料", Rows: rows}}
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
