package globals

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// -----------------------------------------------------------------------------
// Shared, or "global", dynamic values that are referenced between packages.
// This is not meant to contain constant values.
// -----------------------------------------------------------------------------

// The address the server will use to listen for message.
var Address string = "127.0.0.1"

// The port the server will listen on.
var Port int = 5000

// How much output to send to the console.
var Verbosity int = 0 // 0=quiet, 1=info, 2=debug, 3=trace...

// Whether to print the broker shape on update.
var PrintTree bool = false

// Whether to remove empty routes + channels when items are removed.
var AutoConsolidate bool = true

// The deadline time when talking to a external machine.
var TransformTimeout time.Duration = time.Duration(5) * time.Second

func UpdateVerbosityEnvironVar() {
	os.Setenv("VERBOSITY", strconv.Itoa(Verbosity))
}

func PrintDynamicValues() {
	fmt.Println("----------Current Dynamic Global Values----------")
	fmt.Printf("Address: %s\n", Address)
	fmt.Printf("Port: %v\n", Port)
	fmt.Printf("Verbosity: %v\n", Verbosity)
	fmt.Printf("PrintTree: %v\n", PrintTree)
	fmt.Printf("TransformTimeout: %s\n", TransformTimeout.String())
	fmt.Printf("AutoConsolidate: %v\n", AutoConsolidate)
	fmt.Println("-------------------------------------------------")
}
