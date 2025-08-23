package boot

import (
	"testing"
)

// resetCommandList clears the global CommandList between tests.
func resetCommandList() {
	CommandList = CommandList[:0]
}

func TestParseRouteCmds_AppendsAddRoute(t *testing.T) {
	resetCommandList()

	routeData := []map[string]any{
		{
			"name":     "r1",
			"channels": []map[string]any{}, // empty is fine; just ensures the key exists
		},
	}

	base := len(CommandList)
	parseRouteCmds(routeData)
	got := len(CommandList) - base

	if got != 1 {
		t.Fatalf("expected 1 command appended (AddRoute), got %d", got)
	}
}

func TestParseChannelCmds_AddsChannelAndTransformers(t *testing.T) {
	resetCommandList()

	// One channel with one transformer.
	chanData := []map[string]any{
		{
			"name": "ch1",
			"transformers": []map[string]string{
				{"address": "127.0.0.1:7001"},
			},
			// NOTE: no subscribers in this test
		},
	}

	base := len(CommandList)
	parseChannelCmds("r1", chanData)
	got := len(CommandList) - base

	// Expect: AddChannel + AddTransformer = 2
	if got != 2 {
		t.Fatalf("expected 2 commands appended (AddChannel + AddTransformer), got %d", got)
	}
}

func TestParseChannelCmds_SubscribersKeyTypo_NoSubscribersAdded(t *testing.T) {
	resetCommandList()

	// Intentional: use the CORRECT English key "subscribers".
	// The implementation looks for "subscribres" (typo), so NO subscriber commands should be added.
	chanData := []map[string]any{
		{
			"name": "ch1",
			"subscribers": []map[string]string{
				{"address": "127.0.0.1:9001"},
			},
		},
	}

	base := len(CommandList)
	parseChannelCmds("r1", chanData)
	got := len(CommandList) - base

	// Only AddChannel should be appended (no subscribers due to key mismatch).
	if got != 1 {
		t.Fatalf("expected 1 command appended (AddChannel only), got %d", got)
	}
}

func TestParseChannelCmds_SubscribresKey_AddsSubscribers(t *testing.T) {
	resetCommandList()

	// Use the exact key that the implementation expects: "subscribres" (typo).
	chanData := []map[string]any{
		{
			"name": "ch1",
			"subscribres": []map[string]string{
				{"address": "127.0.0.1:9001"},
				{"address": "127.0.0.1:9002"},
			},
		},
	}

	base := len(CommandList)
	parseChannelCmds("r1", chanData)
	got := len(CommandList) - base

	// Expect: AddChannel + 2*AddSubscriber = 3
	if got != 3 {
		t.Fatalf("expected 3 commands appended (AddChannel + 2 AddSubscriber), got %d", got)
	}
}

func TestParseXformCmds_AppendsOnePerTransformer(t *testing.T) {
	resetCommandList()

	xforms := []map[string]any{
		{"address": "127.0.0.1:7100"},
		{"address": "127.0.0.1:7101"},
		{"address": "127.0.0.1:7102"},
	}

	base := len(CommandList)
	parseXformCmds("rX", "cX", xforms)
	got := len(CommandList) - base

	if got != len(xforms) {
		t.Fatalf("expected %d AddTransformer commands, got %d", len(xforms), got)
	}
}

func TestParseSubscriberCmds_AppendsOnePerSubscriber(t *testing.T) {
	resetCommandList()

	subs := []map[string]any{
		{"address": "127.0.0.1:9200"},
		{"address": "127.0.0.1:9201"},
	}

	base := len(CommandList)
	parseSubscriberCmds("rY", "cY", subs)
	got := len(CommandList) - base

	if got != len(subs) {
		t.Fatalf("expected %d AddSubscriber commands, got %d", len(subs), got)
	}
}
