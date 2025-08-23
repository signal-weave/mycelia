package boot

import (
	"reflect"
	"testing"
	"time"
)

// ------Command Line Args------------------------------------------------------

func TestParseRuntimeArgs_Defaults(t *testing.T) {
	// No flags -> defaults
	cfg, err := parseRuntimeArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := runtimeConfig{
		Address:          "127.0.0.1",
		Port:             5000,
		Verbosity:        0,
		PrintTree:        false,
		TransformTimeout: 5 * time.Second,
	}

	if !reflect.DeepEqual(cfg, want) {
		t.Fatalf("defaults mismatch:\n got: %#v\nwant: %#v", cfg, want)
	}
}

func TestParseRuntimeArgs_WithFlags(t *testing.T) {
	args := []string{
		"-address", "0.0.0.0",
		"-port", "8080",
		"-verbosity", "2",
		"-print-tree",
		"-xform-timeout", "45s",
	}

	cfg, err := parseRuntimeArgs(args)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Address != "0.0.0.0" {
		t.Errorf("Address = %q, want %q", cfg.Address, "0.0.0.0")
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want %d", cfg.Port, 8080)
	}
	if cfg.Verbosity != 2 {
		t.Errorf("Verbosity = %d, want %d", cfg.Verbosity, 2)
	}
	if !cfg.PrintTree {
		t.Errorf("PrintTree = %v, want %v", cfg.PrintTree, true)
	}
	if cfg.TransformTimeout != 45*time.Second {
		t.Errorf("TransformTimeout = %v, want %v", cfg.TransformTimeout, 45*time.Second)
	}
}

func TestParseRuntimeArgs_HostnameAllowed(t *testing.T) {
	// Hostnames are allowed; only reject strings that *look like* IPs but aren't valid IPs.
	args := []string{"-address", "myhost.local"}

	_, err := parseRuntimeArgs(args)
	if err != nil {
		t.Fatalf("expected hostname to be accepted, got error: %v", err)
	}
}

func TestParseRuntimeArgs_InvalidIP(t *testing.T) {
	// "999.168.1.1" looks like an IP, but is invalid; should be rejected.
	args := []string{"-address", "999.168.1.1"}

	_, err := parseRuntimeArgs(args)
	if err == nil {
		t.Fatalf("expected error for invalid IP, got nil")
	}
}

func TestValidateRuntimeConfig_PortBounds(t *testing.T) {
	t.Run("too low", func(t *testing.T) {
		cfg := defaultRuntimeConfig()
		cfg.Port = 0
		if err := validateRuntimeConfig(cfg); err == nil {
			t.Fatalf("expected error for port=0, got nil")
		}
	})

	t.Run("too high", func(t *testing.T) {
		cfg := defaultRuntimeConfig()
		cfg.Port = 70000
		if err := validateRuntimeConfig(cfg); err == nil {
			t.Fatalf("expected error for port=70000, got nil")
		}
	})

	t.Run("ok", func(t *testing.T) {
		cfg := defaultRuntimeConfig()
		cfg.Port = 65535
		if err := validateRuntimeConfig(cfg); err != nil {
			t.Fatalf("unexpected error for port=65535: %v", err)
		}
	})
}

func TestValidateRuntimeConfig_TransformTimeout(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		cfg := defaultRuntimeConfig()
		cfg.TransformTimeout = 0
		if err := validateRuntimeConfig(cfg); err == nil {
			t.Fatalf("expected error for zero timeout, got nil")
		}
	})

	t.Run("negative", func(t *testing.T) {
		cfg := defaultRuntimeConfig()
		cfg.TransformTimeout = -1 * time.Second
		if err := validateRuntimeConfig(cfg); err == nil {
			t.Fatalf("expected error for negative timeout, got nil")
		}
	})

	t.Run("positive ok", func(t *testing.T) {
		cfg := defaultRuntimeConfig()
		cfg.TransformTimeout = 2500 * time.Millisecond
		if err := validateRuntimeConfig(cfg); err != nil {
			t.Fatalf("unexpected error for valid timeout: %v", err)
		}
	})
}

func TestLooksLikeIP(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"127.0.0.1", true},
		{"192.168.1.10", true},
		{"...1", true},            // has 4 parts and empties -> treated as IP-like by current logic
		{"1..2.3", true},          // empty part -> true
		{"abc.def.ghi.jkl", true}, // non-numeric parts but 4 segments -> true
		{"localhost", false},      // not 4 parts
		{"myhost.local", false},   // not 4 parts
		{"10.0.0", false},         // only 3 parts
		{"10.0.0.1.5", false},     // 5 parts
	}

	for _, tt := range tests {
		got := looksLikeIP(tt.in)
		if got != tt.want {
			t.Errorf("looksLikeIP(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestParseRuntimeArgs_UpdatesGlobalViaWrapper(t *testing.T) {
	// Ensure the public wrapper updates RuntimeCfg.
	orig := RuntimeCfg
	defer func() { RuntimeCfg = orig }()

	args := []string{
		"-address", "0.0.0.0",
		"-port", "6001",
		"-verbosity", "3",
		"-xform-timeout", "3s",
	}

	if err := ParseRuntimeArgs(args); err != nil {
		t.Fatalf("ParseRuntimeArgs returned error: %v", err)
	}

	if RuntimeCfg.Address != "0.0.0.0" ||
		RuntimeCfg.Port != 6001 ||
		RuntimeCfg.Verbosity != 3 ||
		RuntimeCfg.TransformTimeout != 3*time.Second {
		t.Fatalf("RuntimeCfg not updated as expected: %#v", RuntimeCfg)
	}
}
