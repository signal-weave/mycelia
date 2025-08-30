package main

import (
	"fmt"
	"os"

	"mycelia/boot"
	"mycelia/globals"
	"mycelia/server"
	"mycelia/str"
)

var disclaimer string = "Mycelia is a work-in-progress concurrent message broker."

var majorVersion int = 0 // Proud version
var minorVersion int = 8 // Real version
var patchVersion int = 0 // Sucky verison

func main() {
	str.PrintStartupText(majorVersion, minorVersion, patchVersion)
	readArgs()
	globals.PrintDynamicValues()
	startServer() // Must be last, contains infinite for loop.
}

// Parses CLI args and PreInit.json file.
func readArgs() {
	err := boot.ParseRuntimeArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

// Starts the server - checks for pre-loaded commands from the PreInit.json file
// and loads them into the server's broker, then runs the server.
func startServer() {
	server := server.NewServer(globals.Address, globals.Port)
	if len(boot.CommandList) > 0 {
		for _, cmd := range boot.CommandList {
			server.Broker.HandleCommand(cmd)
		}
	}

	server.Run()
}
