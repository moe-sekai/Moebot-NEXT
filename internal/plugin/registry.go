package plugin

import (
	"errors"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"moebot-next/internal/models"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Registry 是进程内的插件注册表与启用状态管理器。
// 通常通过 init() 在各插件包内调用 Register 完成自注册：
//
//	func init() { plugin.Register(&moesekaiPlugin{}) }
//
// main.go 拿到 db 后构造 Registry，然后调用 InitEnabled 启动所有
// 启用的插件。
type Registry struct {
	mu      sync.RWMutex
	plugins []Plugin

	db        *gorm.DB
	dataDir   string // data/plugins/
	contexts  map[string]*Context
	restartCh chan struct{}
}

var (
	globalMu       sync.Mutex
	globalPlugins  []Plugin
	globalRegistry *Registry
)

// Register 把一个插件实例加入全局注册表。线程安全；多次注册同名插件会被忽略。
//
// 该函数通常在每个插件包的 init() 里调用，因此调用顺序由 import 顺序决定。
func Register(p Plugin) {
	if p == nil {
		return
	}
	m := p.Manifest()
	if m.Name == "" {
		log.Warn().Msg("plugin: ignoring registration with empty manifest name")
		return
	}
	globalMu.Lock()
	defer globalMu.Unlock()
	for _, existing := range globalPlugins {
		if existing.Manifest().Name == m.Name {
			log.Warn().Str("plugin", m.Name).Msg("plugin: duplicate registration ignored")
			return
		}
	}
	globalPlugins = append(globalPlugins, p)
}

// AllRegistered 返回当前进程已注册的插件清单（按名称字典序）。
//
// 同时把 ExternalDiscover（若已安装）暴露的上游 ZeroBot-Plugin / control
// 插件并入。
func AllRegistered() []Plugin {
	globalMu.Lock()
	out := make([]Plugin, len(globalPlugins))
	copy(out, globalPlugins)
	globalMu.Unlock()
	out = applyExternal(out)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Manifest().Name < out[j].Manifest().Name
	})
	return out
}

// NewRegistry 构造运行期 Registry。
func NewRegistry(db *gorm.DB, dataDir string) *Registry {
	r := &Registry{
		db:        db,
		dataDir:   dataDir,
		contexts:  map[string]*Context{},
		restartCh: make(chan struct{}, 1),
	}
	r.plugins = AllRegistered()
	globalMu.Lock()
	globalRegistry = r
	globalMu.Unlock()
	return r
}

// SeedDefaults 写入初始启用状态：对于 DB 中尚不存在记录的插件，按 defaults
// 列表（包含其 Name）默认为 enabled，否则为 disabled。已有记录不会被覆盖。
func (r *Registry) SeedDefaults(defaults []string) error {
	defaultSet := map[string]struct{}{}
	for _, name := range defaults {
		defaultSet[name] = struct{}{}
	}
	for _, p := range r.plugins {
		name := p.Manifest().Name
		var existing models.PluginState
		err := r.db.Where("name = ?", name).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		_, enabled := defaultSet[name]
		state := models.PluginState{Name: name, Enabled: enabled, UpdatedAt: time.Now()}
		if err := r.db.Create(&state).Error; err != nil {
			return err
		}
	}
	return nil
}

// IsEnabled 查询启用状态。
func (r *Registry) IsEnabled(name string) bool {
	var s models.PluginState
	if err := r.db.Where("name = ?", name).First(&s).Error; err != nil {
		return false
	}
	return s.Enabled
}

// SetEnabled 持久化启用状态变更。
//
//   - enabled=false 时立即触发该插件已注册的 OnShutdown 钩子（例如停止
//     周期任务、断开外部连接），从而无需重启即可"停服"。
//   - enabled=true 仍需要进程重启才会重新 Init（webroutes / 命令注册等无
//     法在 fiber/zerobot 运行期安全重复挂载）。
func (r *Registry) SetEnabled(name string, enabled bool) error {
	found := false
	for _, p := range r.plugins {
		if p.Manifest().Name == name {
			found = true
			break
		}
	}
	if !found {
		return errors.New("plugin not found: " + name)
	}
	s := models.PluginState{Name: name, Enabled: enabled, UpdatedAt: time.Now()}
	if err := r.db.Save(&s).Error; err != nil {
		return err
	}
	if !enabled {
		r.mu.Lock()
		ctx, ok := r.contexts[name]
		if ok {
			delete(r.contexts, name)
		}
		r.mu.Unlock()
		if ok && ctx != nil {
			log.Info().Str("plugin", name).Msg("plugin disabled, running shutdown hooks")
			ctx.RunShutdownHooks()
		}
	} else {
		// 启用 = 需要重新 Init（命令 / Web 路由等无法运行时安全热加），
		// 触发一次进程内 supervisor 重启，让该插件随重启一起加载。
		// 若该插件其实已经处于 loaded 状态（例如重复点击），则不重启。
		if !r.IsLoaded(name) {
			log.Info().Str("plugin", name).Msg("plugin enabled, requesting in-process restart")
			r.RequestRestart()
		}
	}
	return nil
}

// RequestRestart 向 supervisor 发出"重启进程"信号；非阻塞，重复请求会被合并。
func (r *Registry) RequestRestart() {
	if r == nil || r.restartCh == nil {
		return
	}
	select {
	case r.restartCh <- struct{}{}:
	default:
	}
}

// RestartChan 返回 supervisor 用来等待重启请求的只读通道。
func (r *Registry) RestartChan() <-chan struct{} {
	if r == nil {
		return nil
	}
	return r.restartCh
}

// IsLoaded 返回插件当前是否已 Init 且未 Shutdown。与 IsEnabled 不同，
// 后者反映持久化偏好；本方法反映运行期实际状态。
func (r *Registry) IsLoaded(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.contexts[name]
	return ok
}

// Plugins 返回 Registry 里的插件副本（按名字典序）。
func (r *Registry) Plugins() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Plugin, len(r.plugins))
	copy(out, r.plugins)
	return out
}

// Lookup 按名字查找。
func (r *Registry) Lookup(name string) Plugin {
	for _, p := range r.plugins {
		if p.Manifest().Name == name {
			return p
		}
	}
	return nil
}

// InitEnabled 调用所有启用插件的 Init，传入由 baseCtx 构造的 Context。
// baseCtx 包含全局依赖；每个插件得到独立的 Context 副本，PluginName /
// PluginConfigPath 已填充。
//
// 任意单个插件 Init 失败只会日志告警，不影响其它插件。
func (r *Registry) InitEnabled(baseCtx Context) {
	for _, p := range r.plugins {
		name := p.Manifest().Name
		if !r.IsEnabled(name) {
			log.Info().Str("plugin", name).Msg("plugin disabled, skipping init")
			continue
		}
		ctx := baseCtx
		ctx.PluginName = name
		ctx.PluginDataDir = r.dataDir
		ctx.PluginConfigPath = filepath.Join(r.dataDir, name+".yml")
		if err := p.Init(&ctx); err != nil {
			log.Error().Err(err).Str("plugin", name).Msg("plugin init failed")
			continue
		}
		r.mu.Lock()
		r.contexts[name] = &ctx
		r.mu.Unlock()
		log.Info().Str("plugin", name).Str("version", p.Manifest().Version).Msg("plugin loaded")
	}
}

// Shutdown 调用所有已加载插件的关闭钩子。
func (r *Registry) Shutdown() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for name, ctx := range r.contexts {
		log.Debug().Str("plugin", name).Msg("plugin shutdown hooks")
		ctx.RunShutdownHooks()
	}
	r.contexts = map[string]*Context{}
}

// Global 返回当前进程构造的 Registry，未构造时返回 nil。
// 仅供 web 层等需要查询启用状态/manifest 的地方使用。
func Global() *Registry {
	globalMu.Lock()
	defer globalMu.Unlock()
	return globalRegistry
}
