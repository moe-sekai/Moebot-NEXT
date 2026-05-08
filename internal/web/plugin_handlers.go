package web

import (
	"os"

	"moebot-next/internal/plugin"

	"github.com/gofiber/fiber/v2"
)

type pluginListItem struct {
	plugin.Manifest
	Enabled      bool `json:"enabled"`
	Loaded       bool `json:"loaded"`
	Configurable bool `json:"configurable"`
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
		out = append(out, pluginListItem{
			Manifest:     m,
			Enabled:      reg.IsEnabled(m.Name),
			Loaded:       reg.IsLoaded(m.Name),
			Configurable: isConfigurable(p),
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
		// 启用 / 禁用均在进程内即时生效：禁用触发 OnShutdown 钩子，启用
		// 触发 re-Init（由 Registry.SetEnabled 内部处理）。只有当
		// registry.baseCtx 尚未就绪（理论上不会发生）时才会 fallback 到
		// 进程重启，此时 IsLoaded 仍为 false，requires_restart=true。
		loaded := reg.IsLoaded(name)
		return c.JSON(fiber.Map{
			"name":             name,
			"enabled":          enabled,
			"loaded":           loaded,
			"requires_restart": enabled && !loaded,
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

// handleGetPluginSettings 返回 schema (Manifest.Settings) + 当前值。
// 仅当插件实现了 plugin.Configurable 时才返回 values；否则 values 为空。
func (s *Server) handleGetPluginSettings(c *fiber.Ctx) error {
	reg := plugin.Global()
	if reg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "plugin registry not ready")
	}
	name := c.Params("name")
	p := reg.Lookup(name)
	if p == nil {
		return fiber.NewError(fiber.StatusNotFound, "plugin not found")
	}
	m := p.Manifest()
	values := map[string]any{}
	if cp, ok := p.(plugin.Configurable); ok {
		v, err := cp.GetSettings()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if v != nil {
			values = v
		}
	}
	return c.JSON(fiber.Map{
		"name":         name,
		"schema":       m.Settings,
		"values":       values,
		"configurable": isConfigurable(p),
	})
}

// handleUpdatePluginSettings 接收 {values: {key: any}}，调用 Configurable.UpdateSettings。
func (s *Server) handleUpdatePluginSettings(c *fiber.Ctx) error {
	reg := plugin.Global()
	if reg == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "plugin registry not ready")
	}
	name := c.Params("name")
	p := reg.Lookup(name)
	if p == nil {
		return fiber.NewError(fiber.StatusNotFound, "plugin not found")
	}
	cp, ok := p.(plugin.Configurable)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "plugin is not configurable")
	}
	var body struct {
		Values map[string]any `json:"values"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := cp.UpdateSettings(body.Values); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	values, _ := cp.GetSettings()
	if values == nil {
		values = map[string]any{}
	}
	return c.JSON(fiber.Map{"name": name, "values": values})
}

func isConfigurable(p plugin.Plugin) bool {
	_, ok := p.(plugin.Configurable)
	return ok
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
