package globals

import (
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

// -----------------------------------------------------------------------------
// Shared, or "global", dynamic values that are referenced between packages.
// This is not meant to contain constant values.
// -----------------------------------------------------------------------------

// If the server should begin the shutdown process.
var PerformShutdown atomic.Bool

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

// The list of accepted security tokens to update the broker at runtime.
// Clients need any one of these to have permission.
var SecurityTokens []string = []string{}

// The default number of partitions a channel is created with.
var DefaultNumPartitions int = 4

// The number of workers to allocate to the server listener.
var WorkerCount int = 4

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

	fmt.Println("Accepted security tokens:")
	if len(SecurityTokens) > 0 {
		for _, v := range SecurityTokens {
			fmt.Printf("  %s\n", v)
		}
	} else {
		fmt.Println("  No registered security tokens.")
	}

	fmt.Println("-------------------------------------------------")
}
