package global

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"mycelia/commands"
)

// -----------------------------------------------------------------------------
// Shared, or "global", dynamic values that are referenced between packages.
// This is not meant to contain constant values.
// -----------------------------------------------------------------------------

var Address string = "127.0.0.1"
var Port int = 5000
var Verbosity int = 0 // 0=quiet, 1=info, 2=debug, 3=trace...
var PrintTree bool = false
var TransformTimeout time.Duration = time.Duration(5) * time.Second

func UpdateVerbosityEnvironVar() {
	os.Setenv("VERBOSITY", strconv.Itoa(Verbosity))
}

func UpdateGlobalsByMessage(m *commands.Globals) {
	Address = m.Address
	Port = m.Port
	Verbosity = m.Verbosity
	UpdateVerbosityEnvironVar()
	PrintTree = m.PrintTree

	oldTimeout := TransformTimeout
	newTimeout, err := time.ParseDuration(m.TransformTimeout)
	if err != nil {
		TransformTimeout = oldTimeout
		return
	}
	TransformTimeout = newTimeout

	PrintDynamicValues()
}

func PrintDynamicValues() {
	fmt.Println("----------Current Dynamic Global Values----------")
	fmt.Printf("Address: %s\n", Address)
	fmt.Printf("Port: %v\n", Port)
	fmt.Printf("Verbosity: %v\n", Verbosity)
	fmt.Printf("PrintTree: %v\n", PrintTree)
	fmt.Printf("TransformTimeout: %s\n", TransformTimeout.String())
	fmt.Println("-------------------------------------------------")
}
