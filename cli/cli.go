// mycelia/cli/args.go
package cli

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type runtimeConfig struct {
	Address          string
	Port             int
	Verbosity        int // 0=quiet, 1=info, 2=debug, 3=trace...
	PrintTree        bool
	TransformTimeout time.Duration
}

func defaultRuntimeConfig() runtimeConfig {
	return runtimeConfig{
		Address:          "127.0.0.1",
		Port:             5000,
		Verbosity:        0,
		PrintTree:        false,
		TransformTimeout: 5 * time.Second,
	}
}

var RuntimeCfg = defaultRuntimeConfig()

// Parses and stores the runtime flags in public var.
func ParseRuntimeArgs(argv []string) error {
	cfg, err := parseRuntimeArgs(argv)
	if err != nil {
		return err
	}
	RuntimeCfg = cfg
	return nil
}

// ParseRuntimeArgs parses only runtime flags validates, and returns
// (config, error).
//
// Duration examples: 500ms, 3s, 2m, 1h.
func parseRuntimeArgs(argv []string) (runtimeConfig, error) {
	cfg := defaultRuntimeConfig()

	fs := flag.NewFlagSet("runtime", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	fs.StringVar(&cfg.Address, "addr", cfg.Address, "Bind address (IP or hostname)")
	fs.IntVar(&cfg.Port, "port", cfg.Port, "Bind port (1-65535)")
	fs.BoolVar(&cfg.PrintTree, "print-tree", cfg.PrintTree, "Print router tree at startup")
	fs.DurationVar(&cfg.TransformTimeout, "xform-timeout", cfg.TransformTimeout, "Transformer timeout (e.g. 30s, 2m)")
	fs.IntVar(&cfg.Verbosity, "verbosity", cfg.Verbosity,
		`0 - None
    1 - Errors
    2 - Warnings + Errors
    3 - Errors + Warnings + Actions`)

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), `Mycelia runtime options:

  -addr string         Bind address (IP or hostname)
  -port int            Bind port (1-65535)
  -v int               0, 1, 2, or 3
  -print-tree          Print router tree at startup
  -xform-timeout dur   Transformer timeout

Examples:
  mycelia -addr 0.0.0.0 -port 8080 -verbosity 2 -print-tree -xform-timeout 45s
  MYC_ADDR=0.0.0.0 MYC_PORT=8080 mycelia -v
`)
	}

	if err := fs.Parse(argv); err != nil {
		return cfg, err
	}

	if err := validateRuntimeConfig(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func validateRuntimeConfig(c runtimeConfig) error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port %d (expected 1-65535)", c.Port)
	}
	// Allow hostnames; validate if it looks like an IP.
	if ip := net.ParseIP(c.Address); ip == nil && looksLikeIP(c.Address) {
		return fmt.Errorf("invalid IP address %q", c.Address)
	}
	if c.TransformTimeout <= 0 {
		return errors.New("xform-timeout must be > 0")
	}
	return nil
}

func looksLikeIP(s string) bool {
	// crude: "n.n.n.n" suggests intended IP; otherwise treat as hostname
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return true
		}
		if _, err := strconv.Atoi(p); err != nil {
			return true
		}
	}
	return true
}
