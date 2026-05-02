package commands

import (
	"moebot-next/internal/database"
	"moebot-next/internal/masterdata"
	"moebot-next/internal/ranking"
	"moebot-next/internal/renderer"
	"moebot-next/internal/sekai"

	"github.com/rs/zerolog/log"
)

// Deps holds shared dependencies for all commands.
type Deps struct {
	Store    *masterdata.Store
	DB       *database.DB
	Renderer *renderer.Client
	Sekai    *sekai.Client
	Ranking  *ranking.Client
}

// RegisterAll registers all bot commands.
func RegisterAll(deps *Deps) {
	log.Info().Msg("Registering bot commands...")

	RegisterHelp()
	RegisterCard(deps)
	RegisterMusic(deps)
	RegisterEvent(deps)
	RegisterGacha(deps)
	RegisterProfile(deps)
	RegisterRanking(deps)
	RegisterBirthday()
	// RegisterSticker(deps)  // Phase 2
	// RegisterGuess(deps)    // Phase 3

	log.Info().Msg("All commands registered")
}
