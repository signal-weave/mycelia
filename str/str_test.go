package str_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"mycelia/str"
)

// captureOutput captures stdout during a function's execution.
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

// setVerbosity sets the VERBOSITY environment variable.
func setVerbosity(level string) {
	_ = os.Setenv("VERBOSITY", level)
}

func TestSprintfLn(t *testing.T) {
	t.Run("formats and prints message with args", func(t *testing.T) {
		out := captureOutput(func() {
			str.SprintfLn("Hello %s!", "world")
		})
		if !strings.Contains(out, "Hello world!") {
			t.Errorf("Expected output to contain 'Hello world!', got: %s", out)
		}
	})
}

func TestActionPrint(t *testing.T) {
	t.Run("prints when verbosity is ACTION", func(t *testing.T) {
		setVerbosity("ACTION")
		out := captureOutput(func() {
			str.ActionPrint("doing something")
		})
		if !strings.Contains(out, "[ACTION] - doing something") {
			t.Errorf("Expected action output, got: %s", out)
		}
	})

	t.Run("suppresses output when verbosity is WARNING", func(t *testing.T) {
		setVerbosity("WARNING")
		out := captureOutput(func() {
			str.ActionPrint("doing something")
		})
		if out != "" {
			t.Errorf("Expected no output at WARNING level, got: %s", out)
		}
	})
}

func TestWarningPrint(t *testing.T) {
	t.Run("prints when verbosity is WARNING", func(t *testing.T) {
		setVerbosity("WARNING")
		out := captureOutput(func() {
			str.WarningPrint("be careful")
		})
		if !strings.Contains(out, "[WARNING] - be careful") {
			t.Errorf("Expected warning output, got: %s", out)
		}
	})

	t.Run("suppresses output when verbosity is ERROR", func(t *testing.T) {
		setVerbosity("ERROR")
		out := captureOutput(func() {
			str.WarningPrint("be careful")
		})
		if out != "" {
			t.Errorf("Expected no output at ERROR level, got: %s", out)
		}
	})
}

func TestErrorPrint(t *testing.T) {
	t.Run("prints when verbosity is ERROR", func(t *testing.T) {
		setVerbosity("ERROR")
		out := captureOutput(func() {
			str.ErrorPrint("failure")
		})
		if !strings.Contains(out, "[ERROR] - failure") {
			t.Errorf("Expected error output, got: %s", out)
		}
	})

	t.Run("suppresses output when verbosity is NONE", func(t *testing.T) {
		setVerbosity("NONE")
		out := captureOutput(func() {
			str.ErrorPrint("failure")
		})
		if out != "" {
			t.Errorf("Expected no output at NONE level, got: %s", out)
		}
	})
}

func TestDebugPrintLn(t *testing.T) {
	t.Run("always prints debug messages", func(t *testing.T) {
		setVerbosity("NONE") // Should not affect debug output
		out := captureOutput(func() {
			str.DebugPrintLn("checking stuff")
		})
		if !strings.Contains(out, "[DEBUG] - checking stuff") {
			t.Errorf("Expected debug output, got: %s", out)
		}
	})
}
