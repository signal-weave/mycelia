package cli

import (
	"flag"
)

var Address string
var Port int

func ParseCLIArgs() {
	addressHelp := "The TCP address, without port, defaults to 127.0.0.1"
	addressArg := flag.String("address", "127.0.0.1", addressHelp)

	portHelp := "The port to listen to, defaults to 5000."
	portArg := flag.Int("port", 5000, portHelp)

	flag.Parse()
	Address = *addressArg
	Port = *portArg
}
