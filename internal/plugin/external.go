package plugin

// ExternalDiscover is an optional hook that a control bridge can install at
// init time so that plugins registered through FloatTech/zbputils/control are
// reflected in our Manifest list and respond to enable/disable from the WebUI.
//
// When non-nil, Registry.AllRegistered and Plugins() will append the plugins
// returned by ExternalDiscover() to the in-process list. Each returned plugin
// must wrap the upstream registration into a plugin.Plugin (typically by
// embedding a no-op Init that delegates to control.Lookup at call time).
//
// The default value is nil; the project compiles cleanly without zbputils.
// To wire it up, add a file under internal/plugin/control_bridge.go behind
// a build tag and assign ExternalDiscover in init().
var ExternalDiscover func() []Plugin

// applyExternal merges externally discovered plugins into the given slice,
// preserving stable ordering (existing first, then external sorted by name).
func applyExternal(in []Plugin) []Plugin {
	if ExternalDiscover == nil {
		return in
	}
	extra := ExternalDiscover()
	if len(extra) == 0 {
		return in
	}
	seen := map[string]struct{}{}
	for _, p := range in {
		seen[p.Manifest().Name] = struct{}{}
	}
	for _, p := range extra {
		if p == nil {
			continue
		}
		name := p.Manifest().Name
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		in = append(in, p)
	}
	return in
}
