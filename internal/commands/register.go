package commands

import (
	"moebot-next/internal/database"
	"moebot-next/internal/renderer"
	"moebot-next/internal/servers"

	"github.com/rs/zerolog/log"
)

// Deps holds shared dependencies for all commands.
type Deps struct {
	DB       *database.DB
	Renderer *renderer.Client
	Servers  *servers.Manager
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
