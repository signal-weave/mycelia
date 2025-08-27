package main

import (
	"fmt"
	"os"

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

var majorVersion int = 0
var minorVersion int = 6
var patchVersion int = 0
var brokerVersion string = fmt.Sprintf(
	"%d.%d.%d", majorVersion, minorVersion, patchVersion,
)
var verNotice string = fmt.Sprintf(
	"Running verison: %s", brokerVersion,
)

func main() {
	displayStartupText()
	readArgs()

	startServer() // Must be last, contains infinite for loop.
}

// Prints ascii banner, disclaimer text, and any misc info.
func displayStartupText() {
	str.PrintAsciiLine()
	fmt.Println(startupBanner)
	str.PrintAsciiLine()
	fmt.Println(verNotice)
	str.PrintAsciiLine()
	fmt.Println(disclaimer)
	str.PrintAsciiLine()
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
