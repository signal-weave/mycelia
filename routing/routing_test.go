package routing

import (
	"testing"
	"time"

	"mycelia/commands"
	"mycelia/global"
	"mycelia/test"
)

// helper: check if a route exists on the broker (with lock)
func brokerHasRoute(b *Broker, name string) bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	_, ok := b.routes[name]
	return ok
}

// helper: check if a channel exists on the route (with lock)
func routeHasChannel(r *Route, chName string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	_, ok := r.channels[chName]
	return ok
}

func TestRouteHasBrokerBackPointer(t *testing.T) {
	b := NewBroker()
	r := b.Route("r1")
	if r.broker == nil {
		t.Fatalf("expected route.broker to be set; got nil")
	}
}

func TestRemoveOnlySubscriberRemovesChannelAndRoute(t *testing.T) {
	b := NewBroker()
	r := b.Route("r1")
	ch := r.Channel("c1")

	// add one subscriber
	ch.AddSubscriber(Subscriber{Address: "127.0.0.1:7001"})

	// sanity: route & channel exist
	if !brokerHasRoute(b, "r1") {
		t.Fatalf("expected route r1 to exist on broker")
	}
	if !routeHasChannel(r, "c1") {
		t.Fatalf("expected channel c1 to exist on route r1")
	}

	// remove that subscriber -> channel becomes empty
	// -> channel removed -> route removed
	ch.RemoveSubscriber(Subscriber{Address: "127.0.0.1:7001"})

	// route should be gone (no channels left)
	if brokerHasRoute(b, "r1") {
		t.Fatalf(
			"expected route r1 to be removed after its last channel emptied",
		)
	}
}

func TestRemoveOnlyTransformerRemovesChannelAndRoute(t *testing.T) {
	b := NewBroker()
	r := b.Route("r2")
	ch := r.Channel("c2")

	// add one transformer
	ch.AddTransformer(Transformer{Address: "127.0.0.1:8008"})

	// sanity: route & channel exist
	if !brokerHasRoute(b, "r2") {
		t.Fatalf("expected route r2 to exist on broker")
	}
	if !routeHasChannel(r, "c2") {
		t.Fatalf("expected channel c2 to exist on route r2")
	}

	// remove that transformer -> channel becomes empty
	// -> channel removed -> route removed
	ch.RemoveTransformer(Transformer{Address: "127.0.0.1:8008"})

	// route should be gone (no channels left)
	if brokerHasRoute(b, "r2") {
		t.Fatalf("expected route r2 to be removed after its last channel emptied")
	}
}

func TestAddDoesNotDuplicateByAddress(t *testing.T) {
	b := NewBroker()
	r := b.Route("r3")
	ch := r.Channel("c3")

	// subscribers de-dupe by Address
	ch.AddSubscriber(Subscriber{Address: "127.0.0.1:9001"})
	ch.AddSubscriber(Subscriber{Address: "127.0.0.1:9001"})

	// transformers de-dupe by Address
	ch.AddTransformer(Transformer{Address: "10.0.0.1:9100"})
	ch.AddTransformer(Transformer{Address: "10.0.0.1:9100"})

	// verify lengths are 1 each
	ch.mutex.RLock()
	subsCount := len(ch.subscribers)
	xformCount := len(ch.transformers)
	ch.mutex.RUnlock()

	if subsCount != 1 {
		t.Fatalf("expected exactly 1 subscriber, got %d", subsCount)
	}
	if xformCount != 1 {
		t.Fatalf("expected exactly 1 transformer, got %d", xformCount)
	}
}

// tiny helper: read with timeout so tests don't hang
func recvWithTimeout(
	t *testing.T,
	ch <-chan string,
	d time.Duration,
) (string, bool) {
	t.Helper()
	select {
	case v := <-ch:
		return v, true
	case <-time.After(d):
		return "", false
	}
}

func TestSubscriberReceivesDelivery(t *testing.T) {
	// Speed up tests that hit the net
	global.TransformTimeout = 1

	// Start a one-way mock server that captures what we send to it.
	subAddr, gotBody, stopSub := test.MockOneWayServer(t)
	defer stopSub()

	// Wire up route/channel/subscriber
	b := NewBroker()
	r := b.Route("r-sub")
	ch := r.Channel("c-sub")
	ch.AddSubscriber(Subscriber{Address: subAddr})

	// Send a delivery through the route
	payload := "hello-subscriber"
	msg := &commands.Delivery{
		ID:    "d1",
		Route: "r-sub",
		Body:  []byte(payload),
	}

	r.ProcessDelivery(msg)

	// Verify subscriber got the original body
	got, ok := recvWithTimeout(t, gotBody, 500*time.Millisecond)
	if !ok {
		t.Fatalf("timeout: subscriber did not receive body")
	}
	if got != payload {
		t.Fatalf("expected subscriber body %q, got %q", payload, got)
	}

	if msg.Status != commands.StatusResolved {
		t.Fatalf(
			"expected message status to be StatusResolved, got %v", msg.Status,
		)
	}
}

func TestTransformerThenSubscriber_PathTransformsBody(t *testing.T) {
	global.TransformTimeout = 1

	// Transformer echoes back with prefix + body
	const prefix = "tx:"
	txAddr, stopTx := test.MockTwoWayServer(t, prefix)
	defer stopTx()

	// Subscriber captures final body sent
	subAddr, gotBody, stopSub := test.MockOneWayServer(t)
	defer stopSub()

	b := NewBroker()
	r := b.Route("r-tx")
	ch := r.Channel("c-tx")

	// Add transformer then subscriber.
	// Will dial, write, read with timeout.
	ch.AddTransformer(Transformer{Address: txAddr})
	ch.AddSubscriber(Subscriber{Address: subAddr})

	payload := "hello-transformer"
	msg := &commands.Delivery{
		ID:    "d2",
		Route: "r-tx",
		Body:  []byte(payload),
	}

	// Channel runs all transformers (in order), then fans out to subscribers
	r.ProcessDelivery(msg)

	// Expect transformed payload at subscriber
	want := prefix + payload
	got, ok := recvWithTimeout(t, gotBody, 500*time.Millisecond)
	if !ok {
		t.Fatalf("timeout: subscriber did not receive transformed body")
	}
	if got != want {
		t.Fatalf("expected subscriber body %q, got %q", want, got)
	}
}

func TestTransformerFailureFallsBackToOriginalBody(t *testing.T) {
	global.TransformTimeout = 1

	// Use an address that won't accept connections to force a dial error
	// When transformer fails, it returns the original delivery, and the channel
	// continues with the unmodified message.
	badTransformer := "127.0.0.1:1" // privileged/closed port on localhost

	subAddr, gotBody, stopSub := test.MockOneWayServer(t)
	defer stopSub()

	b := NewBroker()
	r := b.Route("r-fail")
	ch := r.Channel("c-fail")
	ch.AddTransformer(Transformer{Address: badTransformer})
	ch.AddSubscriber(Subscriber{Address: subAddr})

	payload := "original-body"
	msg := &commands.Delivery{
		ID:    "d3",
		Route: "r-fail",
		Body:  []byte(payload),
	}

	// transformer fails -> proceed with original body
	r.ProcessDelivery(msg)

	got, ok := recvWithTimeout(t, gotBody, 500*time.Millisecond)
	if !ok {
		t.Fatalf(
			"timeout: subscriber did not receive body despite transformer failure",
		)
	}
	if got != payload {
		t.Fatalf(
			"expected original body %q after transformer failure, got %q",
			payload, got,
		)
	}
}
