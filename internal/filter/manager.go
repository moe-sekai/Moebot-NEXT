package filter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"moebot-next/internal/database"
	"moebot-next/internal/models"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Manager owns the lifecycle of the OneBot filter gateway.
//
// Responsibilities:
//   - Read/write the gateway + apps configuration from the database
//   - Run an HTTP server that accepts a single OneBot client (reverse-WS)
//   - Maintain ws clients to each enabled downstream FilterApp
//   - Hot-reload all clients on configuration change
type Manager struct {
	db *database.DB

	mu        sync.Mutex
	cancel    context.CancelFunc
	httpSrv   *http.Server
	server    *wsServer
	upgrader  websocket.Upgrader
	clients   map[string]*wsClient
	filters   map[string]*Filter
	bus       *eventBus
	gateway   models.FilterGateway
	debug     bool
	startedAt time.Time
	running   bool
}

// New constructs a Manager. Call Start to bring the gateway up.
func New(db *database.DB) *Manager {
	return &Manager{
		db:      db,
		clients: map[string]*wsClient{},
		filters: map[string]*Filter{},
		bus:     newEventBus(512),
	}
}

// RecentEvents returns the last `limit` filter events.
func (m *Manager) RecentEvents(limit int) []Event {
	return m.bus.Snapshot(limit)
}

// Subscribe returns a channel of new events and an unsubscribe func.
func (m *Manager) Subscribe() (<-chan Event, func()) {
	return m.bus.Subscribe()
}

// Start brings the gateway up. It is a no-op when the configuration is disabled.
// Returns nil when disabled (so callers can ignore the result).
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.running {
		return errors.New("filter manager already running")
	}

	gw, err := m.db.GetOrCreateFilterGateway()
	if err != nil {
		return fmt.Errorf("load filter gateway: %w", err)
	}
	m.gateway = *gw
	if !gw.Enabled {
		log.Info().Msg("Filter gateway disabled in config; not starting")
		return nil
	}

	m.upgrader = websocket.Upgrader{
		ReadBufferSize:  gw.BufferSize,
		WriteBufferSize: gw.BufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	m.server = newWsServer()

	mux := http.NewServeMux()
	suffix := gw.Suffix
	if suffix == "" {
		suffix = "/ws"
	}
	mux.HandleFunc(suffix, m.handleUpstream)
	addr := fmt.Sprintf("%s:%d", gw.Host, gw.Port)
	m.httpSrv = &http.Server{Addr: addr, Handler: mux}

	managerCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.debug = gw.Debug
	m.startedAt = time.Now()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		cancel()
		m.cancel = nil
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	go func() {
		log.Info().Str("addr", addr).Str("path", suffix).Msg("Filter gateway listening")
		if err := m.httpSrv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("Filter gateway server stopped")
		}
	}()

	if err := m.startClientsLocked(managerCtx); err != nil {
		log.Warn().Err(err).Msg("Filter clients failed to (re)load")
	}
	m.running = true
	return nil
}

// Stop gracefully shuts the gateway down.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running {
		return
	}
	m.running = false
	m.stopClientsLocked()
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	if m.httpSrv != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = m.httpSrv.Shutdown(shutdownCtx)
		m.httpSrv = nil
	}
	m.server = nil
}

// Reload picks up new config from the database without dropping the upstream connection.
func (m *Manager) Reload(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running {
		return nil
	}
	gw, err := m.db.GetOrCreateFilterGateway()
	if err != nil {
		return err
	}
	m.gateway = *gw
	m.debug = gw.Debug
	m.stopClientsLocked()
	return m.startClientsLocked(ctx)
}

// IsRunning reports whether the gateway is up.
func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// Status returns a small struct for the WebUI status endpoint.
type Status struct {
	Running    bool             `json:"running"`
	Listen     string           `json:"listen"`
	Suffix     string           `json:"suffix"`
	UpstreamUp bool             `json:"upstream_up"`
	StartedAt  *time.Time       `json:"started_at,omitempty"`
	Upstreams  []UpstreamStatus `json:"upstreams"`
	Clients    []ClientStatus   `json:"clients"`
}

// UpstreamStatus describes one connected OneBot client.
type UpstreamStatus struct {
	SelfID    string     `json:"self_id"`
	Remote    string     `json:"remote"`
	Connected bool       `json:"connected"`
	Since     *time.Time `json:"since,omitempty"`
}

type ClientStatus struct {
	Name      string `json:"name"`
	URI       string `json:"uri"`
	Connected bool   `json:"connected"`
	Builtin   bool   `json:"builtin"`
}

// Status returns a snapshot of the gateway and clients.
func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := Status{
		Running: m.running,
		Suffix:  m.gateway.Suffix,
	}
	if m.running {
		st.Listen = fmt.Sprintf("%s:%d", m.gateway.Host, m.gateway.Port)
		t := m.startedAt
		st.StartedAt = &t
	}
	if m.server != nil {
		st.Upstreams = m.server.snapshotUpstreams()
		st.UpstreamUp = len(st.Upstreams) > 0
	}
	apps, _ := m.db.ListFilterApps()
	for _, app := range apps {
		c, ok := m.clients[app.Name]
		st.Clients = append(st.Clients, ClientStatus{
			Name:      app.Name,
			URI:       app.URI,
			Builtin:   app.Builtin,
			Connected: ok && c.isConnected(),
		})
	}
	return st
}

func (m *Manager) handleUpstream(w http.ResponseWriter, r *http.Request) {
	if m.server == nil {
		http.Error(w, "filter gateway not ready", http.StatusServiceUnavailable)
		return
	}
	// Read the configured token under the lock; the field is short and copying
	// avoids holding the lock across the upgrade/serve call.
	m.mu.Lock()
	expected := m.gateway.AccessToken
	m.mu.Unlock()
	if expected != "" && !checkAccessToken(r, expected) {
		log.Warn().Str("remote", r.RemoteAddr).Msg("Filter: upstream rejected, token mismatch")
		w.Header().Set("WWW-Authenticate", `Bearer realm="moebot-filter"`)
		http.Error(w, "invalid access token", http.StatusUnauthorized)
		return
	}
	selfID := r.Header.Get("x-self-id")
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warn().Err(err).Msg("Filter: upstream upgrade failed")
		return
	}
	log.Info().Str("remote", r.RemoteAddr).Str("self_id", selfID).Msg("Filter: upstream OneBot client connected")
	m.bus.Publish(Event{Kind: EventUpstreamUp, Reason: r.RemoteAddr, Filter: selfID})
	defer m.bus.Publish(Event{Kind: EventUpstreamDown, Reason: r.RemoteAddr, Filter: selfID})
	if err := m.server.serve(r.Context(), conn, selfID, r.RemoteAddr); err != nil {
		log.Info().Err(err).Str("self_id", selfID).Msg("Filter: upstream OneBot client disconnected")
	}
}

// checkAccessToken validates the upstream request's access token against the
// configured gateway token. OneBot v11 conventions: `Authorization: Bearer <t>`
// or `Authorization: Token <t>`, plus `?access_token=<t>` query fallback.
func checkAccessToken(r *http.Request, expected string) bool {
	if expected == "" {
		return true
	}
	if got := r.URL.Query().Get("access_token"); got != "" && got == expected {
		return true
	}
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false
	}
	for _, prefix := range []string{"Bearer ", "Token "} {
		if strings.HasPrefix(auth, prefix) {
			return strings.TrimSpace(auth[len(prefix):]) == expected
		}
	}
	// Allow bare token (some clients omit the scheme prefix).
	return strings.TrimSpace(auth) == expected
}

func (m *Manager) startClientsLocked(ctx context.Context) error {
	apps, err := m.db.ListFilterApps()
	if err != nil {
		return err
	}
	templates, err := m.db.ListFilterTemplates()
	if err != nil {
		return err
	}
	tplByID := map[uint]*models.FilterTemplate{}
	for i := range templates {
		t := &templates[i]
		tplByID[t.ID] = t
	}
	defaultTpl, err := m.db.GetDefaultFilterTemplate()
	if err != nil {
		return err
	}
	defaultUserID := DecodeIDRule(defaultTpl.UserIDRules)
	defaultGroupID := DecodeIDRule(defaultTpl.GroupIDRules)
	snap := gatewaySnapshot{
		BotID:      m.gateway.BotID,
		UserAgent:  m.gateway.UserAgent,
		BufferSize: m.gateway.BufferSize,
		SleepTime:  m.gateway.SleepTime,
		Debug:      m.gateway.Debug,
	}
	if snap.BufferSize <= 0 {
		snap.BufferSize = 4096
	}
	if snap.SleepTime <= 0 {
		snap.SleepTime = 5
	}
	for _, app := range apps {
		if !app.Enabled {
			continue
		}
		// When the app references a template, source rules from it; otherwise
		// from the app's own fields.
		userID, groupID, msg, priv, grp := appOrTemplateRules(&app, tplByID)
		// "moebot-builtin" 是网关→Bot 主进程的传输闸门：把它当成纯透传，
		// 防止与各 plugin:<name> 的 internal app 形成 AND 串联语义。
		// 任何用户残留的规则在此处被强制忽略（控制台 UI 也会隐藏其规则编辑器）。
		if IsBuiltinTransport(app.Name) {
			userID = IDRule{Mode: ModeOn}
			groupID = IDRule{Mode: ModeOn}
			msg = MessageRule{Mode: ModeOn}
			priv = MessageRule{Mode: ModeOn}
			grp = MessageRule{Mode: ModeOn}
		}
		f := &Filter{}
		f.Compile(CompiledRules{
			Name:           app.Name,
			UserID:         userID,
			GroupID:        groupID,
			Message:        msg,
			PrivateMessage: priv,
			GroupMessage:   grp,
			DefaultUserID:  defaultUserID,
			DefaultGroupID: defaultGroupID,
		})
		f.SetPublisher(m.bus.Publish)
		m.filters[app.Name] = f
		// Internal apps 仅作为规则容器，不需要 ws 下游客户端。
		if app.Internal {
			continue
		}
		c := newWsClient(app.Name, app.URI, app.AccessToken, f, m.debug, m.bus.Publish)
		m.clients[app.Name] = c
		go c.run(ctx, m.server, snap)
	}
	return nil
}

// AllowMessage 让插件按 filter app 的规则过滤消息事件。
// 适用于"内部 app"——插件自身处理消息，但希望复用控制台的 group_id /
// user_id / 文本规则做白名单/黑名单/前缀/正则等过滤。
//
// 规则的命中口径与下游 ws 客户端完全一致（共用同一份 Filter.Allow）。
//
// 找不到对应名字的 app（未启用 / 未 seed）时返回 true，让插件按默认行为
// 走自己的逻辑——插件应当在调用前自检 IsAppEnabled。
func (m *Manager) AllowMessage(appName string, groupID, userID int64, isPrivate bool, raw string) bool {
	m.mu.Lock()
	f := m.filters[appName]
	m.mu.Unlock()
	if f == nil {
		return true
	}
	mt := MessageTypeGroup
	if isPrivate {
		mt = MessageTypePrivate
	}
	probe := &OneBotMessage{
		Partial: OneBotMessagePartial{
			MessageType:   mt,
			MessageFormat: MessageFormatString,
			MessageString: raw,
			RawMessage:    raw,
			UserID:        userID,
			GroupID:       groupID,
		},
	}
	return f.Allow(probe, m.debug)
}

// IsAppEnabled 报告 filter 网关是否已为某个 app 编译了规则。
// 当 app 不存在 / 未启用 / 网关未运行时返回 false。
func (m *Manager) IsAppEnabled(appName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.filters[appName]
	return ok
}

// appOrTemplateRules returns the five rule values that should be used to compile
// `app`. When app.TemplateID is set and points to a known template, the template's
// rules win; otherwise the app's own fields are used.
func appOrTemplateRules(app *models.FilterApp, tplByID map[uint]*models.FilterTemplate) (IDRule, IDRule, MessageRule, MessageRule, MessageRule) {
	if app.TemplateID != nil {
		if t, ok := tplByID[*app.TemplateID]; ok {
			return DecodeIDRule(t.UserIDRules),
				DecodeIDRule(t.GroupIDRules),
				DecodeMessageRule(t.MessageRules),
				DecodeMessageRule(t.PrivateMessageRules),
				DecodeMessageRule(t.GroupMessageRules)
		}
	}
	return DecodeIDRule(app.UserIDRules),
		DecodeIDRule(app.GroupIDRules),
		DecodeMessageRule(app.MessageRules),
		DecodeMessageRule(app.PrivateMessageRules),
		DecodeMessageRule(app.GroupMessageRules)
}

func (m *Manager) stopClientsLocked() {
	for name, c := range m.clients {
		close(c.stop)
		<-c.stopped
		_ = name
	}
	m.clients = map[string]*wsClient{}
	m.filters = map[string]*Filter{}
}
