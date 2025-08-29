package str

import (
	"bytes"
	"io"
	"mycelia/global"
	"os"
	"strings"
	"testing"
)

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	// Redirect stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	fn() // Run the function

	// Restore and read
	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}

func TestSprintfLn(t *testing.T) {
	out := captureOutput(t, func() {
		SprintfLn("%s-%s-%s", "a", "b", "c")
	})
	want := "a-b-c\n"
	if out != want {
		t.Fatalf("SprintfLn output = %q, want %q", out, want)
	}
}

func TestActionWarningErrorPrint_VerbosityThresholds(t *testing.T) {
	type tc struct {
		name        string
		verbosity   int
		call        func()
		shouldPrint bool
		prefix      string
	}

	tests := []tc{
		{"ActionPrint_v3_prints", 3, func() { ActionPrint("go!") }, true, "[ACTION] - "},
		{"ActionPrint_v2_suppressed", 2, func() { ActionPrint("nope") }, false, ""},

		{"WarningPrint_v2_prints", 2, func() { WarningPrint("heads up") }, true, "[WARNING] - "},
		{"WarningPrint_v1_suppressed", 1, func() { WarningPrint("nope") }, false, ""},

		{"ErrorPrint_v1_prints", 1, func() { ErrorPrint("uh oh") }, true, "[ERROR] - "},
		{"ErrorPrint_v0_suppressed", 0, func() { ErrorPrint("nope") }, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			global.Verbosity = tt.verbosity
			out := captureOutput(t, tt.call)
			if tt.shouldPrint {
				if out == "" {
					t.Fatalf("expected output, got empty")
				}
				if !strings.HasPrefix(out, tt.prefix) {
					t.Fatalf("output %q does not start with %q", out, tt.prefix)
				}
			} else {
				if out != "" {
					t.Fatalf("expected no output, got %q", out)
				}
			}
		})
	}
}

func TestDebugPrintLn(t *testing.T) {
	out := captureOutput(t, func() {
		DebugPrintLn("trace msg")
	})
	want := "[DEBUG] - trace msg\n"
	if out != want {
		t.Fatalf("DebugPrintLn output = %q, want %q", out, want)
	}
}

func TestPrettyPrintStrKeyJson_Success(t *testing.T) {
	data := map[string]any{
		"a": 1,
		"b": "x",
	}
	out := captureOutput(t, func() {
		PrettyPrintStrKeyJson(data)
	})
	// json.MarshalIndent sorts keys for maps, so this is stable.
	want := "{\n    \"a\": 1,\n    \"b\": \"x\"\n}\n"
	if out != want {
		t.Fatalf("PrettyPrintStrKeyJson output = %q, want %q", out, want)
	}
}

func TestPrettyPrintStrKeyJson_Error(t *testing.T) {
	// Use an unsupported JSON type to force marshal error (e.g., channel).
	data := map[string]any{
		"bad": make(chan int),
	}
	out := captureOutput(t, func() {
		PrettyPrintStrKeyJson(data)
	})
	want := "Could not pretty print json data\n"
	if out != want {
		t.Fatalf("PrettyPrintStrKeyJson error output = %q, want %q", out, want)
	}
}
