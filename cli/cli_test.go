// mycelia/cli/args_test.go
package cli

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestDefaults_NoEnv_NoFlags(t *testing.T) {
	t.Setenv("MYC_ADDR", "")
	t.Setenv("MYC_PORT", "")
	t.Setenv("MYC_VERBOSITY", "")
	t.Setenv("MYC_PRINT_TREE", "")
	t.Setenv("MYC_XFORM_TIMEOUT", "")

	err := ParseRuntimeArgs([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := defaultRuntimeConfig()
	if !reflect.DeepEqual(RuntimeCfg, want) {
		t.Fatalf("cfg mismatch\nhave: %#v\nwant: %#v", RuntimeCfg, want)
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("MYC_ADDR", "0.0.0.0")
	t.Setenv("MYC_PORT", "8081")
	t.Setenv("MYC_VERBOSITY", "3")
	t.Setenv("MYC_PRINT_TREE", "true")
	t.Setenv("MYC_XFORM_TIMEOUT", "45s")

	err := ParseRuntimeArgs([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := RuntimeCfg

	if cfg.Address != "0.0.0.0" {
		t.Errorf("Address = %q, want %q", cfg.Address, "0.0.0.0")
	}
	if cfg.Port != 8081 {
		t.Errorf("Port = %d, want %d", cfg.Port, 8081)
	}
	if cfg.Verbosity != 3 {
		t.Errorf("Verbosity = %d, want %d", cfg.Verbosity, 3)
	}
	if cfg.PrintTree != true {
		t.Errorf("PrintTree = %v, want %v", cfg.PrintTree, true)
	}
	if cfg.TransformTimeout != 45*time.Second {
		t.Errorf("TransformTimeout = %v, want %v", cfg.TransformTimeout, 45*time.Second)
	}
}

func TestFlagOverridesWinOverEnv(t *testing.T) {
	// Set env, then override with flags
	t.Setenv("MYC_ADDR", "1.2.3.4")
	t.Setenv("MYC_PORT", "9999")
	t.Setenv("MYC_VERBOSITY", "1")
	t.Setenv("MYC_PRINT_TREE", "false")
	t.Setenv("MYC_XFORM_TIMEOUT", "10s")

	argv := []string{
		"-addr", "127.0.0.1",
		"-port", "5501",
		"-verbosity", "2",
		"-print-tree",
		"-xform-timeout", "1m",
		"send", "--route", "foo",
	}
	err := ParseRuntimeArgs(argv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := RuntimeCfg

	if cfg.Address != "127.0.0.1" {
		t.Errorf("Address = %q, want %q", cfg.Address, "127.0.0.1")
	}
	if cfg.Port != 5501 {
		t.Errorf("Port = %d, want %d", cfg.Port, 5501)
	}
	if cfg.Verbosity != 2 {
		t.Errorf("Verbosity = %d, want %d", cfg.Verbosity, 2)
	}
	if cfg.PrintTree != true {
		t.Errorf("PrintTree = %v, want %v", cfg.PrintTree, true)
	}
	if cfg.TransformTimeout != time.Minute {
		t.Errorf("TransformTimeout = %v, want %v", cfg.TransformTimeout, time.Minute)
	}
}

func TestValidation_InvalidPort(t *testing.T) {
	err := ParseRuntimeArgs([]string{"-port", "0"})
	if err == nil {
		t.Fatalf("expected error for invalid port, got none (cfg=%#v)",
			RuntimeCfg)
	}
}

func TestValidation_InvalidIPWhenLooksLikeIP(t *testing.T) {
	// 256.1.1.1 parses as "looks like an IP", but net.ParseIP will fail
	err := ParseRuntimeArgs([]string{"-addr", "256.1.1.1"})
	if err == nil {
		t.Fatalf("expected error for invalid IP address, got none")
	}
}

func TestValidation_NegativeTimeout(t *testing.T) {
	err := ParseRuntimeArgs([]string{"-xform-timeout", "-5s"})
	if err == nil {
		t.Fatalf("expected error for non-positive timeout, got none")
	}
}

func TestPrintTreeEnvVariants(t *testing.T) {
	variants := map[string]bool{
		"1":     true,
		"true":  true,
		"yes":   true,
		"on":    true,
		"0":     false,
		"false": false,
		"no":    false,
		"off":   false,
	}
	for val, want := range variants {
		t.Run("MYC_PRINT_TREE="+val, func(t *testing.T) {
			// Clear other env that could interfere
			clearMycEnv(t)
			t.Setenv("MYC_PRINT_TREE", val)

			err := ParseRuntimeArgs([]string{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if RuntimeCfg.PrintTree != want {
				t.Errorf("PrintTree = %v, want %v (env %q)",
					RuntimeCfg.PrintTree, want, val)
			}
		})
	}
}

func TestEnvGarbageIsIgnoredWhereAppropriate(t *testing.T) {
	// Bad ints/durations should simply not override defaults (per your implementation).
	clearMycEnv(t)
	t.Setenv("MYC_PORT", "not-an-int")
	t.Setenv("MYC_VERBOSITY", "NaN")
	t.Setenv("MYC_XFORM_TIMEOUT", "forever-ish")

	err := ParseRuntimeArgs([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := RuntimeCfg
	def := defaultRuntimeConfig()
	if cfg.Port != def.Port {
		t.Errorf("Port = %d, want default %d", cfg.Port, def.Port)
	}
	if cfg.Verbosity != def.Verbosity {
		t.Errorf("Verbosity = %d, want default %d", cfg.Verbosity, def.Verbosity)
	}
	if cfg.TransformTimeout != def.TransformTimeout {
		t.Errorf("TransformTimeout = %v, want default %v", cfg.TransformTimeout, def.TransformTimeout)
	}
}

func clearMycEnv(t *testing.T) {
	t.Helper()
	unset := []string{
		"MYC_ADDR",
		"MYC_PORT",
		"MYC_VERBOSITY",
		"MYC_PRINT_TREE",
		"MYC_XFORM_TIMEOUT",
	}
	for _, k := range unset {
		_ = os.Unsetenv(k)
	}
}
