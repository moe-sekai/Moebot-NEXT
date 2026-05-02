package bot

import (
	"fmt"

	"moebot-next/internal/config"

	"github.com/rs/zerolog/log"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
)

// Bot wraps the ZeroBot instance and its configuration.
type Bot struct {
	Config config.BotConfig
}

// New creates and configures a new Bot instance.
func New(cfg config.BotConfig) *Bot {
	return &Bot{Config: cfg}
}

// Run starts the ZeroBot event loop. This blocks until shutdown.
func (b *Bot) Run() {
	drivers := b.buildDrivers()
	if len(drivers) == 0 {
		log.Fatal().Msg("No drivers configured. Please configure a WebSocket driver in config.yml")
	}

	zeroCfg := zero.Config{
		NickName:      b.Config.Nickname,
		CommandPrefix: b.Config.CommandPrefix,
		SuperUsers:    b.Config.SuperUsers,
		Driver:        drivers,
	}

	log.Info().
		Strs("nickname", b.Config.Nickname).
		Str("prefix", b.Config.CommandPrefix).
		Str("driver_type", b.Config.Driver.Type).
		Msg("Starting ZeroBot")

	zero.RunAndBlock(&zeroCfg, nil)
}

// buildDrivers creates ZeroBot drivers from configuration.
func (b *Bot) buildDrivers() []zero.Driver {
	var drivers []zero.Driver

	switch b.Config.Driver.Type {
	case "ws":
		// Forward WebSocket: bot connects to the OneBot implementation
		url := b.Config.Driver.URL
		if url == "" {
			url = "ws://127.0.0.1:6700"
		}
		log.Info().Str("url", url).Msg("Using forward WebSocket driver")
		drivers = append(drivers, driver.NewWebSocketClient(url, b.Config.Driver.Token))

	case "ws-reverse":
		// Reverse WebSocket: bot listens for connections from the OneBot implementation
		listen := b.Config.Driver.Listen
		if listen == "" {
			listen = "0.0.0.0:6700"
		}
		log.Info().Str("listen", listen).Msg("Using reverse WebSocket driver")
		// 16 is the max number of connections
		drivers = append(drivers, driver.NewWebSocketServer(16, fmt.Sprintf("ws://%s", listen), b.Config.Driver.Token))

	default:
		log.Error().Str("type", b.Config.Driver.Type).Msg("Unknown driver type, defaulting to ws-reverse")
		drivers = append(drivers, driver.NewWebSocketServer(16, "ws://0.0.0.0:6700", ""))
	}

	return drivers
}
