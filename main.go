package main

import (
	"fmt"
	"mycelia/cli"
	"mycelia/server"
	"os"
)

func main() {
	err := cli.ParseRuntimeArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	server := server.NewServer(cli.RuntimeCfg.Address, cli.RuntimeCfg.Port)
	server.Run()
}
