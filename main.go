package main

import (
	"os"

	"mycelia/globals"
	"mycelia/server"
	"mycelia/str"
	"mycelia/system"
	"mycelia/system/shutdown"
	"mycelia/system/startup"
)

var disclaimer string = "Mycelia is a work-in-progress concurrent message broker."

var majorVersion int = 0  // Proud version
var minorVersion int = 13 // Real version
var patchVersion int = 1  // Sucky verison

func main() {
	str.PrintStartupText(majorVersion, minorVersion, patchVersion)

	startup.Startup(os.Args[1:])

	globals.PrintDynamicValues()

	startServer() // Performs loop until globals.PerformShutdown is true.

	shutdown.Shutdown()
}

// Starts the server - checks for pre-loaded commands from the PreInit.json file
// and loads them into the server's broker, then runs the server.
func startServer() {
	server := server.NewServer(globals.Address, globals.Port)
	for _, cmd := range system.ObjectList {
		server.Broker.HandleObject(cmd)
	}

	server.Run()
}
