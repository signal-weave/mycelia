package main

import (
	"fmt"
	"os"
	"strings"

	"mycelia/boot"
	"mycelia/server"
	"mycelia/str"
)

var startupBanner string = `
            ███╗   ███╗██╗   ██╗ ██████╗███████╗██╗     ██╗ █████╗
            ████╗ ████║╚██╗ ██╔╝██╔════╝██╔════╝██║     ██║██╔══██╗
            ██╔████╔██║ ╚████╔╝ ██║     █████╗  ██║     ██║███████║
            ██║╚██╔╝██║  ╚██╔╝  ██║     ██╔══╝  ██║     ██║██╔══██║
            ██║ ╚═╝ ██║   ██║   ╚██████╗███████╗███████╗██║██║  ██║
            ╚═╝     ╚═╝   ╚═╝    ╚═════╝╚══════╝╚══════╝╚═╝╚═╝  ╚═╝`

var disclaimer string = "Mycelia is a work-in-progress concurrent message broker."

func main() {
	displayStartupText()
	readArgs()
	startServer()
}

// Prints ascii banner, disclaimer text, and any misc info.
func displayStartupText() {
	line := strings.Repeat("-", 80)

	fmt.Println(line)
	fmt.Println(startupBanner)
	fmt.Println(line)
	fmt.Println(disclaimer)
	fmt.Println(line)
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
	server := server.NewServer(boot.RuntimeCfg.Address, boot.RuntimeCfg.Port)
	if len(boot.CommandList) > 0 {
		for _, cmd := range boot.CommandList {
			err := server.Broker.HandleCommand(cmd)
			if err != nil {
				str.WarningPrint(err.Error())
			}
		}
	}

	server.Run()
}
