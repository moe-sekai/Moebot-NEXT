package filter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// wsClient connects to one downstream bot application and forwards messages
// after running them through its filter.
type wsClient struct {
	name        string
	uri         string
	accessToken string
	filter      *Filter
	debug       bool
	publish     func(Event)

	conn      *websocket.Conn
	writeChan chan wsMsg

	connected int32 // atomic bool
	stop      chan struct{}
	stopped   chan struct{}
}

func newWsClient(name, uri, token string, f *Filter, debug bool, publish func(Event)) *wsClient {
	return &wsClient{
		name:        name,
		uri:         uri,
		accessToken: token,
		filter:      f,
		debug:       debug,
		publish:     publish,
		writeChan:   make(chan wsMsg, 64),
		stop:        make(chan struct{}),
		stopped:     make(chan struct{}),
	}
}

func (c *wsClient) emit(kind EventKind, reason string) {
	if c.publish == nil {
		return
	}
	c.publish(Event{Kind: kind, Filter: c.name, Reason: reason})
}

func (c *wsClient) isConnected() bool {
	return atomic.LoadInt32(&c.connected) == 1
}

// run keeps the client connected with reconnect backoff. It exits when stop is closed.
func (c *wsClient) run(parent context.Context, server *wsServer, gateway gatewaySnapshot) {
	defer close(c.stopped)
	header := http.Header{}
	header.Set("x-self-id", gateway.BotID)
	header.Set("user-agent", gateway.UserAgent)
	header.Set("x-client-role", "Universal")
	if c.accessToken != "" {
		header.Set("authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	}

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
		go c.writeLoop(ctx, server)
		go func() {
			for {
				mt, data, err := conn.ReadMessage()
				if err != nil {
					readErr <- err
					return
				}
				if err := server.writeMessage(mt, data); err != nil {
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

func (c *wsClient) write(mt int, data []byte) error {
	if !c.isConnected() {
		return errors.New("filter: client not connected")
	}
	select {
	case c.writeChan <- wsMsg{mt, data}:
		return nil
	default:
		return errors.New("filter: client write channel full")
	}
}

func (c *wsClient) writeLoop(ctx context.Context, _ *wsServer) {
	for {
		select {
		case msg := <-c.writeChan:
			c.handleWrite(msg)
		case <-ctx.Done():
			return
		}
	}
}

func (c *wsClient) handleWrite(msg wsMsg) {
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

// gatewaySnapshot is an immutable view passed to clients.
type gatewaySnapshot struct {
	BotID      string
	UserAgent  string
	BufferSize int
	SleepTime  float32
	Debug      bool
}
