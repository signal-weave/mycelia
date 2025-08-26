package parsing

import (
	"testing"

	"mycelia/commands"
)

func TestParseDataV1_SendMessage_OK(t *testing.T) {
	tokens := []string{"MESSAGE.SEND", "id-1", "orders.created", "hello"}
	typ, cmd := parseDataV1(tokens)

	if typ != "MESSAGE.SEND" {
		t.Fatalf("type mismatch: want %q, got %q", "MESSAGE.SEND", typ)
	}
	if cmd == nil {
		t.Fatalf("expected non-nil command")
	}

	sm, ok := cmd.(*commands.SendMessage)
	if !ok {
		t.Fatalf("expected *commands.SendMessage, got %T", cmd)
	}
	if sm.ID != "id-1" || sm.Route != "orders.created" || sm.Body != "hello" {
		t.Fatalf("fields mismatch: %+v", *sm)
	}
}

func TestParseDataV1_AddSubscriber_OK(t *testing.T) {
	tokens := []string{
		"SUBSCRIBER.ADD", "id-3", "orders", "primary", "127.0.0.1:7001",
	}
	typ, cmd := parseDataV1(tokens)

	if typ != "SUBSCRIBER.ADD" || cmd == nil {
		t.Fatalf("want SUBSCRIBER.ADD non-nil, got typ=%q cmd=%v", typ, cmd)
	}
	as, ok := cmd.(*commands.AddSubscriber)
	if !ok {
		t.Fatalf("expected *commands.AddSubscriber, got %T", cmd)
	}
	if as.ID != "id-3" || as.Route != "orders" || as.Channel != "primary" || as.Address != "127.0.0.1:7001" {
		t.Fatalf("fields mismatch: %+v", *as)
	}
}

func TestParseDataV1_AddTransformer_OK(t *testing.T) {
	tokens := []string{
		"TRANSFORMER.ADD", "id-5", "orders", "primary", "127.0.0.1:7100",
	}
	typ, cmd := parseDataV1(tokens)

	if typ != "TRANSFORMER.ADD" || cmd == nil {
		t.Fatalf("want TRANSFORMER.ADD non-nil, got typ=%q cmd=%v", typ, cmd)
	}
	at, ok := cmd.(*commands.AddTransformer)
	if !ok {
		t.Fatalf("expected *commands.AddTransformer, got %T", cmd)
	}
	if at.ID != "id-5" || at.Route != "orders" || at.Channel != "primary" || at.Address != "127.0.0.1:7100" {
		t.Fatalf("fields mismatch: %+v", *at)
	}
}

func TestParseDataV1_Unknown_ReturnsEmptyAndNil(t *testing.T) {
	tokens := []string{"totally_unknown", "id-9"}
	typ, cmd := parseDataV1(tokens)
	if typ != "" || cmd != nil {
		t.Fatalf("expected empty type and nil cmd, got typ=%q cmd=%v", typ, cmd)
	}
}

// -------token length validation paths (verifyTokenLength false)---------------

func TestParseSendMsgV1_BadLength_ReturnsNil(t *testing.T) {
	typ, cmd := parseSendMsgV1([]string{"only-two", "tokens"})
	if typ != "MESSAGE.SEND" {
		t.Fatalf("want type 'MESSAGE.SEND', got %q", typ)
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd for bad length, got %T", cmd)
	}
}

func TestParseAddSubscriberV1_BadLength_ReturnsNil(t *testing.T) {
	typ, cmd := parseAddSubscriberV1([]string{"id", "route", "only-three"})
	if typ != "SUBSCRIBER.ADD" || cmd != nil {
		t.Fatalf(
			"expected SUBSCRIBER.ADD + nil cmd, got typ=%q cmd=%v", typ, cmd,
		)
	}
}

func TestParseAddTransformerV1_BadLength_ReturnsNil(t *testing.T) {
	typ, cmd := parseAddTransformerV1([]string{"id", "route", "channel"})
	// missing address
	if typ != "TRANSFORMER.ADD" || cmd != nil {
		t.Fatalf(
			"expected TRANSFORMER.ADD + nil cmd, got typ=%q cmd=%v", typ, cmd,
		)
	}
}
