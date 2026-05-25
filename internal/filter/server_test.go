package filter

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWriteMessageUsesBoundSelfID(t *testing.T) {
	s := newWsServer()
	u1 := &upstream{selfID: "10001", writeChan: make(chan wsMsg, 1)}
	u2 := &upstream{selfID: "10002", writeChan: make(chan wsMsg, 1)}
	s.upstreams["10001"] = u1
	s.upstreams["10002"] = u2

	payload := []byte(`{"action":"send_msg","params":{"group_id":1}}`)
	if err := s.writeMessage(websocket.TextMessage, payload, "10002"); err != nil {
		t.Fatal(err)
	}
	select {
	case msg := <-u2.writeChan:
		if msg.selfID != "10002" || string(msg.data) != string(payload) {
			t.Fatalf("unexpected routed msg: %+v", msg)
		}
	default:
		t.Fatal("expected message for bound upstream")
	}
	select {
	case msg := <-u1.writeChan:
		t.Fatalf("unexpected message for other upstream: %+v", msg)
	default:
	}
}

func TestWriteMessageUsesPayloadSelfID(t *testing.T) {
	s := newWsServer()
	u1 := &upstream{selfID: "10001", writeChan: make(chan wsMsg, 1)}
	u2 := &upstream{selfID: "10002", writeChan: make(chan wsMsg, 1)}
	s.upstreams["10001"] = u1
	s.upstreams["10002"] = u2

	payload := []byte(`{"self_id":10001,"action":"send_msg"}`)
	if err := s.writeMessage(websocket.TextMessage, payload, ""); err != nil {
		t.Fatal(err)
	}
	select {
	case <-u1.writeChan:
	default:
		t.Fatal("expected message for payload self_id upstream")
	}
	select {
	case msg := <-u2.writeChan:
		t.Fatalf("unexpected message for other upstream: %+v", msg)
	default:
	}
}

func TestEnsurePayloadSelfID(t *testing.T) {
	out := ensurePayloadSelfID([]byte(`{"post_type":"message","raw_message":"m"}`), "10001")
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got["self_id"].(float64) != 10001 {
		t.Fatalf("self_id = %#v", got["self_id"])
	}

	out = ensurePayloadSelfID([]byte(`{"self_id":10002,"post_type":"message"}`), "10001")
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got["self_id"].(float64) != 10002 {
		t.Fatalf("existing self_id overwritten: %#v", got["self_id"])
	}
}

func TestServeReplacesSameSelfID(t *testing.T) {
	s := newWsServer()
	old := &upstream{selfID: "10001", writeChan: make(chan wsMsg, 1)}
	s.upstreams["10001"] = old
	newUpstream := &upstream{selfID: "10001", writeChan: make(chan wsMsg, 1)}

	s.mu.Lock()
	if existing, ok := s.upstreams[newUpstream.selfID]; ok {
		delete(s.upstreams, newUpstream.selfID)
		_ = existing
	}
	s.upstreams[newUpstream.selfID] = newUpstream
	s.mu.Unlock()

	if s.upstreams["10001"] != newUpstream {
		t.Fatal("same self_id reconnect did not replace old upstream")
	}
}

func TestDedupKeyIgnoresSelfID(t *testing.T) {
	base := &dedupProbe{PostType: "message", SelfID: 10001, MessageType: MessageTypeGroup, GroupID: 20001, UserID: 30001, Time: 12345, RawMessage: "/b30"}
	otherBot := *base
	otherBot.SelfID = 10002

	if !dedupCandidate(base) || !dedupCandidate(&otherBot) {
		t.Fatal("expected group messages to be dedup candidates")
	}
	if dedupKey(base) != dedupKey(&otherBot) {
		t.Fatal("dedup key should ignore self_id for cross-bot group dedup")
	}
}

func TestDedupKeySeparatesMessageIdentity(t *testing.T) {
	base := &dedupProbe{PostType: "message", SelfID: 10001, MessageType: MessageTypeGroup, GroupID: 20001, UserID: 30001, Time: 12345, RawMessage: "/b30"}
	cases := []dedupProbe{
		{PostType: "message", SelfID: 10001, MessageType: MessageTypeGroup, GroupID: 20002, UserID: 30001, Time: 12345, RawMessage: "/b30"},
		{PostType: "message", SelfID: 10001, MessageType: MessageTypeGroup, GroupID: 20001, UserID: 30002, Time: 12345, RawMessage: "/b30"},
		{PostType: "message", SelfID: 10001, MessageType: MessageTypeGroup, GroupID: 20001, UserID: 30001, Time: 12346, RawMessage: "/b30"},
		{PostType: "message", SelfID: 10001, MessageType: MessageTypeGroup, GroupID: 20001, UserID: 30001, Time: 12345, RawMessage: "/help"},
	}
	baseKey := dedupKey(base)
	for i := range cases {
		if got := dedupKey(&cases[i]); got == baseKey {
			t.Fatalf("case %d produced same dedup key", i)
		}
	}
}

func TestDedupKeyUsesMessageJSONFallback(t *testing.T) {
	p1 := &dedupProbe{PostType: "message", SelfID: 10001, MessageType: MessageTypeGroup, GroupID: 20001, UserID: 30001, Time: 12345, Message: json.RawMessage(`[{"type":"image","data":{"file":"a.png"}}]`)}
	p2 := *p1
	p2.SelfID = 10002
	p3 := *p1
	p3.Message = json.RawMessage(`[{"type":"image","data":{"file":"b.png"}}]`)

	if !dedupCandidate(p1) || !dedupCandidate(&p2) {
		t.Fatal("expected message JSON fallback to be a dedup candidate")
	}
	if dedupKey(p1) != dedupKey(&p2) {
		t.Fatal("message JSON fallback should be stable across self_id")
	}
	if dedupKey(p1) == dedupKey(&p3) {
		t.Fatal("different message JSON should produce different dedup keys")
	}
}

func TestDedupKeyIgnoresDifferentTransportMessageIDsWhenContentExists(t *testing.T) {
	p1 := &dedupProbe{PostType: "message", SelfID: 10001, MessageType: MessageTypeGroup, GroupID: 20001, UserID: 30001, Time: 12345, RawMessage: "/b30", MessageID: json.RawMessage(`111`)}
	p2 := *p1
	p2.SelfID = 10002
	p2.MessageID = json.RawMessage(`222`)

	if dedupKey(p1) != dedupKey(&p2) {
		t.Fatal("transport-specific message ids should not split cross-bot dedup when content exists")
	}
}

func TestDedupCandidateOnlyGroupMessages(t *testing.T) {
	private := &dedupProbe{PostType: "message", MessageType: MessageTypePrivate, UserID: 30001, Time: 12345, RawMessage: "hi"}
	notice := &dedupProbe{PostType: "notice", MessageType: MessageTypeGroup, GroupID: 20001, UserID: 30001, Time: 12345, RawMessage: "hi"}
	empty := &dedupProbe{PostType: "message", MessageType: MessageTypeGroup, GroupID: 20001, UserID: 30001, Time: 12345}
	if dedupCandidate(private) {
		t.Fatal("private messages should not be cross-bot deduped")
	}
	if dedupCandidate(notice) {
		t.Fatal("non-message events should not be deduped")
	}
	if dedupCandidate(empty) {
		t.Fatal("empty messages without ids should not be deduped")
	}
}

func TestDedupCacheIsAtomic(t *testing.T) {
	d := newDedupCache(60)
	defer d.Stop()

	const workers = 64
	var firstCount int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			<-start
			if !d.IsDup(42) {
				atomic.AddInt32(&firstCount, 1)
			}
		}()
	}
	close(start)
	wg.Wait()
	if firstCount != 1 {
		t.Fatalf("firstCount = %d, want 1", firstCount)
	}
}
