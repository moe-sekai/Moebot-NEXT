package webroutes

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"moebot-next/internal/config"

	"github.com/gofiber/fiber/v2"
)

// RegisterSekaiTest registers /api/config/sekai/test-system. The request hits
// SEKAI API's /system endpoint to verify reachability + auth headers; this is
// inherently moesekai-specific and lives next to the rest of the PJSK config
// surface.
func RegisterSekaiTest(api fiber.Router, d Deps) {
	h := &sekaiTestHandlers{d: d}
	api.Post("/config/sekai/test-system", h.testSystem)
}

type sekaiTestHandlers struct {
	d Deps
}

type sekaiSystemTestRequest struct {
	BaseURL string            `json:"base_url"`
	Region  string            `json:"region"`
	Headers map[string]string `json:"headers"`
	Timeout int               `json:"timeout"`
}

func (h *sekaiTestHandlers) testSystem(c *fiber.Ctx) error {
	var req sekaiSystemTestRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	region := config.NormalizeRegion(req.Region)
	if region == "" && h.d.Config != nil {
		region = config.NormalizeRegion(h.d.Config.Server.Region)
	}
	if region == "" {
		region = config.RegionJP
	}
	base := strings.TrimSpace(req.BaseURL)
	if base == "" {
		base = config.DefaultSekaiAPIURL
	}
	systemURL, err := sekaiSystemURL(base, region)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	timeout := time.Duration(req.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	httpReq, err := http.NewRequest(http.MethodGet, systemURL, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("build system request: %v", err))
	}
	for key, value := range req.Headers {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			httpReq.Header.Set(key, value)
		}
	}
	started := time.Now()
	resp, err := (&http.Client{Timeout: timeout}).Do(httpReq)
	duration := time.Since(started)
	if err != nil {
		return c.JSON(fiber.Map{
			"ok":          false,
			"url":         systemURL,
			"duration_ms": duration.Milliseconds(),
			"message":     err.Error(),
		})
	}
	defer resp.Body.Close()
	ok := resp.StatusCode >= 200 && resp.StatusCode < 300
	message := "SEKAI API /system 连通正常"
	if !ok {
		message = fmt.Sprintf("SEKAI API /system 返回 HTTP %d", resp.StatusCode)
	}
	return c.JSON(fiber.Map{
		"ok":          ok,
		"url":         systemURL,
		"status_code": resp.StatusCode,
		"duration_ms": duration.Milliseconds(),
		"message":     message,
	})
}

func sekaiSystemURL(baseURL, region string) (string, error) {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = config.DefaultSekaiAPIURL
	}
	region = config.NormalizeRegion(region)
	if region == "" {
		region = config.RegionJP
	}
	base = strings.ReplaceAll(base, "{region}", region)
	if strings.Contains(base, "{uid}") || strings.Contains(base, "{user_id}") {
		return "", fmt.Errorf("SEKAI API Base URL 用于 /system 测试时不能包含 {uid} 或 {user_id}")
	}
	if _, err := url.ParseRequestURI(base); err != nil {
		return "", fmt.Errorf("invalid sekai api base url: %w", err)
	}
	if strings.Contains(baseURL, "{region}") {
		return strings.TrimRight(base, "/") + "/system", nil
	}
	return url.JoinPath(base, "api", region, "system")
}
