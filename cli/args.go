package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"mycelia/environ"
)

var (
	Address   = "127.0.0.1"
	Port      = 5000
	PrintTree = false

	verbosityLevel      = 0
	xformTimeoutSeconds = 5
)

func ParseCLIArgs() {
	setupCliArgs()
	flag.Parse()
	parseEnvironment()
}

// Register each CLI arg for parsing.
func setupCliArgs() {
	flag.StringVar(&Address, "address", Address,
		fmt.Sprintf("The TCP address, without port. Defaults to %s.", Address))

	flag.IntVar(&Port, "port", Port,
		fmt.Sprintf("The port to listen to. Defaults to %d.", Port))

	verbosityHelp := `The verbosity level for console output:
    0 - None
    1 - Errors
    2 - Warnings + Errors
    3 - Errors + Warnings + Actions`
	flag.IntVar(&verbosityLevel, "verbosity", 0, verbosityHelp)

	flag.BoolVar(&PrintTree, "printTree", PrintTree,
		"Print an ascii map after each registration.")

	flag.IntVar(&xformTimeoutSeconds, "xformTimeout", xformTimeoutSeconds,
		"Transformer response timeout in seconds.")
}

// Parse each variable that will get stored in the environment.
func parseEnvironment() {
	// ----------Verbosity----------
	if verbosity, ok := environ.VerbosityStatusMap[verbosityLevel]; ok {
		os.Setenv(environ.VERBOSITY_ENV, verbosity)
	} else {
		fmt.Println("Invalid verbosity specified, defaulting to NONE!")
		os.Setenv(environ.VERBOSITY_ENV, environ.VerbosityStatusMap[0])
	}

	// ----------XForm Timeout----------
	os.Setenv(environ.XFORM_TIMEOUT_ENV, strconv.Itoa(xformTimeoutSeconds))
}
