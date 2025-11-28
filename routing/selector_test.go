package routing

import (
	"testing"

	"mycelia/globals"
)

func TestPubSubSelector_AllSubscribersReturned(t *testing.T) {
	subs := mkSubs(3)
	ps := &pubSubSelector{
		strategy: globals.SelStratPubSub,
		loadFn:   func() []subscriber { return subs },
	}

	got := ps.Select()
	if len(got) != len(subs) {
		t.Fatalf("got %d subs, want %d", len(got), len(subs))
	}

	for i := range subs {
		if got[i].Address != subs[i].Address {
			t.Fatalf("index %d: got %q, want %q", i, got[i].Address, subs[i].Address)
		}
	}
}

func TestRoundRobinSelector_CyclesOrder(t *testing.T) {
	subs := mkSubs(3)
	rr := &roundRobinSelector{
		strategy: globals.SelStratRoundRobin,
		last:     -1,
		loadFn:   func() []subscriber { return subs },
	}

	wantIdx := []int{0, 1, 2, 0, 1}
	for i, w := range wantIdx {
		got := rr.Select()
		if len(got) != 1 {
			t.Fatalf("step %d: expected 1 sub, got %d", i, len(got))
		}
		if got[0].Address != subs[w].Address {
			t.Fatalf("step %d: got %q, want %q", i, got[0].Address, subs[w].Address)
		}
	}
}

func TestRoundRobinSelector_EmptyReturnsNil(t *testing.T) {
	rr := &roundRobinSelector{
		strategy: globals.SelStratRoundRobin,
		last:     -1,
		loadFn:   func() []subscriber { return nil },
	}
	if got := rr.Select(); got != nil {
		t.Fatalf("expected nil on empty, got %v", got)
	}
}

func TestRandomSelector_UsesChooseFn(t *testing.T) {
	subs := mkSubs(4)

	// deterministic chooser picks index 2
	choose := func(s []subscriber) (subscriber, bool) {
		if len(s) == 0 {
			return subscriber{}, false
		}
		return s[2], true
	}

	rs := &randomSelector{
		strategy: globals.SelStratRandom,
		loadFn:   func() []subscriber { return subs },
		chooseFn: choose,
	}

	got := rs.Select()
	if len(got) != 1 {
		t.Fatalf("expected 1 sub, got %d", len(got))
	}
	if got[0].Address != subs[2].Address {
		t.Fatalf("got %q, want %q", got[0].Address, subs[2].Address)
	}
}

func TestRandomSelector_EmptyReturnsEmptySlice(t *testing.T) {
	rs := &randomSelector{
		strategy: globals.SelStratRandom,
		loadFn:   func() []subscriber { return nil },
		chooseFn: func(s []subscriber) (subscriber, bool) { return subscriber{}, false },
	}
	got := rs.Select()
	if got == nil || len(got) != 0 {
		t.Fatalf("expected empty slice, got %#v", got)
	}
}

func TestGetStrategyName(t *testing.T) {
	ps := &pubSubSelector{
		strategy: globals.SelStratPubSub,
		loadFn: func() []subscriber {
			return nil
		},
	}
	if ps.GetStrategyName() != "pub-sub" {
		t.Fatal("expected 'pub-sub' selection strategy name")
	}

	rr := &roundRobinSelector{
		strategy: globals.SelStratRoundRobin,
		loadFn: func() []subscriber {
			return nil
		},
	}
	if rr.GetStrategyName() != "round-robin" {
		t.Fatal("expected 'round-robin' selection strategy name")
	}

	rs := &randomSelector{
		strategy: globals.SelStratRandom,
		loadFn:   func() []subscriber { return nil },
		chooseFn: func(s []subscriber) (subscriber, bool) {
			return subscriber{}, false
		},
	}
	if rs.GetStrategyName() != "random" {
		t.Fatal("expected 'random' selection strategy name")
	}
}
