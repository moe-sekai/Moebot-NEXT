package commands

import (
	"moebot-next/internal/b30"
	"moebot-next/internal/commandparser"
	"moebot-next/internal/database"
	"moebot-next/internal/renderer"
	"moebot-next/internal/servers"

	"github.com/rs/zerolog/log"
)

// Deps holds shared dependencies for all commands.
type Deps struct {
	DB          *database.DB
	Renderer    *renderer.Client
	Servers     *servers.Manager
	B30         *b30.Client
	Definitions []commandparser.Definition
}

// RegisterAll registers all bot commands.
func RegisterAll(deps *Deps) {
	log.Info().Msg("Registering bot commands...")
	if len(deps.Definitions) == 0 {
		deps.Definitions = commandparser.BaseDefinitions()
	}

	RegisterHelp(deps)
	RegisterCard(deps)
	RegisterMusic(deps)
	RegisterEvent(deps)
	RegisterGacha(deps)
	RegisterVirtualLive(deps)
	RegisterGachaSimulation(deps)
	RegisterProfile(deps)
	RegisterSuite(deps)
	RegisterBond(deps)
	RegisterMusicOverview(deps)
	RegisterBest30(deps)
	RegisterChallengeInfo(deps)
	RegisterEventRecord(deps)
	RegisterLeaderCount(deps)
	RegisterCharacterRankMission(deps)
	RegisterAnvo(deps)
	RegisterSuiteCardBox(deps)
	RegisterRanking(deps)
	RegisterBirthday()
	// RegisterSticker(deps)  // Phase 2
	// RegisterGuess(deps)    // Phase 3

	log.Info().Msg("All commands registered")
}
