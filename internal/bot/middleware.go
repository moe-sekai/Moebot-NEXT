package bot

import (
	"fmt"
	"time"

	"moebot-next/internal/database"
	"moebot-next/internal/models"

	"github.com/rs/zerolog/log"
	zero "github.com/wdvxdr1123/ZeroBot"
)

// middlewareEngine 是 bot 中间件（如日志）注册到的 ZeroBot 引擎；
// 每次 RegisterMiddleware 会先 Delete 旧 engine，避免进程内重启
// 后中间件被重复注册（导致每条消息日志打多次）。
var middlewareEngine *zero.Engine

// RegisterMiddleware sets up global middleware for the bot.
func RegisterMiddleware(db *database.DB) {
	if middlewareEngine != nil {
		middlewareEngine.Delete()
	}
	middlewareEngine = zero.New()
	// Logging middleware: log every message received
	middlewareEngine.On("message").SetBlock(false).SetPriority(0).Handle(func(ctx *zero.Ctx) {
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
	if db == nil || ctx == nil || ctx.Event == nil {
		return
	}
	elapsed := time.Since(startTime).Milliseconds()
	clientID := ""
	if ctx.Event.SelfID != 0 {
		clientID = fmt.Sprintf("%d", ctx.Event.SelfID)
	}
	groupID := ""
	if ctx.Event.GroupID != 0 {
		groupID = fmt.Sprintf("%d", ctx.Event.GroupID)
	}

	stat := &models.CommandStat{
		Command:    command,
		Platform:   "onebot",
		ClientID:   clientID,
		UserID:     fmt.Sprintf("%d", ctx.Event.UserID),
		GroupID:    groupID,
		Region:     region,
		Args:       fmt.Sprintf("%v", ctx.State["args"]),
		ResponseMs: elapsed,
	}

	if groupID != "" {
		if err := db.EnsureGroup("onebot", clientID, groupID, ""); err != nil {
			log.Warn().Err(err).Str("client_id", clientID).Str("group_id", groupID).Msg("Failed to ensure group")
		}
	}
	if err := db.RecordCommandStat(stat); err != nil {
		log.Warn().Err(err).Str("command", command).Msg("Failed to record command stat")
	}
}
