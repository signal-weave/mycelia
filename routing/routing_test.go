package routing

import (
	"testing"
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

	// remove that subscriber -> channel becomes empty -> channel removed -> route removed
	ch.RemoveSubscriber(Subscriber{Address: "127.0.0.1:7001"})

	// route should be gone (no channels left)
	if brokerHasRoute(b, "r1") {
		t.Fatalf("expected route r1 to be removed after its last channel emptied")
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

	// remove that transformer -> channel becomes empty -> channel removed -> route removed
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
