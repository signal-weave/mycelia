package globals

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"
)

// -----------------------------------------------------------------------------
// Shared, or "global", dynamic values that are referenced between packages.
// This is not meant to contain constant values.
// -----------------------------------------------------------------------------

// -------Pseudo-Constants------------------------------------------------------

// If the server should begin the shutdown process.
var PerformShutdown atomic.Bool

// The address the server will use to listen for message.
var Address string = "127.0.0.1"

// The port the server will listen on.
var Port int = 5000

// How much output to send to the console.
// By default it is 3 until cli or config file adjust - This way we get the most
// visibility until user narrows view.
var Verbosity int = VERB_ACT // 0=quiet, 1=info, 2=debug, 3=trace...

// Where log messages should go to.
// Defaults to .log file.
var LogOutput int = LOG_TO_FILE // 0=.log file, 1=console.

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

// How many days a log file should be kept around.
// If set to 0, log file cleanup is ignored.
var MaxLogAge int = 14

func UpdateVerbosityEnvironVar() {
	os.Setenv("VERBOSITY", strconv.Itoa(Verbosity))
}

func PrintDynamicValues() {
	fmt.Println("----------Current Dynamic Global Values----------")
	fmt.Printf("Address: %s\n", Address)
	fmt.Printf("Port: %v\n", Port)
	fmt.Printf("Verbosity: %v\n", Verbosity)
	fmt.Printf("LogOutput: %v\n", LogOutput)
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

// -------Directories and Files-------------------------------------------------

// The directory the program is running from.
func GetExecDirectory() string {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	exeDir := filepath.Dir(exePath)
	return exeDir
}

// The directory the .exe file is ran rome.
var ExeDir = GetExecDirectory()

// The directory for log and chunk files.
var LogDirectory = filepath.Join(ExeDir, "logs")
