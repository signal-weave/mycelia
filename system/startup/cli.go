package startup

import (
	"errors"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"regexp"
	"strings"

	"mycelia/globals"

	"github.com/signal-weave/siglog"
)

// ParseRuntimeArgs parses only runtime flags validates, and returns error.
//
// Duration examples: 500ms, 3s, 2m, 1h.
func parseRuntimeArgs(argv []string) error {
	// ! REMEMBER TO UPDATE fs.Usage() STRING WHEN ADDING / REMOVING CLI VARS !

	fs := flag.NewFlagSet("runtime", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	addrHelp := "Bind address (IP or hostname)"
	fs.StringVar(&globals.Address, "address", globals.Address, addrHelp)

	fs.IntVar(&globals.Port, "port", globals.Port, "Bind port (1-65535)")

	printHelp := "Print router tree at startup"
	fs.BoolVar(&globals.PrintTree, "print-tree", globals.PrintTree, printHelp)

	xformTimeoutHelp := "Transformer timeout (e.g. 30s, 2m)"
	fs.DurationVar(
		&globals.TransformTimeout, "xform-timeout", globals.TransformTimeout,
		xformTimeoutHelp,
	)

	workersHelp := "The number of server workers to allocate to the listener"
	fs.IntVar(&globals.WorkerCount, "workers", globals.WorkerCount, workersHelp)

	cleanHelp := "Whether to auto-consolidate router shape on component removal"
	fs.BoolVar(
		&globals.AutoConsolidate, "consolidate", globals.AutoConsolidate,
		cleanHelp,
	)

	verbosityHelp := `0 - None
    1 - Errors
    2 - Warnings + Errors
    3 - Errors + Warnings + Actions`
	var temp int
	fs.IntVar(&temp, "verbosity", int(globals.Verbosity), verbosityHelp)
	globals.Verbosity = siglog.LogLevel(temp)
	globals.UpdateVerbosityEnvironVar()

	logOutputHelp := `0 - .log file
	1 - console
	2 - both`
	fs.IntVar(
		&globals.LogOutput, "log-output", globals.LogOutput, logOutputHelp,
	)

	const usageString = `Mycelia runtime options:

  -address string      Bind address (IP or hostname)
  -port int            Bind port (1-65535)
  -workers int		   The server listener worker count (1-1024)
  -verbosity int       0, 1, 2, or 3
  -log-output int	   0, 1, or 2
  -print-tree          Print router tree at startup
  -xform-timeout dur   Transformer timeout

Examples:
  mycelia -addr 0.0.0.0 -port 8080 -verbosity 2 -print-tree -xform-timeout 45s
`
	fs.Usage = func() { fmt.Print(fs.Output(), usageString) }

	if err := fs.Parse(argv); err != nil {
		return err
	}

	if err := validateRuntimeConfig(); err != nil {
		return err
	}

	return nil
}

func validateRuntimeConfig() error {
	if globals.Port < 1 || globals.Port > 65535 {
		return fmt.Errorf("invalid port %d (expected 1-65535)", globals.Port)
	}
	// Allow hostnames; validate if it looks like an IP.
	if !isIPLiteral(globals.Address) && !isValidHostname(globals.Address) {
		return fmt.Errorf("invalid IP address %q", globals.Address)
	}
	if globals.TransformTimeout <= 0 {
		return errors.New("xform-timeout must be > 0")
	}
	if globals.WorkerCount <= 0 || globals.WorkerCount > 1024 {
		return fmt.Errorf("invalid worker count %d", globals.WorkerCount)
	}
	if globals.LogOutput < 0 || globals.LogOutput > 2 {
		return fmt.Errorf("invalid log output value %d", globals.LogOutput)
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
var hostnameLabelRE = regexp.MustCompile(
	`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`,
)

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
