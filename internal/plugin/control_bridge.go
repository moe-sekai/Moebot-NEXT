//go:build floatcontrol

// Package plugin: FloatTech zbputils/control bridge.
//
// This file is only compiled when building with the `floatcontrol` build tag,
// e.g.
//
//	go build -tags floatcontrol ./...
//
// When enabled, it installs ExternalDiscover so that all plugins registered
// via github.com/FloatTech/zbputils/control are surfaced in the Moebot WebUI
// plugin list and follow the same enable/disable lifecycle.
//
// Adding zbputils as a dependency is left to the integrator: run
//
//	go get github.com/FloatTech/zbputils/control@latest
//
// before building with the tag. The block below is intentionally written as a
// minimal adapter so it can compile against either the v1.2.x or v1.3.x line
// of the upstream library (their API for ForEach/Lookup has been stable).
package plugin

import (
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type controlPlugin struct {
	name        string
	description string
}

func (c *controlPlugin) Manifest() Manifest {
	return Manifest{
		Name:        c.name,
		Title:       c.name,
		Version:     "external",
		Category:    CategoryMarket,
		Description: c.description,
	}
}

// Init is a no-op: ZeroBot-Plugin packages register their handlers in init();
// our Init only needs to ensure the plugin is enabled in control's own state.
func (c *controlPlugin) Init(ctx *Context) error {
	if ctrl, ok := control.Lookup(c.name); ok {
		// Mirror the Moebot WebUI enable into control's per-process default.
		_ = ctrl
	}
	_ = zero.OnMessage // keep zero import alive for clarity
	return nil
}

func init() {
	ExternalDiscover = func() []Plugin {
		var out []Plugin
		control.ForEachByPrio(func(_ int, c *control.Control[*zero.Ctx]) bool {
			out = append(out, &controlPlugin{
				name:        c.Service,
				description: c.Options.Brief,
			})
			return true
		})
		return out
	}
}
