package parsing

import (
	"testing"

	"mycelia/commands"
)

func TestParseDataV1_SendMessage_OK(t *testing.T) {
	tokens := []string{"send_message", "id-1", "orders.created", "hello"}
	typ, cmd := parseDataV1(tokens)

	if typ != "send_message" {
		t.Fatalf("type mismatch: want %q, got %q", "send_message", typ)
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

func TestParseDataV1_AddRoute_OK(t *testing.T) {
	tokens := []string{"add_route", "id-2", "orders"}
	typ, cmd := parseDataV1(tokens)

	if typ != "add_route" || cmd == nil {
		t.Fatalf("want add_route non-nil, got typ=%q cmd=%v", typ, cmd)
	}
	ar, ok := cmd.(*commands.AddRoute)
	if !ok {
		t.Fatalf("expected *commands.AddRoute, got %T", cmd)
	}
	if ar.ID != "id-2" || ar.Name != "orders" {
		t.Fatalf("fields mismatch: %+v", *ar)
	}
}

func TestParseDataV1_AddSubscriber_OK(t *testing.T) {
	tokens := []string{
		"add_subscriber", "id-3", "orders", "primary", "127.0.0.1:7001",
	}
	typ, cmd := parseDataV1(tokens)

	if typ != "add_subscriber" || cmd == nil {
		t.Fatalf("want add_subscriber non-nil, got typ=%q cmd=%v", typ, cmd)
	}
	as, ok := cmd.(*commands.AddSubscriber)
	if !ok {
		t.Fatalf("expected *commands.AddSubscriber, got %T", cmd)
	}
	if as.ID != "id-3" || as.Route != "orders" || as.Channel != "primary" || as.Address != "127.0.0.1:7001" {
		t.Fatalf("fields mismatch: %+v", *as)
	}
}

func TestParseDataV1_AddChannel_OK(t *testing.T) {
	tokens := []string{"add_channel", "id-4", "orders", "primary"}
	typ, cmd := parseDataV1(tokens)

	if typ != "add_channel" || cmd == nil {
		t.Fatalf("want add_channel non-nil, got typ=%q cmd=%v", typ, cmd)
	}
	ac, ok := cmd.(*commands.AddChannel)
	if !ok {
		t.Fatalf("expected *commands.AddChannel, got %T", cmd)
	}
	if ac.ID != "id-4" || ac.Route != "orders" || ac.Name != "primary" {
		t.Fatalf("fields mismatch: %+v", *ac)
	}
}

func TestParseDataV1_AddTransformer_OK(t *testing.T) {
	tokens := []string{
		"add_transformer", "id-5", "orders", "primary", "127.0.0.1:7100",
	}
	typ, cmd := parseDataV1(tokens)

	if typ != "add_transformer" || cmd == nil {
		t.Fatalf("want add_transformer non-nil, got typ=%q cmd=%v", typ, cmd)
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
	if typ != "send_message" {
		t.Fatalf("want type 'send_message', got %q", typ)
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd for bad length, got %T", cmd)
	}
}

func TestParseAddRouteV1_BadLength_ReturnsNil(t *testing.T) {
	typ, cmd := parseAddRouteV1([]string{"only-one"})
	if typ != "add_route" || cmd != nil {
		t.Fatalf("expected add_route + nil cmd, got typ=%q cmd=%v", typ, cmd)
	}
}

func TestParseAddSubscriberV1_BadLength_ReturnsNil(t *testing.T) {
	typ, cmd := parseAddSubscriberV1([]string{"id", "route", "only-three"})
	if typ != "add_subscriber" || cmd != nil {
		t.Fatalf(
			"expected add_subscriber + nil cmd, got typ=%q cmd=%v", typ, cmd,
		)
	}
}

func TestParseAddChannelV1_BadLength_ReturnsNil(t *testing.T) {
	typ, cmd := parseAddChannelV1([]string{"id", "route"})
	if typ != "add_channel" || cmd != nil {
		t.Fatalf("expected add_channel + nil cmd, got typ=%q cmd=%v", typ, cmd)
	}
}

func TestParseAddTransformerV1_BadLength_ReturnsNil(t *testing.T) {
	typ, cmd := parseAddTransformerV1([]string{"id", "route", "channel"})
	// missing address
	if typ != "add_transformer" || cmd != nil {
		t.Fatalf(
			"expected add_transformer + nil cmd, got typ=%q cmd=%v", typ, cmd,
		)
	}
}
