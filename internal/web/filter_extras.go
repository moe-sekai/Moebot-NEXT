package web

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"moebot-next/internal/filter"
	"moebot-next/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

// handleFilterEvents streams filter events via Server-Sent Events.
// Query parameter `replay` (default 50) controls how many recent events are
// replayed before live tailing begins.
func (s *Server) handleFilterEvents(c *fiber.Ctx) error {
	if s.Filter == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "filter manager not available")
	}
	replay := c.QueryInt("replay", 50)
	if replay < 0 {
		replay = 0
	}
	if replay > 500 {
		replay = 500
	}
	ch, unsubscribe := s.Filter.Subscribe()
	recent := s.Filter.RecentEvents(replay)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		defer unsubscribe()
		// Replay recent events first.
		for _, ev := range recent {
			if err := writeSSE(w, ev); err != nil {
				return
			}
		}
		if err := w.Flush(); err != nil {
			return
		}
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case ev, ok := <-ch:
				if !ok {
					return
				}
				if err := writeSSE(w, ev); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
			case <-ticker.C:
				// Heartbeat keeps the connection alive across proxies.
				if _, err := fmt.Fprintf(w, ": ping %d\n\n", time.Now().Unix()); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
			}
		}
	}))
	return nil
}

func writeSSE(w *bufio.Writer, ev filter.Event) error {
	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", ev.Seq, ev.Kind, payload); err != nil {
		return err
	}
	return nil
}

// handleFilterRecentEvents returns a JSON snapshot of recent events (no streaming).
func (s *Server) handleFilterRecentEvents(c *fiber.Ctx) error {
	if s.Filter == nil {
		return c.JSON(fiber.Map{"items": []any{}})
	}
	limit := c.QueryInt("limit", 100)
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return c.JSON(fiber.Map{"items": s.Filter.RecentEvents(limit)})
}

// handleFilterTestRegex compiles and tests a regexp2 pattern against a sample.
type regexTestPayload struct {
	Pattern string `json:"pattern"`
	Text    string `json:"text"`
}

func (s *Server) handleFilterTestRegex(c *fiber.Ctx) error {
	var p regexTestPayload
	if err := c.BodyParser(&p); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if p.Pattern == "" {
		return fiber.NewError(fiber.StatusBadRequest, "pattern is required")
	}
	compiled, matched, errMsg := filter.TestRegex(p.Pattern, p.Text)
	return c.JSON(fiber.Map{
		"compiled": compiled,
		"matched":  matched,
		"error":    errMsg,
	})
}

// handleFilterExportYAML returns the current filter configuration as YAML text.
func (s *Server) handleFilterExportYAML(c *fiber.Ctx) error {
	gw, err := s.DB.GetOrCreateFilterGateway()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	apps, err := s.DB.ListFilterApps()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	defaultTpl, err := s.DB.GetDefaultFilterTemplate()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	templates, err := s.DB.ListFilterTemplates()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	out, err := filter.ExportYAML(gw, defaultTpl, templates, apps)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	c.Set("Content-Type", "application/x-yaml; charset=utf-8")
	c.Set("Content-Disposition", `attachment; filename="moebot-filter.yaml"`)
	return c.Send(out)
}

// handleFilterImportYAML accepts a YAML payload (raw text/yaml or {"yaml":"..."} JSON)
// and replaces non-builtin apps. Builtin app rules are updated when present in YAML.
type importYAMLPayload struct {
	YAML       string `json:"yaml"`
	ReplaceAll bool   `json:"replace_all"` // if true: delete existing non-builtin apps before import
}

func (s *Server) handleFilterImportYAML(c *fiber.Ctx) error {
	body := c.Body()
	contentType := string(c.Request().Header.ContentType())
	var data []byte
	var replaceAll bool
	if contentType == "application/json" {
		var p importYAMLPayload
		if err := json.Unmarshal(body, &p); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		data = []byte(p.YAML)
		replaceAll = p.ReplaceAll
	} else {
		data = body
		replaceAll = c.QueryBool("replace_all", false)
	}
	if len(data) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "empty yaml payload")
	}
	cfg, err := filter.ParseYAML(data)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	gw, err := s.DB.GetOrCreateFilterGateway()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	defaultTpl, err := s.DB.GetDefaultFilterTemplate()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	importedApps, gw := filter.ApplyYAMLToModels(cfg, gw, defaultTpl)
	if err := s.DB.UpdateFilterGateway(gw); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if err := s.DB.UpdateFilterTemplate(defaultTpl); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	existingApps, err := s.DB.ListFilterApps()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	existingByName := map[string]*models.FilterApp{}
	for i := range existingApps {
		existingByName[existingApps[i].Name] = &existingApps[i]
	}

	if replaceAll {
		for _, e := range existingApps {
			if e.Builtin {
				continue
			}
			if _, keep := findAppByName(importedApps, e.Name); keep {
				continue
			}
			_ = s.DB.DeleteFilterApp(e.ID)
		}
	}

	created, updated := 0, 0
	for i := range importedApps {
		incoming := &importedApps[i]
		if existing, ok := existingByName[incoming.Name]; ok {
			incoming.ID = existing.ID
			incoming.Builtin = existing.Builtin
			incoming.Enabled = existing.Enabled
			incoming.SortOrder = existing.SortOrder
			if err := s.DB.UpdateFilterApp(incoming); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
			updated++
			continue
		}
		if err := s.DB.CreateFilterApp(incoming); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		created++
	}

	if s.Filter != nil {
		_ = s.Filter.Reload(c.UserContext())
	}
	return c.JSON(fiber.Map{"created": created, "updated": updated, "total": len(importedApps)})
}

func findAppByName(apps []models.FilterApp, name string) (*models.FilterApp, bool) {
	for i := range apps {
		if apps[i].Name == name {
			return &apps[i], true
		}
	}
	return nil, false
}

// strconv.Atoi alias to keep import balanced even if not used elsewhere later.
var _ = strconv.Atoi
