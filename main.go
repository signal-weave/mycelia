package main

import (
	"fmt"
	"os"

	"mycelia/globals"
	"mycelia/server"
	"mycelia/str"
	"mycelia/system"
	"mycelia/system/shutdown"
	"mycelia/system/startup"
)

var majorVersion int = 0  // Proud version
var minorVersion int = 15 // Real  version
var patchVersion int = 0  // Sucky version

func main() {
	str.PrintStartupText(majorVersion, minorVersion, patchVersion)

	startup.Startup(os.Args[1:])

	globals.PrintDynamicValues()

	startServer() // Performs loop until globals.PerformShutdown is true.

	shutdown.Shutdown()
}

// test

// Starts the server - checks for pre-loaded commands from the PreInit.json file
// and loads them into the server's broker, then runs the server.
func startServer() {
	s := server.NewServer(globals.Address, globals.Port)
	for _, cmd := range system.ObjectList {
		err := s.Broker.HandleObject(cmd)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	if err := s.Run(); err != nil {
		fmt.Println(err.Error())
	}
}
