package commands

import (
	"time"

	"moebot-next/internal/bot"
	"moebot-next/internal/plugins/moesekai/crmission"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/plugins/moesekai/suite"
	"moebot-next/internal/renderer"

	"github.com/rs/zerolog/log"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
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

type characterRankMissionOptions = crmission.Options
type characterRankMissionPayload = crmission.Payload
type characterRankMissionRow = crmission.Row
type characterRankMissionAllRow = crmission.AllRow

func characterRankMissionFields() []string {
	return crmission.Fields()
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
	return crmission.ParseArgs(raw)
}

func isCharacterRankAllKeyword(value string) bool {
	return crmission.IsAllKeyword(value)
}

func buildCharacterRankMissionPayload(region string, profile characterRankMissionProfile, store *masterdata.Store, resolver interface{ RendererAssetSource() string }, options characterRankMissionOptions) (characterRankMissionPayload, string, error) {
	return crmission.BuildPayload(region, toCRMissionProfile(profile), store, resolver, options)
}

func buildCharacterRankMissionOverviewRows(store *masterdata.Store, profile characterRankMissionProfile, cid int) []characterRankMissionRow {
	return crmission.BuildOverviewRows(store, toCRMissionProfile(profile), cid)
}

func buildCharacterRankMissionRow(store *masterdata.Store, profile characterRankMissionProfile, cid int, missionType string, exSeq int) characterRankMissionRow {
	return crmission.BuildRow(store, toCRMissionProfile(profile), cid, missionType, exSeq)
}

func buildCharacterRankMissionAllRows(store *masterdata.Store, profile characterRankMissionProfile, cid int, missionType string) ([]characterRankMissionAllRow, error) {
	rows, _, err := crmission.BuildAllRows(store, toCRMissionProfile(profile), cid, missionType)
	return rows, err
}

func characterRankMissionTitle(missionType string) string {
	return crmission.Title(missionType)
}

func characterRankMissionTypeByAlias(raw string) string {
	return crmission.TypeByAlias(raw)
}

func characterRankMissionParameterGroupID(missionType string) int {
	return crmission.ParameterGroupID(missionType)
}

func toCRMissionProfile(profile characterRankMissionProfile) crmission.Profile {
	missions := make([]crmission.MissionV2, 0, len(profile.Missions))
	for _, mission := range profile.Missions {
		missions = append(missions, crmission.MissionV2{CharacterID: mission.CharacterID, CharacterMissionType: mission.CharacterMissionType, Progress: mission.Progress})
	}
	statuses := make([]crmission.MissionV2Status, 0, len(profile.Statuses))
	for _, status := range profile.Statuses {
		statuses = append(statuses, crmission.MissionV2Status{CharacterID: status.CharacterID, ParameterGroupID: status.ParameterGroupID, Seq: status.Seq, MissionStatus: status.MissionStatus})
	}
	return crmission.Profile{
		BaseProfile:  profile.BaseProfile,
		UserGamedata: profile.UserGamedata,
		Characters: append([]struct {
			CharacterID   int `json:"characterId"`
			CharacterRank int `json:"characterRank"`
		}(nil), profile.Characters...),
		Missions: missions,
		Statuses: statuses,
	}
}

func sendCharacterRankMissionOrText(ctx *zero.Ctx, deps *Deps, payload characterRankMissionPayload, fallback string) {
	if deps == nil || deps.Renderer == nil {
		ctx.SendChain(message.Text(fallback))
		return
	}

	logger := log.With().Str("template", "character_rank_mission").Str("mode", payload.Mode).Int("character_id", payload.CharacterID).Str("mission_type", payload.MissionType).Int("rows", len(payload.Rows)).Int("all_rows_rendered", len(payload.AllRows)).Int("all_rows_total", payload.AllRowsTotal).Int("page", payload.Page).Int("total_pages", payload.TotalPages).Logger()
	ok, status, healthErr := deps.Renderer.HealthWithTimeout(2 * time.Second)
	if !ok {
		logger.Warn().Err(healthErr).Int("status", status).Msg("CR mission renderer health check failed; falling back to text")
		ctx.SendChain(message.Text(fallback))
		return
	}

	started := time.Now()
	logger.Info().Msg("Rendering CR mission payload")
	png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "character_rank_mission", Data: payload})
	if err == nil {
		logger.Info().Dur("elapsed", time.Since(started)).Int("png_bytes", len(png)).Msg("Rendered CR mission payload")
		ctx.SendChain(message.ImageBytes(png))
		return
	}
	logger.Warn().Err(err).Dur("elapsed", time.Since(started)).Msg("CR mission render failed; falling back to text")
	ctx.SendChain(message.Text(fallback))
}
