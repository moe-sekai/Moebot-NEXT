// Package webroutes hosts the moesekai plugin's HTTP routes. It is registered
// onto the shared *web.Server's fiber app at plugin Init time so that web/
// no longer owns PJSK-specific routes.
//
// Each Register* function takes a typed Deps struct so this package never
// imports internal/web (which would create a cycle).
package webroutes

import (
	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/b30"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/plugins/moesekai/servers"
	"moebot-next/internal/renderer"
)

// Deps is the dependency bag passed by the plugin to each Register* function.
// Fields may be nil if the corresponding subsystem is not configured.
type Deps struct {
	Config     *config.Config
	ConfigPath string
	Renderer   *renderer.Client
	Servers    *servers.Manager
	Store      *masterdata.Store
	B30        *b30.Client
	// SaveConfig persists s.Config back to disk after a mutation. Provided by
	// the caller so we don't reinvent a `config.Save` wrapper here.
	SaveConfig func() error
}
