package web

import (
	"strconv"
	"strings"

	"moebot-next/internal/logbuffer"

	"github.com/gofiber/fiber/v2"
)

// handleListLogs returns recent log entries from the in-memory ring buffer.
//
// Query parameters:
//   - level     comma-separated levels (debug,info,warn,error,fatal)
//   - q         case-insensitive substring match on message + fields
//   - limit     max entries to return (default 200, max 1000)
//   - since_seq only return entries with Seq > since_seq for incremental polling
func (s *Server) handleListLogs(c *fiber.Ctx) error {
	if s.Logs == nil {
		return c.JSON(fiber.Map{
			"data":      []logbuffer.Entry{},
			"total":     0,
			"dropped":   0,
			"next_seq":  0,
			"capacity":  0,
			"available": false,
			"message":   "日志缓冲未初始化",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "200"))
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}
	sinceSeq, _ := strconv.ParseUint(c.Query("since_seq", "0"), 10, 64)

	levels := splitCSV(c.Query("level"))
	query := strings.TrimSpace(c.Query("q"))

	entries, _ := s.Logs.Snapshot(logbuffer.FilterOpts{
		Levels:   levels,
		Query:    query,
		SinceSeq: sinceSeq,
		Limit:    limit,
	})
	total, dropped, nextSeq := s.Logs.Stats()

	return c.JSON(fiber.Map{
		"data":      entries,
		"total":     total,
		"dropped":   dropped,
		"next_seq":  nextSeq,
		"capacity":  s.Logs.Capacity(),
		"available": true,
	})
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
