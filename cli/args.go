package cli

import (
	"flag"
	"fmt"
	"os"

	"mycelia/environ"
)

var (
	Address = "127.0.0.1"
	Port    = 5000
)

func ParseCLIArgs() {
	flag.StringVar(&Address, "address", Address,
		fmt.Sprintf("The TCP address, without port. Defaults to %s.", Address))

	flag.IntVar(&Port, "port", Port,
		fmt.Sprintf("The port to listen to. Defaults to %d.", Port))

	var verbosityLevel int
	verbosityHelp := `The verbosity level for console output:
    0 - None
    1 - Actions
    2 - Warnings + Actions
    3 - Errors + Warnings + Actions`
	flag.IntVar(&verbosityLevel, "verbosity", 0, verbosityHelp)

	flag.Parse()

	if verbosity, ok := environ.VerbosityStatusMap[verbosityLevel]; ok {
		os.Setenv(environ.VERBOSITY_ENV, verbosity)
	} else {
		fmt.Println("Invalid verbosity specified, defaulting to NONE!")
		os.Setenv(environ.VERBOSITY_ENV, environ.VerbosityStatusMap[0])
	}
}
