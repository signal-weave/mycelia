package boot

import (
	"errors"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"regexp"
	"strings"

	"mycelia/global"
)

// ParseRuntimeArgs parses only runtime flags validates, and returns
// (config, error).
//
// Duration examples: 500ms, 3s, 2m, 1h.
func parseRuntimeArgs(argv []string) error {
	fs := flag.NewFlagSet("runtime", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	fs.StringVar(&global.Address, "address", global.Address, "Bind address (IP or hostname)")
	fs.IntVar(&global.Port, "port", global.Port, "Bind port (1-65535)")
	fs.BoolVar(&global.PrintTree, "print-tree", global.PrintTree, "Print router tree at startup")
	fs.DurationVar(&global.TransformTimeout, "xform-timeout", global.TransformTimeout, "Transformer timeout (e.g. 30s, 2m)")
	fs.IntVar(&global.Verbosity, "verbosity", global.Verbosity,
		`0 - None
    1 - Errors
    2 - Warnings + Errors
    3 - Errors + Warnings + Actions`)
	global.UpdateVerbosityEnvironVar()

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), `Mycelia runtime options:

  -address string      Bind address (IP or hostname)
  -port int            Bind port (1-65535)
  -verbosity int       0, 1, 2, or 3
  -print-tree          Print router tree at startup
  -xform-timeout dur   Transformer timeout

Examples:
  mycelia -addr 0.0.0.0 -port 8080 -verbosity 2 -print-tree -xform-timeout 45s
`)
	}

	if err := fs.Parse(argv); err != nil {
		return err
	}

	if err := validateRuntimeConfig(); err != nil {
		return err
	}

	return nil
}

func validateRuntimeConfig() error {
	if global.Port < 1 || global.Port > 65535 {
		return fmt.Errorf("invalid port %d (expected 1-65535)", global.Port)
	}
	// Allow hostnames; validate if it looks like an IP.
	if !isIPLiteral(global.Address) && !isValidHostname(global.Address) {
		return fmt.Errorf("invalid IP address %q", global.Address)
	}
	if global.TransformTimeout <= 0 {
		return errors.New("xform-timeout must be > 0")
	}
	return nil
}

func isIPLiteral(s string) bool {
	_, err := netip.ParseAddr(s)
	return err == nil
}

// isValidHostname does a syntax-only RFC-1123 style check (no DNS lookups).
// - total length <= 253
// - labels are 1..63 chars, [A-Za-z0-9-], no leading/trailing '-'
var hostnameLabelRE = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)

func isValidHostname(s string) bool {
	if len(s) == 0 || len(s) > 253 {
		return false
	}
	// Special-case: common localhost
	if s == "localhost" {
		return true
	}
	labels := strings.Split(s, ".")
	for _, lbl := range labels {
		if !hostnameLabelRE.MatchString(lbl) {
			return false
		}
	}
	return true
}
