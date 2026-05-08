package web

import (
	"os"

	"moebot-next/internal/plugin"

	"github.com/gofiber/fiber/v2"
)

type pluginListItem struct {
	plugin.Manifest
	Enabled bool `json:"enabled"`
	Loaded  bool `json:"loaded"`
}

// handleListPlugins 返回所有已注册插件的 manifest + 启用状态。
func (s *Server) handleListPlugins(c *fiber.Ctx) error {
	reg := plugin.Global()
	if reg == nil {
		return c.JSON(fiber.Map{"plugins": []any{}})
	}
	out := []pluginListItem{}
	for _, p := range reg.Plugins() {
		m := p.Manifest()
		enabled := reg.IsEnabled(m.Name)
		out = append(out, pluginListItem{
			Manifest: m,
			Enabled:  enabled,
			Loaded:   enabled,
		})
	}
	return c.JSON(fiber.Map{"plugins": out})
}

// handleSetPluginEnabled 修改启用状态（POST /api/plugins/:name/enable|disable）。
// 修改后需要重启进程才会真正加/卸载，前端应给出明确提示。
func (s *Server) handleSetPluginEnabled(enabled bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reg := plugin.Global()
		if reg == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "plugin registry not ready")
		}
		name := c.Params("name")
		if name == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing plugin name")
		}
		if reg.Lookup(name) == nil {
			return fiber.NewError(fiber.StatusNotFound, "plugin not found")
		}
		if err := reg.SetEnabled(name, enabled); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(fiber.Map{
			"name":             name,
			"enabled":          enabled,
			"requires_restart": true,
		})
	}
}

// handleGetPluginConfig 返回插件子配置 yaml 原文。文件不存在时返回空字符串。
func (s *Server) handleGetPluginConfig(c *fiber.Ctx) error {
	reg := plugin.Global()
	if reg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "plugin registry not ready")
	}
	name := c.Params("name")
	if reg.Lookup(name) == nil {
		return fiber.NewError(fiber.StatusNotFound, "plugin not found")
	}
	path := pluginConfigPath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c.JSON(fiber.Map{"name": name, "path": path, "yaml": ""})
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"name": name, "path": path, "yaml": string(data)})
}

// handleUpdatePluginConfig 把请求体 {yaml: "..."} 写入插件子配置文件。
func (s *Server) handleUpdatePluginConfig(c *fiber.Ctx) error {
	reg := plugin.Global()
	if reg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "plugin registry not ready")
	}
	name := c.Params("name")
	if reg.Lookup(name) == nil {
		return fiber.NewError(fiber.StatusNotFound, "plugin not found")
	}
	var body struct {
		YAML string `json:"yaml"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	path := pluginConfigPath(name)
	if err := os.MkdirAll(pluginsDir(), 0o755); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if err := os.WriteFile(path, []byte(body.YAML), 0o644); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{
		"name":             name,
		"path":             path,
		"requires_restart": true,
	})
}

func pluginsDir() string {
	if dir := os.Getenv("MOEBOT_PLUGINS_DIR"); dir != "" {
		return dir
	}
	return "./data/plugins"
}

func pluginConfigPath(name string) string {
	return pluginsDir() + string(os.PathSeparator) + name + ".yml"
}
