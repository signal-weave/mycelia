package boot

import (
	"os"
	"testing"
	"time"

	"mycelia/commands"
)

// Test that parseRuntimeConfigurable maps all supported runtime fields
// and updates the VERBOSITY environment variable.
func TestParseRuntimeConfigurable_SetsFieldsAndEnv(t *testing.T) {
	// Save/restore env
	prevVerbosity := os.Getenv("VERBOSITY")
	t.Cleanup(func() { _ = os.Setenv("VERBOSITY", prevVerbosity) })

	cfg := &runtimeConfig{
		Address:          "",
		Port:             0,
		Verbosity:        0,
		PrintTree:        false,
		TransformTimeout: 0,
	}

	input := map[string]any{
		"runtime": map[string]any{
			"address":       "0.0.0.0",
			"port":          8080, // will survive via json round-trip
			"verbosity":     2,    // ditto
			"print-tree":    true,
			"xform-timeout": "45s",
		},
	}

	parseRuntimeConfigurable(cfg, input)

	if cfg.Address != "0.0.0.0" {
		t.Fatalf("Address = %q, want %q", cfg.Address, "0.0.0.0")
	}
	if cfg.Port != 8080 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 8080)
	}
	if cfg.Verbosity != 2 {
		t.Fatalf("Verbosity = %d, want %d", cfg.Verbosity, 2)
	}
	if !cfg.PrintTree {
		t.Fatalf("PrintTree = %v, want %v", cfg.PrintTree, true)
	}
	if cfg.TransformTimeout != 45*time.Second {
		t.Fatalf("TransformTimeout = %v, want %v", cfg.TransformTimeout, 45*time.Second)
	}
	if got := os.Getenv("VERBOSITY"); got != "2" {
		t.Fatalf("env VERBOSITY = %q, want %q", got, "2")
	}
}

func TestParseRouteCmds_AppendsTransformerAndSubscriberCommands(t *testing.T) {
	// Reset global CommandList and restore afterwards.
	old := CommandList
	CommandList = nil
	t.Cleanup(func() { CommandList = old })

	routeData := []map[string]any{
		{
			"name": "default",
			"channels": []map[string]any{
				{
					"name": "inmem",
					// FIXED: use "transformers" (not "transformeres")
					"transformers": []map[string]any{
						{"address": "127.0.0.1:7010"},
						{"address": "10.0.0.52:8008"},
					},
					"subscribers": []map[string]any{
						{"address": "127.0.0.1:1234"},
						{"address": "16.70.18.1:9999"},
					},
				},
			},
		},
	}

	parseRouteCmds(routeData)

	if len(CommandList) != 4 {
		t.Fatalf("CommandList length = %d, want %d", len(CommandList), 4)
	}

	var nXforms, nSubs int
	for _, cmd := range CommandList {
		switch cmd.(type) {
		case *commands.AddTransformer:
			nXforms++
		case *commands.AddSubscriber:
			nSubs++
		}
	}
	if nXforms != 2 {
		t.Fatalf("AddTransformer count = %d, want %d", nXforms, 2)
	}
	if nSubs != 2 {
		t.Fatalf("AddSubscriber count = %d, want %d", nSubs, 2)
	}
}
