package str

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"mycelia/boot"
)

func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	orig := os.Stdout
	os.Stdout = w
	defer func() {
		_ = w.Close()
		os.Stdout = orig
	}()

	f()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}

func withVerbosity(v int, f func()) {
	orig := boot.RuntimeCfg
	boot.RuntimeCfg.Verbosity = v
	defer func() { boot.RuntimeCfg = orig }()
	f()
}

func TestSprintfLn(t *testing.T) {
	out := captureOutput(func() {
		SprintfLn("hello %s %s", "a", "b")
	})
	if out != "hello a b\n" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestActionPrint_VerbosityGate(t *testing.T) {
	// verbosity < 3 => no output
	withVerbosity(2, func() {
		out := captureOutput(func() {
			ActionPrint("did something")
		})
		if out != "" {
			t.Fatalf("expected no output at verbosity=2, got: %q", out)
		}
	})

	// verbosity >= 3 => prints
	withVerbosity(3, func() {
		out := captureOutput(func() {
			ActionPrint("did something")
		})
		if strings.TrimSpace(out) != "[ACTION] - did something" {
			t.Fatalf("unexpected output at verbosity=3: %q", out)
		}
	})
}

func TestWarningPrint_VerbosityGate(t *testing.T) {
	// verbosity < 2 => no output
	withVerbosity(1, func() {
		out := captureOutput(func() {
			WarningPrint("careful")
		})
		if out != "" {
			t.Fatalf("expected no output at verbosity=1, got: %q", out)
		}
	})

	// verbosity >= 2 => prints
	withVerbosity(2, func() {
		out := captureOutput(func() {
			WarningPrint("careful")
		})
		if strings.TrimSpace(out) != "[WARNING] - careful" {
			t.Fatalf("unexpected output at verbosity=2: %q", out)
		}
	})
}

func TestErrorPrint_VerbosityGate(t *testing.T) {
	// verbosity < 1 => no output
	withVerbosity(0, func() {
		out := captureOutput(func() {
			ErrorPrint("oh no")
		})
		if out != "" {
			t.Fatalf("expected no output at verbosity=0, got: %q", out)
		}
	})

	// verbosity >= 1 => prints
	withVerbosity(1, func() {
		out := captureOutput(func() {
			ErrorPrint("oh no")
		})
		if strings.TrimSpace(out) != "[ERROR] - oh no" {
			t.Fatalf("unexpected output at verbosity=1: %q", out)
		}
	})
}

func TestDebugPrintLn_AlwaysPrints(t *testing.T) {
	out := captureOutput(func() {
		DebugPrintLn("details")
	})
	if strings.TrimSpace(out) != "[DEBUG] - details" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestPrettyPrintStrKeyJson_SingleKey(t *testing.T) {
	// Use a single-key map so key-order isn’t an issue.
	payload := map[string]any{"a": 1}

	out := captureOutput(func() {
		PrettyPrintStrKeyJson(payload)
	})

	// Expect pretty JSON with 4-space indent and trailing newline (fmt.Println).
	expected := "{\n    \"a\": 1\n}\n"
	if out != expected {
		t.Fatalf("unexpected pretty JSON output:\n--- got ---\n%s--- want ---\n%s", out, expected)
	}
}
