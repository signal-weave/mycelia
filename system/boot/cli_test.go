package boot

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"mycelia/globals"
)

// snapshot captures the mutable global config so we can restore it after
// each test.
type snapshot struct {
	Address          string
	Port             int
	PrintTree        bool
	Verbosity        int
	TransformTimeout time.Duration
	AutoConsolidate  bool
	DoRecovery       bool
}

func takeSnapshot() snapshot {
	return snapshot{
		Address:          globals.Address,
		Port:             globals.Port,
		PrintTree:        globals.PrintTree,
		Verbosity:        globals.Verbosity,
		TransformTimeout: globals.TransformTimeout,
		AutoConsolidate:  globals.AutoConsolidate,
	}
}

func restoreSnapshot(s snapshot) {
	globals.Address = s.Address
	globals.Port = s.Port
	globals.PrintTree = s.PrintTree
	globals.Verbosity = s.Verbosity
	globals.TransformTimeout = s.TransformTimeout
	globals.AutoConsolidate = s.AutoConsolidate
}

func TestParseRuntimeArgs_ValidFlags(t *testing.T) {
	t.Cleanup(func() { restoreSnapshot(takeSnapshot()) })
	// Set known baselines first.
	globals.Address = "127.0.0.1"
	globals.Port = 5000
	globals.PrintTree = false
	globals.Verbosity = 0
	globals.TransformTimeout = 5 * time.Second
	globals.AutoConsolidate = false

	args := []string{
		"-address", "0.0.0.0",
		"-port", "8088",
		"-verbosity", "3",
		"-print-tree",
		"-xform-timeout", "45s",
		"-consolidate",
	}
	if err := parseRuntimeArgs(args); err != nil {
		t.Fatalf("parseRuntimeArgs returned error: %v", err)
	}

	if globals.Address != "0.0.0.0" {
		t.Errorf("Address = %q, want %q", globals.Address, "0.0.0.0")
	}
	if globals.Port != 8088 {
		t.Errorf("Port = %d, want %d", globals.Port, 8088)
	}
	if !globals.PrintTree {
		t.Errorf("PrintTree = %v, want %v", globals.PrintTree, true)
	}
	if globals.Verbosity != 3 {
		t.Errorf("Verbosity = %d, want %d", globals.Verbosity, 3)
	}
	if globals.TransformTimeout != 45*time.Second {
		t.Errorf(
			"TransformTimeout = %v, want %v",
			globals.TransformTimeout, 45*time.Second,
		)
	}
	if !globals.AutoConsolidate {
		t.Errorf("AutoConsolidate = %v, want %v", globals.AutoConsolidate, true)
	}
}

func TestParseRuntimeArgs_AllowsHostname(t *testing.T) {
	t.Cleanup(func() { restoreSnapshot(takeSnapshot()) })
	globals.Address = "localhost" // start with something valid
	globals.Port = 5000
	globals.TransformTimeout = 5 * time.Second

	err := parseRuntimeArgs([]string{"-address", "example.internal"})
	if err != nil {
		t.Fatalf("unexpected error for hostname: %v", err)
	}
	if globals.Address != "example.internal" {
		t.Errorf("Address = %q, want %q", globals.Address, "example.internal")
	}
}

func TestParseRuntimeArgs_InvalidPort(t *testing.T) {
	t.Cleanup(func() { restoreSnapshot(takeSnapshot()) })
	globals.TransformTimeout = 5 * time.Second // valid timeout

	err := parseRuntimeArgs([]string{"-port", "0"})
	if err == nil {
		t.Fatalf("expected error for invalid port, got nil")
	}
}

func TestParseRuntimeArgs_InvalidIP(t *testing.T) {
	t.Cleanup(func() { restoreSnapshot(takeSnapshot()) })
	globals.TransformTimeout = 5 * time.Second // valid timeout

	err := parseRuntimeArgs([]string{"-address", "256.0.0.1"})
	if err == nil {
		t.Fatalf("expected error for invalid IP, got nil")
	}
}

func TestParseRuntimeArgs_NonPositiveTimeout(t *testing.T) {
	t.Cleanup(func() { restoreSnapshot(takeSnapshot()) })
	err := parseRuntimeArgs([]string{"-xform-timeout", "0s"})
	if err == nil {
		t.Fatalf("expected error for non-positive timeout, got nil")
	}
}

func TestIsValidHostname(t *testing.T) {
	tooLong := strings.Repeat("a", 64)
	tests := []struct {
		in   string
		want bool
	}{
		// Simple valid hostnames
		{"localhost", true},
		{"example.com", true},
		{"sub.domain.org", true},
		{"my-host", true},
		{"a", true},             // single char
		{"123domain.com", true}, // digits allowed if not start/end with '-'

		// Invalid cases
		{"", false},                             // empty
		{"-badstart.com", false},                // leading dash
		{"badend-.com", false},                  // trailing dash
		{fmt.Sprintf("%s.com", tooLong), false}, // >63 chars label
		{"a..b.com", false},                     // empty label
		{"bad!char.com", false},                 // illegal character
		{string(make([]byte, 254)), false},      // too long overall (>253)
	}

	for _, tt := range tests {
		got := isValidHostname(tt.in)
		if got != tt.want {
			t.Errorf("isValidHostname(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
