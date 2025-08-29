package main

import (
	"fmt"
	"os"

	"mycelia/boot"
	"mycelia/global"
	"mycelia/server"
	"mycelia/str"
)

var disclaimer string = "Mycelia is a work-in-progress concurrent message broker."

var majorVersion int = 0 // Proud version
var minorVersion int = 7 // Real version
var patchVersion int = 1 // Sucky verison

func main() {
	str.PrintStartupText(majorVersion, minorVersion, patchVersion)
	readArgs()
	startServer() // Must be last, contains infinite for loop.
}

func readArgs() {
	err := boot.ParseRuntimeArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

// Starts the server - checks for pre-loaded commands in the PreInit.json file
// and loads them into the server's broker, then runs the server.
func startServer() {
	server := server.NewServer(global.Address, global.Port)
	if len(boot.CommandList) > 0 {
		for _, cmd := range boot.CommandList {
			server.Broker.HandleCommand(cmd)
		}
	}

	server.Run()
}
