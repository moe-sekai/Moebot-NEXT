package filter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// dedupCache 提供基于 TTL 的消息去重缓存。
type dedupCache struct {
	ttl   time.Duration
	store sync.Map // key(uint64) → expireAt(int64, UnixNano)
	stop  chan struct{}
}

func newDedupCache(ttlSeconds int) *dedupCache {
	if ttlSeconds <= 0 {
		ttlSeconds = 60
	}
	d := &dedupCache{
		ttl:  time.Duration(ttlSeconds) * time.Second,
		stop: make(chan struct{}),
	}
	go d.cleanup()
	return d
}

// IsDup 检查 key 是否已在缓存中（未过期）。若不存在则写入并返回 false。
func (d *dedupCache) IsDup(key uint64) bool {
	now := time.Now().UnixNano()
	if v, ok := d.store.Load(key); ok {
		if v.(int64) > now {
			return true
		}
	}
	d.store.Store(key, now+int64(d.ttl))
	return false
}

func (d *dedupCache) cleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-d.stop:
			return
		case <-ticker.C:
			now := time.Now().UnixNano()
			d.store.Range(func(k, v any) bool {
				if v.(int64) <= now {
					d.store.Delete(k)
				}
				return true
			})
		}
	}
}

func (d *dedupCache) Stop() {
	close(d.stop)
}

// dedupProbe 是从 OneBot 事件 JSON 中提取的去重字段。
type dedupProbe struct {
	PostType   string `json:"post_type"`
	GroupID    int64  `json:"group_id"`
	UserID     int64  `json:"user_id"`
	Time       int64  `json:"time"`
	RawMessage string `json:"raw_message"`
}

func dedupKey(p *dedupProbe) uint64 {
	h := fnv.New64a()
	fmt.Fprintf(h, "%d:%d:%d:%s", p.GroupID, p.UserID, p.Time, p.RawMessage)
	return h.Sum64()
}

// wsServer is the upstream side: it accepts one or more connections from
// real OneBot clients (e.g. NapCat) keyed by their self-id, and fans incoming
// events out to all downstream clients. Outgoing action calls coming back from
// downstream clients are routed by `self_id` when present.
type wsServer struct {
	mu        sync.RWMutex
	upstreams map[string]*upstream
	dedup     *dedupCache

	clientsMu sync.RWMutex
	clients   []*wsClient
}

// upstream represents a single connected OneBot client.
type upstream struct {
	selfID    string
	remote    string
	conn      *websocket.Conn
	writeChan chan wsMsg
	connected time.Time
}

func newWsServer() *wsServer {
	return &wsServer{upstreams: map[string]*upstream{}}
}

// serve runs the read/write loops for one OneBot client connection. It blocks
// until the connection closes or ctx is cancelled. selfID is taken from the
// upstream's `x-self-id` header; if empty a synthetic id is used so we can
// still hold the connection.
func (s *wsServer) serve(ctx context.Context, conn *websocket.Conn, selfID, remote string) error {
	if selfID == "" {
		selfID = "anon-" + remote
	}
	u := &upstream{
		selfID:    selfID,
		remote:    remote,
		conn:      conn,
		writeChan: make(chan wsMsg, 64),
		connected: time.Now(),
	}
	s.mu.Lock()
	if existing, ok := s.upstreams[selfID]; ok {
		s.mu.Unlock()
		_ = existing
		return fmt.Errorf("filter: upstream self-id %s already connected", selfID)
	}
	s.upstreams[selfID] = u
	s.mu.Unlock()

	innerCtx, cancel := context.WithCancel(ctx)
	defer s.removeUpstream(selfID, cancel)

	go s.upstreamWriteLoop(innerCtx, u)

	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		if s.dedup != nil && mt == websocket.TextMessage {
			var p dedupProbe
			if json.Unmarshal(data, &p) == nil && p.PostType == "message" && p.RawMessage != "" {
				if s.dedup.IsDup(dedupKey(&p)) {
					log.Debug().Str("self_id", selfID).Int64("group_id", p.GroupID).Msg("filter: dedup skipped duplicate message")
					continue
				}
			}
		}
		s.broadcastToClients(wsMsg{mt, data})
	}
}

// removeUpstream stops a single upstream connection and cleans state.
func (s *wsServer) removeUpstream(selfID string, cancel context.CancelFunc) {
	cancel()
	s.mu.Lock()
	u, ok := s.upstreams[selfID]
	if ok {
		delete(s.upstreams, selfID)
	}
	s.mu.Unlock()
	if ok && u.conn != nil {
		_ = u.conn.Close()
	}
}

// broadcastToClients fans an upstream-originated message to every downstream client.
func (s *wsServer) broadcastToClients(msg wsMsg) {
	for _, c := range s.snapshotClients() {
		go func(c *wsClient, m wsMsg) {
			if err := c.write(m.mt, m.data); err != nil {
				log.Debug().Str("client", c.name).Err(err).Msg("filter: forward to client failed")
			}
		}(c, msg)
	}
}

// writeMessage routes a downstream-originated message to the correct upstream(s).
// If `self_id` is present in the JSON payload and matches a connected upstream,
// only that upstream receives the message; otherwise we fall back to broadcast
// (covers single-account deployments and best-effort multi-account routing).
func (s *wsServer) writeMessage(mt int, data []byte) error {
	s.mu.RLock()
	count := len(s.upstreams)
	s.mu.RUnlock()
	if count == 0 {
		return errors.New("filter: no upstream OneBot client")
	}

	target := ""
	if mt == websocket.TextMessage {
		var probe struct {
			SelfID json.Number `json:"self_id"`
		}
		if err := json.Unmarshal(data, &probe); err == nil {
			target = probe.SelfID.String()
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	if target != "" {
		if u, ok := s.upstreams[target]; ok {
			return enqueue(u, wsMsg{mt, data})
		}
	}
	// Fan-out to all upstreams.
	var firstErr error
	for _, u := range s.upstreams {
		if err := enqueue(u, wsMsg{mt, data}); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func enqueue(u *upstream, msg wsMsg) error {
	select {
	case u.writeChan <- msg:
		return nil
	default:
		return fmt.Errorf("filter: upstream %s write channel full", u.selfID)
	}
}

func (s *wsServer) addClient(c *wsClient) error {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for _, existing := range s.clients {
		if existing.name == c.name {
			return fmt.Errorf("filter: client %s already connected", c.name)
		}
	}
	s.clients = append(s.clients, c)
	return nil
}

func (s *wsServer) removeClient(name string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for i, c := range s.clients {
		if c.name == name {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			return
		}
	}
}

func (s *wsServer) snapshotClients() []*wsClient {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	out := make([]*wsClient, len(s.clients))
	copy(out, s.clients)
	return out
}

// snapshotUpstreams returns immutable info about all connected upstreams.
func (s *wsServer) snapshotUpstreams() []UpstreamStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]UpstreamStatus, 0, len(s.upstreams))
	for _, u := range s.upstreams {
		t := u.connected
		out = append(out, UpstreamStatus{
			SelfID:    u.selfID,
			Remote:    u.remote,
			Connected: true,
			Since:     &t,
		})
	}
	return out
}

func (s *wsServer) upstreamWriteLoop(ctx context.Context, u *upstream) {
	for {
		select {
		case msg, ok := <-u.writeChan:
			if !ok {
				return
			}
			if u.conn == nil {
				continue
			}
			if err := u.conn.WriteMessage(msg.mt, msg.data); err != nil {
				log.Warn().Str("self_id", u.selfID).Err(err).Msg("filter: write to upstream OneBot client failed")
			}
		case <-ctx.Done():
			return
		}
	}
}
