package main

import (
	"fmt"
	"os"

	"mycelia/globals"
	"mycelia/server"
	"mycelia/system"
	"mycelia/system/shutdown"
	"mycelia/system/startup"
)

func main() {
	updateVersion() // Be sure to update!

	startup.Startup(os.Args[1:])

	globals.PrintDynamicValues()

	startServer() // Performs loop until globals.PerformShutdown is true.

	shutdown.Shutdown()
}

// Sets the build metadata fields denoting what kind of build that will be
// generated.
func updateVersion() {
	system.BuildMetadata.MajorVersion = 0  // Proud version
	system.BuildMetadata.MinorVersion = 17 // Real  version
	system.BuildMetadata.PatchVersion = 1  // Sucky version

	system.BuildMetadata.ReleaseType = system.ReleaseDev
	system.BuildMetadata.DevVersion = 1
}

// Starts the server - checks for preloaded commands from the PreInit.json file
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
