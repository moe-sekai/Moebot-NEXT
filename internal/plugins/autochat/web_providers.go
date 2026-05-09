package autochat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// testProviderRequest 测试提供商连通性。
type testProviderRequest struct {
	Type    string `json:"type"`     // "openai" | "anthropic" | "embedding" | "rerank"
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Timeout int    `json:"timeout"`
}

// handleTestProvider 测试指定 provider 是否可达（GET /models 或简单 ping）。
func (p *pluginImpl) handleTestProvider(c *fiber.Ctx) error {
	var body testProviderRequest
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if body.BaseURL == "" {
		return fiber.NewError(fiber.StatusBadRequest, "base_url is required")
	}
	if body.Timeout <= 0 {
		body.Timeout = 10
	}

	client := &http.Client{Timeout: time.Duration(body.Timeout) * time.Second}
	baseURL := strings.TrimRight(body.BaseURL, "/")
	var url string

	switch body.Type {
	case "anthropic":
		// Anthropic doesn't have /models; test with a messages request that will fail but prove connectivity
		url = baseURL + "/v1/messages"
		req, err := http.NewRequest("POST", url, strings.NewReader(`{"model":"ping","max_tokens":1,"messages":[{"role":"user","content":"ping"}]}`))
		if err != nil {
			return c.JSON(fiber.Map{"ok": false, "error": err.Error()})
		}
		req.Header.Set("Content-Type", "application/json")
		if body.APIKey != "" {
			req.Header.Set("x-api-key", body.APIKey)
			req.Header.Set("anthropic-version", "2023-06-01")
		}
		resp, err := client.Do(req)
		if err != nil {
			return c.JSON(fiber.Map{"ok": false, "error": err.Error()})
		}
		defer resp.Body.Close()
		// Any non-timeout response means the server is reachable
		// 401 = bad key but reachable, 400 = bad model but reachable, etc.
		if resp.StatusCode == 401 {
			return c.JSON(fiber.Map{"ok": false, "error": "API Key 无效（401 Unauthorized）", "reachable": true})
		}
		return c.JSON(fiber.Map{"ok": true, "status": resp.StatusCode, "message": "连接成功"})

	default: // openai, embedding, rerank — all OpenAI-compatible
		url = baseURL + "/models"
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return c.JSON(fiber.Map{"ok": false, "error": err.Error()})
		}
		if body.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+body.APIKey)
		}
		resp, err := client.Do(req)
		if err != nil {
			return c.JSON(fiber.Map{"ok": false, "error": err.Error()})
		}
		defer resp.Body.Close()
		if resp.StatusCode == 401 {
			return c.JSON(fiber.Map{"ok": false, "error": "API Key 无效（401 Unauthorized）", "reachable": true})
		}
		if resp.StatusCode != 200 {
			return c.JSON(fiber.Map{"ok": false, "error": fmt.Sprintf("HTTP %d", resp.StatusCode), "reachable": true})
		}
		return c.JSON(fiber.Map{"ok": true, "status": 200, "message": "连接成功"})
	}
}

// listModelsRequest 获取模型列表。
type listModelsRequest struct {
	Type    string `json:"type"`     // "openai" | "anthropic" | "embedding" | "rerank"
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Timeout int    `json:"timeout"`
	Prefix  string `json:"prefix"`  // 给模型名加上的前缀，如 "openai" -> "openai:gpt-4o"
}

// handleListModels 查询 OpenAI 兼容 /models 端点返回的模型列表。
func (p *pluginImpl) handleListModels(c *fiber.Ctx) error {
	var body listModelsRequest
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if body.BaseURL == "" {
		return fiber.NewError(fiber.StatusBadRequest, "base_url is required")
	}
	if body.Timeout <= 0 {
		body.Timeout = 15
	}

	client := &http.Client{Timeout: time.Duration(body.Timeout) * time.Second}
	baseURL := strings.TrimRight(body.BaseURL, "/")

	if body.Type == "anthropic" {
		// Anthropic has no /models API; return a static known list
		known := []string{
			"claude-sonnet-4-20250514",
			"claude-3-7-sonnet-20250219",
			"claude-3-5-sonnet-20241022",
			"claude-3-5-haiku-20241022",
			"claude-3-haiku-20240307",
		}
		models := make([]fiber.Map, 0, len(known))
		for _, m := range known {
			id := m
			if body.Prefix != "" {
				id = body.Prefix + ":" + m
			}
			models = append(models, fiber.Map{"id": id, "name": m})
		}
		return c.JSON(fiber.Map{"models": models, "source": "static"})
	}

	// OpenAI compatible /models
	url := baseURL + "/models"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if body.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+body.APIKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fiber.NewError(resp.StatusCode, fmt.Sprintf("upstream: %s", string(raw)))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err := json.Unmarshal(raw, &result); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "parse models response: "+err.Error())
	}

	models := make([]fiber.Map, 0, len(result.Data))
	for _, m := range result.Data {
		id := m.ID
		if body.Prefix != "" {
			id = body.Prefix + ":" + m.ID
		}
		models = append(models, fiber.Map{"id": id, "name": m.ID})
	}
	return c.JSON(fiber.Map{"models": models, "source": "remote"})
}
