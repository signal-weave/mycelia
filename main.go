package main

import (
	"os"

	"mycelia/globals"
	"mycelia/server"
	"mycelia/str"
	"mycelia/system"
	"mycelia/system/boot"
)

var disclaimer string = "Mycelia is a work-in-progress concurrent message broker."

var majorVersion int = 0  // Proud version
var minorVersion int = 11 // Real version
var patchVersion int = 0  // Sucky verison

func main() {
	str.PrintStartupText(majorVersion, minorVersion, patchVersion)

	boot.Boot(os.Args[1:])

	globals.PrintDynamicValues()

	startServer()
}

// Starts the server - checks for pre-loaded commands from the PreInit.json file
// and loads them into the server's broker, then runs the server.
func startServer() {
	server := server.NewServer(globals.Address, globals.Port)
	if len(system.CommandList) > 0 {
		for _, cmd := range system.CommandList {
			server.Broker.HandleCommand(cmd)
		}
	}

	server.Run()
}
