package routing

import (
	"encoding/json"
	"testing"
	"time"

	"mycelia/globals"
	"mycelia/protocol"
	"mycelia/test"
)

type fakeServer struct {
	addr    string
	port    int
	updated bool
}

func (f *fakeServer) Run()               {}
func (f *fakeServer) UpdateListener()    { f.updated = true }
func (f *fakeServer) GetAddress() string { return f.addr }
func (f *fakeServer) GetPort() int       { return f.port }
func (f *fakeServer) Shutdown() {}

func TestBroker_AddSubscriber_ThenSend_ForwardsToSubscriber(t *testing.T) {
	subAddr, gotBody, stop := test.MockOneWayServer(t)
	t.Cleanup(stop)

	fs := &fakeServer{addr: "0.0.0.0", port: 1}
	b := NewBroker(fs)

	// Add subscriber: route "r1", channel "c1" -> subAddr
	addSub := protocol.NewCommand(
		globals.OBJ_SUBSCRIBER, globals.CMD_ADD,
		"sender", "uid-1", "r1", "c1", subAddr, "", nil,
	)
	if err := b.HandleCommand(addSub); err != nil {
		t.Fatalf("HandleCommand(add subscriber) error: %v", err)
	}

	// Send delivery on route "r1"
	send := protocol.NewCommand(
		globals.OBJ_DELIVERY, globals.CMD_SEND,
		"sender", "uid-2", "r1", "", "", "", []byte("hello"),
	)
	if err := b.HandleCommand(send); err != nil {
		t.Fatalf("HandleCommand(send) error: %v", err)
	}

	select {
	case got := <-gotBody:
		if got != "hello" {
			t.Fatalf(
				"subscriber payload mismatch: got %q want %q", got, "hello",
			)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for subscriber to receive payload")
	}
}

func TestBroker_AddTransformer_ThenSend_PayloadTransformedAndDelivered(t *testing.T) {
	trAddr, stopTr := test.MockTwoWayServer(t, "X:")
	t.Cleanup(stopTr)
	subAddr, gotBody, stopSub := test.MockOneWayServer(t)
	t.Cleanup(stopSub)

	fs := &fakeServer{addr: "0.0.0.0", port: 1}
	b := NewBroker(fs)
	globals.TransformTimeout = 2 * time.Second

	// Add transformer on r2.c1
	addTr := protocol.NewCommand(
		globals.OBJ_TRANSFORMER, globals.CMD_ADD,
		"sender", "uid-t", "r2", "c1", trAddr, "", nil,
	)
	if err := b.HandleCommand(addTr); err != nil {
		t.Fatalf("HandleCommand(add transformer) error: %v", err)
	}

	// Add subscriber on r2.c1
	addSub := protocol.NewCommand(
		globals.OBJ_SUBSCRIBER, globals.CMD_ADD,
		"sender", "uid-s", "r2", "c1", subAddr, "", nil,
	)
	if err := b.HandleCommand(addSub); err != nil {
		t.Fatalf("HandleCommand(add subscriber) error: %v", err)
	}

	// Send delivery through route r2
	send := protocol.NewCommand(
		globals.OBJ_DELIVERY, globals.CMD_SEND,
		"sender", "uid-3", "r2", "", "", "", []byte("ping"),
	)
	if err := b.HandleCommand(send); err != nil {
		t.Fatalf("HandleCommand(send) error: %v", err)
	}

	select {
	case got := <-gotBody:
		if want := "X:ping"; got != want {
			t.Fatalf("transformed payload mismatch: got %q want %q", got, want)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for transformed payload at subscriber")
	}
}

func TestBroker_RemoveSubscriber_NoFurtherDeliveries(t *testing.T) {
	subAddr, gotBody, stop := test.MockOneWayServer(t)
	t.Cleanup(stop)

	fs := &fakeServer{addr: "0.0.0.0", port: 1}
	b := NewBroker(fs)

	addSub := protocol.NewCommand(
		globals.OBJ_SUBSCRIBER, globals.CMD_ADD,
		"sender", "uid-1", "r3", "c1", subAddr, "", nil,
	)
	_ = b.HandleCommand(addSub)

	remSub := protocol.NewCommand(
		globals.OBJ_SUBSCRIBER, globals.CMD_REMOVE,
		"sender", "uid-2", "r3", "c1", subAddr, "", nil,
	)
	if err := b.HandleCommand(remSub); err != nil {
		t.Fatalf("HandleCommand(remove subscriber) error: %v", err)
	}

	// Now a send should NOT reach the (removed) subscriber
	send := protocol.NewCommand(
		globals.OBJ_DELIVERY, globals.CMD_SEND,
		"sender", "uid-3", "r3", "", "", "", []byte("ignored"),
	)
	_ = b.HandleCommand(send)

	select {
	case got := <-gotBody:
		t.Fatalf("unexpected delivery after subscriber removal: %q", got)
	case <-time.After(300 * time.Millisecond):
		// no message is success
	}
}

func TestBroker_UpdateGlobals_InvokesUpdateListener(t *testing.T) {
	fs := &fakeServer{addr: "0.0.0.0", port: 1}
	b := NewBroker(fs)
	b.ManagingServer = fs

	payload := map[string]any{
		"address":           "127.0.0.1",
		"port":              6001,
		"verbosity":         2,
		"print_tree":        true,
		"transform_timeout": "500ms",
	}
	raw, _ := json.Marshal(payload)

	cmd := protocol.NewCommand(
		globals.OBJ_GLOBALS, globals.CMD_UPDATE,
		"sender", "uid-g", "", "", "", "", raw,
	)

	if err := b.HandleCommand(cmd); err != nil {
		t.Fatalf("HandleCommand(update globals) error: %v", err)
	}

	// Address/port changed vs fakeServer → UpdateListener must be called
	if !fs.updated {
		t.Fatal("expected UpdateListener to be invoked after globals update")
	}
}
