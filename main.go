package main

import (
	"mycelia/server"
	"mycelia/cli"
)

func main() {
	cli.ParseCLIArgs()
	server := server.NewServer(cli.Address, cli.Port)
	server.Run()
}
