package boot

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"mycelia/global"
)

// snapshot captures the mutable global config so we can restore it after
// each test.
type snapshot struct {
	Address          string
	Port             int
	PrintTree        bool
	Verbosity        int
	TransformTimeout time.Duration
}

func takeSnapshot() snapshot {
	return snapshot{
		Address:          global.Address,
		Port:             global.Port,
		PrintTree:        global.PrintTree,
		Verbosity:        global.Verbosity,
		TransformTimeout: global.TransformTimeout,
	}
}

func restoreSnapshot(s snapshot) {
	global.Address = s.Address
	global.Port = s.Port
	global.PrintTree = s.PrintTree
	global.Verbosity = s.Verbosity
	global.TransformTimeout = s.TransformTimeout
}

func TestParseRuntimeArgs_ValidFlags(t *testing.T) {
	t.Cleanup(func() { restoreSnapshot(takeSnapshot()) })
	// Set known baselines first.
	global.Address = "127.0.0.1"
	global.Port = 5000
	global.PrintTree = false
	global.Verbosity = 0
	global.TransformTimeout = 5 * time.Second

	args := []string{
		"-address", "0.0.0.0",
		"-port", "8088",
		"-verbosity", "3",
		"-print-tree",
		"-xform-timeout", "45s",
	}
	if err := parseRuntimeArgs(args); err != nil {
		t.Fatalf("parseRuntimeArgs returned error: %v", err)
	}

	if global.Address != "0.0.0.0" {
		t.Errorf("Address = %q, want %q", global.Address, "0.0.0.0")
	}
	if global.Port != 8088 {
		t.Errorf("Port = %d, want %d", global.Port, 8088)
	}
	if !global.PrintTree {
		t.Errorf("PrintTree = %v, want %v", global.PrintTree, true)
	}
	if global.Verbosity != 3 {
		t.Errorf("Verbosity = %d, want %d", global.Verbosity, 3)
	}
	if global.TransformTimeout != 45*time.Second {
		t.Errorf(
			"TransformTimeout = %v, want %v",
			global.TransformTimeout, 45*time.Second,
		)
	}
}

func TestParseRuntimeArgs_AllowsHostname(t *testing.T) {
	t.Cleanup(func() { restoreSnapshot(takeSnapshot()) })
	global.Address = "localhost" // start with something valid
	global.Port = 5000
	global.TransformTimeout = 5 * time.Second

	err := parseRuntimeArgs([]string{"-address", "example.internal"})
	if err != nil {
		t.Fatalf("unexpected error for hostname: %v", err)
	}
	if global.Address != "example.internal" {
		t.Errorf("Address = %q, want %q", global.Address, "example.internal")
	}
}

func TestParseRuntimeArgs_InvalidPort(t *testing.T) {
	t.Cleanup(func() { restoreSnapshot(takeSnapshot()) })
	global.TransformTimeout = 5 * time.Second // valid timeout

	err := parseRuntimeArgs([]string{"-port", "0"})
	if err == nil {
		t.Fatalf("expected error for invalid port, got nil")
	}
}

func TestParseRuntimeArgs_InvalidIP(t *testing.T) {
	t.Cleanup(func() { restoreSnapshot(takeSnapshot()) })
	global.TransformTimeout = 5 * time.Second // valid timeout

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
