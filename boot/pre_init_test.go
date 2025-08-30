package boot

import (
	"testing"
	"time"

	"mycelia/globals"
	"mycelia/protocol"
)

func TestParseRuntimeConfigurable_UpdatesGlobals(t *testing.T) {
	// Preserve and restore original globals to avoid test bleed.
	oldAddr := globals.Address
	oldPort := globals.Port
	oldVerb := globals.Verbosity
	oldPrint := globals.PrintTree
	oldXform := globals.TransformTimeout
	oldAutoCon := globals.AutoConsolidate
	t.Cleanup(func() {
		globals.Address = oldAddr
		globals.Port = oldPort
		globals.Verbosity = oldVerb
		globals.PrintTree = oldPrint
		globals.TransformTimeout = oldXform
		globals.AutoConsolidate = oldAutoCon
	})

	data := map[string]any{
		"runtime": map[string]any{
			"address":       "127.0.0.1",
			"port":          6001,
			"verbosity":     3,
			"print-tree":    true,
			"xform-timeout": "150ms",
			"consolidate": false,
		},
	}

	parseRuntimeConfigurable(data)

	if globals.Address != "127.0.0.1" {
		t.Fatalf("Address not updated: %q", globals.Address)
	}
	if globals.Port != 6001 {
		t.Fatalf("Port not updated: %d", globals.Port)
	}
	if globals.Verbosity != 3 {
		t.Fatalf("Verbosity not updated: %d", globals.Verbosity)
	}
	if globals.PrintTree != true {
		t.Fatalf("PrintTree not updated: %v", globals.PrintTree)
	}
	if globals.TransformTimeout != 150*time.Millisecond {
		t.Fatalf("TransformTimeout not updated: %v", globals.TransformTimeout)
	}
	if globals.AutoConsolidate != false {
		t.Fatalf("AutoConsolidation no updated: %v", globals.AutoConsolidate)
	}

}

func TestParseRouteCmds_GeneratesCommands(t *testing.T) {
	// Start from a clean command list.
	CommandList = nil

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

	// Expect 4 commands: 2 transformers + 2 subscribers
	if len(CommandList) != 4 {
		t.Fatalf("expected 4 commands, got %d", len(CommandList))
	}

	// Helper to read fields regardless of pointer/value slice element.
	get := func(i int) *protocol.Command {
		switch c := any(CommandList[i]).(type) {
		case *protocol.Command:
			return c
		case protocol.Command:
			return &c
		default:
			t.Fatalf("unexpected CommandList element type at %d", i)
			return nil
		}
	}

	// First two should be transformers (in the order provided)
	c0 := get(0)
	if c0.ObjType != globals.OBJ_TRANSFORMER || c0.CmdType != globals.CMD_ADD {
		t.Fatalf("cmd0 wrong types: obj=%d cmd=%d", c0.ObjType, c0.CmdType)
	}
	if c0.Arg1 != "default" || c0.Arg2 != "inmem" || c0.Arg3 != "127.0.0.1:7010" {
		t.Fatalf("cmd0 args wrong: %q %q %q", c0.Arg1, c0.Arg2, c0.Arg3)
	}
	if c0.Sender != "127.0.0.1:7010" {
		t.Fatalf("cmd0 sender wrong: %q", c0.Sender)
	}

	c1 := get(1)
	if c1.ObjType != globals.OBJ_TRANSFORMER || c1.CmdType != globals.CMD_ADD {
		t.Fatalf("cmd1 wrong types: obj=%d cmd=%d", c1.ObjType, c1.CmdType)
	}
	if c1.Arg3 != "10.0.0.52:8008" {
		t.Fatalf("cmd1 address wrong: %q", c1.Arg3)
	}

	// Next two should be subscribers
	c2 := get(2)
	if c2.ObjType != globals.OBJ_SUBSCRIBER || c2.CmdType != globals.CMD_ADD {
		t.Fatalf("cmd2 wrong types: obj=%d cmd=%d", c2.ObjType, c2.CmdType)
	}
	if c2.Arg1 != "default" || c2.Arg2 != "inmem" || c2.Arg3 != "127.0.0.1:1234" {
		t.Fatalf("cmd2 args wrong: %q %q %q", c2.Arg1, c2.Arg2, c2.Arg3)
	}

	c3 := get(3)
	if c3.ObjType != globals.OBJ_SUBSCRIBER || c3.CmdType != globals.CMD_ADD {
		t.Fatalf("cmd3 wrong types: obj=%d cmd=%d", c3.ObjType, c3.CmdType)
	}
	if c3.Arg3 != "16.70.18.1:9999" {
		t.Fatalf("cmd3 address wrong: %q", c3.Arg3)
	}
}

func TestParseRouteCmds_NoChannels_NoCommands(t *testing.T) {
	CommandList = nil

	routeData := []map[string]any{
		{
			"name":     "empty",
			"channels": []any{}, // no channels
		},
	}
	parseRouteCmds(routeData)

	if len(CommandList) != 0 {
		t.Fatalf(
			"expected no commands for empty channels, got %d", len(CommandList),
		)
	}
}
