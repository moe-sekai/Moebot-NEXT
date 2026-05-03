package bot

import (
	"fmt"
	"time"

	"moebot-next/internal/database"
	"moebot-next/internal/models"

	"github.com/rs/zerolog/log"
	zero "github.com/wdvxdr1123/ZeroBot"
)

// RegisterMiddleware sets up global middleware for the bot.
func RegisterMiddleware(db *database.DB) {
	// Logging middleware: log every message received
	zero.On("message").SetBlock(false).SetPriority(0).Handle(func(ctx *zero.Ctx) {
		log.Debug().
			Str("user_id", fmt.Sprintf("%d", ctx.Event.UserID)).
			Str("group_id", fmt.Sprintf("%d", ctx.Event.GroupID)).
			Str("message", ctx.Event.RawMessage).
			Msg("Message received")
	})

	log.Info().Msg("Bot middleware registered")
}

// RecordCommand records a command invocation to the database.
func RecordCommand(db *database.DB, command string, ctx *zero.Ctx, startTime time.Time) {
	RecordCommandRegion(db, command, "", ctx, startTime)
}

// RecordCommandRegion records a command invocation with the game server region used.
func RecordCommandRegion(db *database.DB, command string, region string, ctx *zero.Ctx, startTime time.Time) {
	elapsed := time.Since(startTime).Milliseconds()

	stat := &models.CommandStat{
		Command:    command,
		Platform:   "onebot",
		UserID:     fmt.Sprintf("%d", ctx.Event.UserID),
		GroupID:    fmt.Sprintf("%d", ctx.Event.GroupID),
		Region:     region,
		Args:       fmt.Sprintf("%v", ctx.State["args"]),
		ResponseMs: elapsed,
	}

	if err := db.RecordCommandStat(stat); err != nil {
		log.Warn().Err(err).Str("command", command).Msg("Failed to record command stat")
	}
}
