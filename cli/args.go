package cli

import (
	"flag"
	"fmt"
	"os"

	"mycelia/environ"
)

var Address string = "127.0.0.1"
var Port int = 5000

func ParseCLIArgs() {
	addressHelp := fmt.Sprintf("The TCP address, without port, defaults to %s",
		Address)
	addressArg := flag.String("address", Address, addressHelp)

	portHelp := fmt.Sprintf("The port to listen to, defaults to %d.", Port)
	portArg := flag.Int("port", Port, portHelp)

	verbosityHelp := "The verbosity level for various console print functions."
	verbosityHelp += "\nThe following will enable each plus the lower values."
	verbosityHelp += "\n0 - None, 1 - Actions, 2 - Warnings, 3 - Errors."
	verbArg := flag.Int("verbosity", 0, verbosityHelp)

	flag.Parse()

	Address = *addressArg
	Port = *portArg

	verbosity, ok := environ.VerbosityStatusMap[*verbArg]
	if !ok {
		fmt.Println("Invalid verbosity specified, defaulting to NONE!")
		os.Setenv(environ.VERBOSITY_ENV, environ.VerbosityStatusMap[0])
		return
	}
	os.Setenv(environ.VERBOSITY_ENV, verbosity)
}
