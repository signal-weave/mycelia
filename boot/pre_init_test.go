package boot

import (
	"fmt"
	"os"
	"testing"
	"time"

	"mycelia/commands"
	"mycelia/global"
)

// Test that parseRuntimeConfigurable maps all supported runtime fields
// and updates the VERBOSITY environment variable.
func TestParseRuntimeConfigurable_SetsFieldsAndEnv(t *testing.T) {
	// Save/restore env
	prevVerbosity := os.Getenv("VERBOSITY")
	t.Cleanup(func() { _ = os.Setenv("VERBOSITY", prevVerbosity) })

	input := map[string]any{
		"runtime": map[string]any{
			"address":       "0.0.0.0",
			"port":          8080, // will survive via json round-trip
			"verbosity":     2,    // ditto
			"print-tree":    true,
			"xform-timeout": "45s",
		},
	}

	parseRuntimeConfigurable(input)

	if global.Address != "0.0.0.0" {
		t.Fatalf("Address = %q, want %q", global.Address, "0.0.0.0")
	}
	if global.Port != 8080 {
		t.Fatalf("Port = %d, want %d", global.Port, 8080)
	}
	if global.Verbosity != 2 {
		t.Fatalf("Verbosity = %d, want %d", global.Verbosity, 2)
	}
	if !global.PrintTree {
		t.Fatalf("PrintTree = %v, want %v", global.PrintTree, true)
	}
	if global.TransformTimeout != 45*time.Second {
		t.Fatalf(
			"TransformTimeout = %v, want %v",
			global.TransformTimeout, 45*time.Second,
		)
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
			"channels": []any{
				map[string]any{
					"name": "inmem",
					"transformers": []any{
						map[string]any{"address": "127.0.0.1:7010"},
						map[string]any{"address": "10.0.0.52:8008"},
					},
					"subscribers": []any{
						map[string]any{"address": "127.0.0.1:1234"},
						map[string]any{"address": "16.70.18.1:9999"},
					},
				},
			},
		},
	}

	parseRouteCmds(routeData)
	fmt.Println(CommandList)

	var nXforms, nSubs int
	for _, cmd := range CommandList {
		switch cmd.(type) {
		case *commands.Transformer:
			nXforms++
		case *commands.Subscriber:
			nSubs++
		}
	}

	if nXforms != 2 {
		t.Fatalf("Transformer count = %d, want %d", nXforms, 2)
	}
	if nSubs != 2 {
		t.Fatalf("Subscriber count = %d, want %d", nSubs, 2)
	}
}
