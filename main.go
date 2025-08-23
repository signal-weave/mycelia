package main

import (
	"fmt"
	"os"

	"mycelia/boot"
	"mycelia/server"
	"mycelia/str"
)

func main() {
	err := boot.ParseRuntimeArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	server := server.NewServer(boot.RuntimeCfg.Address, boot.RuntimeCfg.Port)
	fmt.Println("command list len:", len(boot.CommandList))
	if len(boot.CommandList) > 0 {
		for _, cmd := range boot.CommandList {
			err = server.Broker.HandleCommand(cmd)
			if err != nil {
				str.WarningPrint("Unknown command type")
			}
		}
	}

	server.Run()
}
