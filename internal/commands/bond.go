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
)

const bondDefaultLimit = 10

type bondProfile struct {
	suite.BaseProfile
	UserGamedata suite.UserGamedata `json:"userGamedata"`
	UserDecks    []suite.UserDeck   `json:"userDecks"`
	UserCards    []suite.UserCard   `json:"userCards"`
	UserBonds    []userBond         `json:"userBonds"`
}

type userBond struct {
	BondsGroupID     int `json:"bondsGroupId"`
	CharacterID1     int `json:"characterId1"`
	CharacterID2     int `json:"characterId2"`
	GameCharacterID1 int `json:"gameCharacterId1"`
	GameCharacterID2 int `json:"gameCharacterId2"`
	Rank             int `json:"rank"`
	Exp              int `json:"exp"`
}

func bondFields() []string {
	return suite.Fields(suite.FieldUserBonds)
}

func RegisterBond(deps *Deps) {
	for _, cmd := range parserCommands(deps, "羁绊") {
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
			if !requireSuite(ctx, runtime, "羁绊") {
				return
			}
			if _, ok := requireSuiteVisible(deps, ctx, runtime); !ok {
				return
			}

			var profile bondProfile
			if !fetchSuiteUserData(ctx, runtime, user.GameID, "羁绊", bondFields(), &profile) {
				return
			}
			payload := buildSuitePanel(runtime, suitePanelTitle(runtime, "羁绊查询"), "", profile)
			payload.Subtitle = suitePanelSubtitle(profile.BaseProfile)
			rows, stats := rowsFromBonds(profile, bondDefaultLimit)
			payload.Stats = append(suiteBasicStats(profile.commonSuiteProfile()), stats...)
			payload.Sections = []renderer.SuiteSectionPayload{{Title: "羁绊 TOP", Kind: "bond_list", Note: "角色头像来自本地 assets/characters。", Rows: rows}}
			sendSuitePanelOrText(ctx, deps, payload, formatBondText(runtime.Region, profile, bondDefaultLimit))
			bot.RecordCommandRegion(deps.DB, "羁绊", runtime.Region, ctx, start)
		})
	}
}

func formatBondText(region string, profile bondProfile, limit int) string {
	name := profile.UserGamedata.Name
	if name == "" {
		name = "未知玩家"
	}
	source := suiteSourceText(profile.BaseProfile)
	updateText := suiteUpdateText(profile.UploadTime)

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
	if limit <= 0 || limit > len(bonds) {
		limit = len(bonds)
	}

	lines := []string{
		fmt.Sprintf("%s 羁绊", strings.ToUpper(config.NormalizeRegion(region))),
		fmt.Sprintf("玩家: %s", name),
		fmt.Sprintf("更新时间: %s", updateText),
		fmt.Sprintf("数据来源: %s", source),
	}
	if len(bonds) == 0 {
		lines = append(lines, "暂无羁绊数据")
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "---")
	for i := 0; i < limit; i++ {
		bond := bonds[i]
		cid1, cid2 := bondCharacterIDs(bond)
		lines = append(lines, fmt.Sprintf("%d. %s × %s Lv.%d EXP %d", i+1, characterDisplayName(cid1), characterDisplayName(cid2), bond.Rank, bond.Exp))
	}
	return strings.Join(lines, "\n")
}

func bondCharacterIDs(bond userBond) (int, int) {
	cid1, cid2 := bond.CharacterID1, bond.CharacterID2
	if cid1 == 0 {
		cid1 = bond.GameCharacterID1
	}
	if cid2 == 0 {
		cid2 = bond.GameCharacterID2
	}
	if (cid1 == 0 || cid2 == 0) && bond.BondsGroupID > 0 {
		cid1 = bond.BondsGroupID / 100 % 100
		cid2 = bond.BondsGroupID % 100
	}
	return cid1, cid2
}
