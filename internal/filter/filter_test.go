package filter

import "testing"

func testMessage(groupID, userID int64, raw string) *OneBotMessage {
	return &OneBotMessage{
		Partial: OneBotMessagePartial{
			MessageType:   MessageTypeGroup,
			MessageFormat: MessageFormatString,
			MessageString: raw,
			RawMessage:    raw,
			UserID:        userID,
			GroupID:       groupID,
		},
	}
}

func testFilter(clientID IDRule, defaultClientID IDRule) (*Filter, *[]Event) {
	events := []Event{}
	f := &Filter{}
	f.Compile(CompiledRules{
		Name:            "test",
		ClientID:        clientID,
		UserID:          IDRule{Mode: ModeOn},
		GroupID:         IDRule{Mode: ModeOn},
		Message:         MessageRule{Mode: ModeOn},
		PrivateMessage:  MessageRule{Mode: ModeDefault},
		GroupMessage:    MessageRule{Mode: ModeDefault},
		DefaultClientID: defaultClientID,
		DefaultUserID:   IDRule{Mode: ModeOn},
		DefaultGroupID:  IDRule{Mode: ModeOn},
	})
	f.SetPublisher(func(ev Event) { events = append(events, ev) })
	return f, &events
}

func TestClientIDWhitelistBlocksMissingAndAllowsListed(t *testing.T) {
	f, events := testFilter(IDRule{Mode: ModeWhitelist, IDs: []int64{10001}}, IDRule{Mode: ModeOn})
	if f.AllowFromClient(testMessage(20001, 30001, "/help"), 0, false) {
		t.Fatal("missing client id should not pass a client-id whitelist")
	}
	if len(*events) == 0 || (*events)[len(*events)-1].Reason != "client_id" {
		t.Fatalf("expected client_id block event, got %+v", *events)
	}
	if !f.AllowFromClient(testMessage(20001, 30001, "/help"), 10001, false) {
		t.Fatal("listed client id should pass whitelist")
	}
}

func TestClientIDBlacklistBlocksListed(t *testing.T) {
	f, _ := testFilter(IDRule{Mode: ModeBlacklist, IDs: []int64{10002}}, IDRule{Mode: ModeOn})
	if f.AllowFromClient(testMessage(20001, 30001, "/help"), 10002, false) {
		t.Fatal("blacklisted client id should be blocked")
	}
	if !f.AllowFromClient(testMessage(20001, 30001, "/help"), 10003, false) {
		t.Fatal("non-blacklisted client id should pass")
	}
	if !f.AllowFromClient(testMessage(20001, 30001, "/help"), 0, false) {
		t.Fatal("missing client id should not be blocked by blacklist")
	}
}

func TestClientIDDefaultFallback(t *testing.T) {
	f, _ := testFilter(IDRule{Mode: ModeDefault}, IDRule{Mode: ModeWhitelist, IDs: []int64{10001}})
	if !f.AllowFromClient(testMessage(20001, 30001, "/help"), 10001, false) {
		t.Fatal("client id should pass inherited whitelist")
	}
	if f.AllowFromClient(testMessage(20001, 30001, "/help"), 10002, false) {
		t.Fatal("client id should be blocked by inherited whitelist")
	}
}

func TestAllowClientEventUsesClientIDRuleOnly(t *testing.T) {
	f, events := testFilter(IDRule{Mode: ModeOff}, IDRule{Mode: ModeOn})
	if f.AllowClientEvent(10001, false) {
		t.Fatal("mode=off should block non-message events from this client")
	}
	if len(*events) != 1 || (*events)[0].Reason != "client_id" || (*events)[0].ClientID != 10001 {
		t.Fatalf("unexpected event: %+v", *events)
	}
}
