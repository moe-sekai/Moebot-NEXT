package filter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// wsClient connects to one downstream bot application and forwards messages
// after running them through its filter.
type wsClient struct {
	name            string
	uri             string
	accessToken     string
	filter          *Filter
	debug           bool
	publish         func(Event)
	systemTransport bool

	conn      *websocket.Conn
	writeChan chan wsMsg

	bindingsMu sync.Mutex
	bindings   map[string]*downstreamBinding

	connected int32 // atomic bool
	stop      chan struct{}
	stopped   chan struct{}
}

type downstreamBinding struct {
	selfID string
	conn   *websocket.Conn
}

func newWsClient(name, uri, token string, f *Filter, debug bool, publish func(Event), systemTransport bool) *wsClient {
	return &wsClient{
		name:            name,
		uri:             uri,
		accessToken:     token,
		filter:          f,
		debug:           debug,
		publish:         publish,
		systemTransport: systemTransport,
		writeChan:       make(chan wsMsg, 64),
		bindings:        map[string]*downstreamBinding{},
		stop:            make(chan struct{}),
		stopped:         make(chan struct{}),
	}
}

func (c *wsClient) emit(kind EventKind, reason string) {
	if c.publish == nil {
		return
	}
	c.publish(Event{Kind: kind, Filter: c.name, Reason: reason})
}

func (c *wsClient) isConnected() bool {
	if !c.systemTransport {
		return atomic.LoadInt32(&c.connected) == 1
	}
	c.bindingsMu.Lock()
	defer c.bindingsMu.Unlock()
	return len(c.bindings) > 0
}

// run keeps the client connected with reconnect backoff. It exits when stop is closed.
func (c *wsClient) run(parent context.Context, server *wsServer, gateway gatewaySnapshot) {
	if c.systemTransport {
		c.runSystemTransport(parent, server, gateway)
		return
	}
	defer close(c.stopped)
	header := c.downstreamHeader(gateway.BotID, gateway)

	for {
		select {
		case <-c.stop:
			return
		case <-parent.Done():
			return
		default:
		}

		log.Info().Str("client", c.name).Str("uri", c.uri).Msg("filter: connecting to downstream bot")
		dialer := &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 30 * time.Second,
			ReadBufferSize:   gateway.BufferSize,
			WriteBufferSize:  gateway.BufferSize,
		}
		conn, _, err := dialer.Dial(c.uri, header)
		if err != nil {
			log.Warn().Str("client", c.name).Err(err).Msg("filter: connect failed, will retry")
			if !c.sleep(parent, gateway.SleepTime) {
				return
			}
			continue
		}
		c.conn = conn
		atomic.StoreInt32(&c.connected, 1)
		if err := server.addClient(c); err != nil {
			log.Warn().Str("client", c.name).Err(err).Msg("filter: add client failed")
			_ = conn.Close()
			c.conn = nil
			atomic.StoreInt32(&c.connected, 0)
			if !c.sleep(parent, gateway.SleepTime) {
				return
			}
			continue
		}
		log.Info().Str("client", c.name).Msg("filter: connected to downstream bot")
		c.emit(EventClientUp, c.uri)

		ctx, cancel := context.WithCancel(parent)
		readErr := make(chan error, 1)
		go c.writeLoop(ctx, server, gateway)
		go func() {
			for {
				mt, data, err := conn.ReadMessage()
				if err != nil {
					readErr <- err
					return
				}
				if err := server.writeMessage(mt, data, ""); err != nil {
					log.Debug().Str("client", c.name).Err(err).Msg("filter: forward to upstream failed")
				}
			}
		}()

		select {
		case err := <-readErr:
			log.Warn().Str("client", c.name).Err(err).Msg("filter: downstream connection lost")
		case <-c.stop:
			cancel()
			_ = conn.Close()
			server.removeClient(c.name)
			atomic.StoreInt32(&c.connected, 0)
			c.conn = nil
			return
		case <-parent.Done():
			cancel()
			_ = conn.Close()
			server.removeClient(c.name)
			atomic.StoreInt32(&c.connected, 0)
			c.conn = nil
			return
		}

		cancel()
		_ = conn.Close()
		server.removeClient(c.name)
		atomic.StoreInt32(&c.connected, 0)
		c.conn = nil
		c.emit(EventClientDown, "disconnect")
		if !c.sleep(parent, gateway.SleepTime) {
			return
		}
	}
}

// runSystemTransport keeps one logical system client registered with wsServer,
// but opens one downstream ZeroBot reverse-WS connection per upstream self_id.
// This gives ZeroBot one APICaller per account, so API requests route back to
// the OneBot client that produced the event instead of being broadcast.
func (c *wsClient) runSystemTransport(parent context.Context, server *wsServer, gateway gatewaySnapshot) {
	defer close(c.stopped)
	if err := server.addClient(c); err != nil {
		log.Warn().Str("client", c.name).Err(err).Msg("filter: add system transport failed")
		return
	}
	log.Info().Str("client", c.name).Str("uri", c.uri).Msg("filter: system transport ready")
	c.emit(EventClientUp, c.uri)

	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	go c.writeLoop(ctx, server, gateway)

	select {
	case <-c.stop:
	case <-parent.Done():
	}
	cancel()
	server.removeClient(c.name)
	c.closeAllBindings()
	atomic.StoreInt32(&c.connected, 0)
	c.emit(EventClientDown, "disconnect")
}

func (c *wsClient) downstreamHeader(selfID string, gateway gatewaySnapshot) http.Header {
	header := http.Header{}
	header.Set("x-self-id", selfID)
	header.Set("user-agent", gateway.UserAgent)
	header.Set("x-client-role", "Universal")
	if c.accessToken != "" {
		header.Set("authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	}
	return header
}

func (c *wsClient) dialSystemBinding(ctx context.Context, selfID string, server *wsServer, gateway gatewaySnapshot) (*downstreamBinding, error) {
	if selfID == "" {
		return nil, errors.New("filter: empty upstream self_id")
	}
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 30 * time.Second,
		ReadBufferSize:   gateway.BufferSize,
		WriteBufferSize:  gateway.BufferSize,
	}
	conn, _, err := dialer.Dial(c.uri, c.downstreamHeader(selfID, gateway))
	if err != nil {
		return nil, err
	}
	numericSelfID, err := strconv.ParseInt(selfID, 10, 64)
	if err != nil || numericSelfID == 0 {
		_ = conn.Close()
		return nil, fmt.Errorf("filter: invalid numeric upstream self_id %q", selfID)
	}
	if err := conn.WriteJSON(struct {
		SelfID int64 `json:"self_id"`
	}{SelfID: numericSelfID}); err != nil {
		_ = conn.Close()
		return nil, err
	}
	binding := &downstreamBinding{selfID: selfID, conn: conn}
	go c.systemReadLoop(ctx, server, binding)
	return binding, nil
}

func (c *wsClient) ensureSystemBinding(ctx context.Context, selfID string, server *wsServer, gateway gatewaySnapshot) (*downstreamBinding, error) {
	c.bindingsMu.Lock()
	if existing := c.bindings[selfID]; existing != nil {
		c.bindingsMu.Unlock()
		return existing, nil
	}
	c.bindingsMu.Unlock()

	binding, err := c.dialSystemBinding(ctx, selfID, server, gateway)
	if err != nil {
		return nil, err
	}

	c.bindingsMu.Lock()
	if existing := c.bindings[selfID]; existing != nil {
		c.bindingsMu.Unlock()
		_ = binding.conn.Close()
		return existing, nil
	}
	c.bindings[selfID] = binding
	atomic.StoreInt32(&c.connected, 1)
	c.bindingsMu.Unlock()
	log.Info().Str("client", c.name).Str("self_id", selfID).Msg("filter: connected downstream bot binding")
	return binding, nil
}

func (c *wsClient) systemReadLoop(ctx context.Context, server *wsServer, binding *downstreamBinding) {
	for {
		mt, data, err := binding.conn.ReadMessage()
		if err != nil {
			select {
			case <-ctx.Done():
			default:
				log.Warn().Str("client", c.name).Str("self_id", binding.selfID).Err(err).Msg("filter: downstream binding lost")
			}
			c.closeBinding(binding.selfID, binding)
			return
		}
		if err := server.writeMessage(mt, data, binding.selfID); err != nil {
			log.Debug().Str("client", c.name).Str("self_id", binding.selfID).Err(err).Msg("filter: forward to bound upstream failed")
		}
	}
}

func (c *wsClient) closeBinding(selfID string, expected *downstreamBinding) {
	if !c.systemTransport {
		return
	}
	var conn *websocket.Conn
	c.bindingsMu.Lock()
	binding := c.bindings[selfID]
	if binding != nil && (expected == nil || binding == expected) {
		delete(c.bindings, selfID)
		conn = binding.conn
	}
	if len(c.bindings) == 0 {
		atomic.StoreInt32(&c.connected, 0)
	}
	c.bindingsMu.Unlock()
	if conn != nil {
		_ = conn.Close()
		log.Info().Str("client", c.name).Str("self_id", selfID).Msg("filter: downstream bot binding closed")
	}
}

func (c *wsClient) closeAllBindings() {
	c.bindingsMu.Lock()
	bindings := c.bindings
	c.bindings = map[string]*downstreamBinding{}
	c.bindingsMu.Unlock()
	for _, binding := range bindings {
		if binding != nil && binding.conn != nil {
			_ = binding.conn.Close()
		}
	}
}

func (c *wsClient) sleep(ctx context.Context, seconds float32) bool {
	if seconds <= 0 {
		seconds = 5
	}
	t := time.NewTimer(time.Duration(seconds * float32(time.Second)))
	defer t.Stop()
	select {
	case <-c.stop:
		return false
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func (c *wsClient) write(msg wsMsg) error {
	if !c.systemTransport && !c.isConnected() {
		return errors.New("filter: client not connected")
	}
	select {
	case c.writeChan <- msg:
		return nil
	default:
		return errors.New("filter: client write channel full")
	}
}

func (c *wsClient) writeLoop(ctx context.Context, server *wsServer, gateway gatewaySnapshot) {
	for {
		select {
		case msg := <-c.writeChan:
			c.handleWrite(ctx, server, gateway, msg)
		case <-ctx.Done():
			return
		}
	}
}

func (c *wsClient) handleWrite(ctx context.Context, server *wsServer, gateway gatewaySnapshot, msg wsMsg) {
	if c.systemTransport {
		c.handleSystemWrite(ctx, server, gateway, msg)
		return
	}
	if c.conn == nil {
		return
	}
	if msg.mt == websocket.TextMessage {
		ob := ParseOneBotMessage(msg.data)
		if ob != nil && ob.Partial.RawMessage != "" {
			if !c.filter.Allow(ob, c.debug) {
				return
			}
			if err := c.conn.WriteJSON(ob.Intact); err != nil {
				log.Warn().Str("client", c.name).Err(err).Msg("filter: write JSON to downstream failed")
			}
			return
		}
	}
	if err := c.conn.WriteMessage(msg.mt, msg.data); err != nil {
		log.Warn().Str("client", c.name).Err(err).Msg("filter: write to downstream failed")
	}
}

func (c *wsClient) handleSystemWrite(ctx context.Context, server *wsServer, gateway gatewaySnapshot, msg wsMsg) {
	selfID := msg.selfID
	if selfID == "" && msg.mt == websocket.TextMessage {
		if ob := ParseOneBotMessage(msg.data); ob != nil && ob.Partial.SelfID != 0 {
			selfID = strconv.FormatInt(ob.Partial.SelfID, 10)
		}
	}
	if selfID == "" {
		log.Debug().Str("client", c.name).Msg("filter: system transport skipped event without self_id")
		return
	}
	if msg.mt == websocket.TextMessage {
		ob := ParseOneBotMessage(msg.data)
		if ob != nil && ob.Partial.RawMessage != "" {
			if !c.filter.Allow(ob, c.debug) {
				return
			}
			msg.data = encodeRawMap(ob.Intact)
		}
	}
	binding, err := c.ensureSystemBinding(ctx, selfID, server, gateway)
	if err != nil {
		log.Warn().Str("client", c.name).Str("self_id", selfID).Err(err).Msg("filter: connect downstream binding failed")
		return
	}
	if err := binding.conn.WriteMessage(msg.mt, msg.data); err != nil {
		log.Warn().Str("client", c.name).Str("self_id", selfID).Err(err).Msg("filter: write to downstream binding failed")
		c.closeBinding(selfID, binding)
	}
}

// gatewaySnapshot is an immutable view passed to clients.
type gatewaySnapshot struct {
	BotID      string
	UserAgent  string
	BufferSize int
	SleepTime  float32
	Debug      bool
}
