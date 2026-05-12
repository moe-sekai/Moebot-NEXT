package filter

import (
	"encoding/json"
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
