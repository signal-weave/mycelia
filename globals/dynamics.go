package globals

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/signal-weave/siglog"
)

// -----------------------------------------------------------------------------
// Shared, or "global", dynamic values that are referenced between packages.
// This is not meant to contain constant values.
// -----------------------------------------------------------------------------

// -------Pseudo-Constants------------------------------------------------------

// PerformShutdown begins the server shutdown process when set to true.
var PerformShutdown atomic.Bool

// Address the server will use to listen for message.
var Address = "127.0.0.1"

// Port the server will listen on.
var Port = 5000

// Verbosity is how much output to send to the console.
// By default, it is 3 until cli or config file adjust - This way we get the most
// visibility until user narrows view.
// 0=quiet, 1=info, 2=debug, 3=trace...
var Verbosity = siglog.LL_INFO

// LogOutput where log messages should go to.
// Defaults to .log file.
var LogOutput = LogToFile // 0=.log file, 1=console, 2=both

// PrintTree enables debug tree printing.
var PrintTree = false

// AutoConsolidate removes empty routes + channels when items are removed.
var AutoConsolidate = true

// TransformTimeout is the deadline time when talking to a external machine.
var TransformTimeout = time.Duration(5) * time.Second

// SecurityTokens is the list of accepted security tokens to update the broker
// at runtime.
// Clients need any one of these to have permission.
var SecurityTokens []string

// DefaultNumPartitions is the number of partitions a channel is created with.
var DefaultNumPartitions = 4

// WorkerCount is the number of workers to allocate to the server listener.
var WorkerCount = 4

// MaxLogAge is how many days a log file should be kept around.
// If set to 0, log file cleanup is ignored.
var MaxLogAge = 14

func UpdateVerbosityEnvironVar() {
	_ = os.Setenv("VERBOSITY", strconv.Itoa(int(Verbosity)))
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

// getExecDirectory returns the directory the program is running from.
func getExecDirectory() string {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	exeDir := filepath.Dir(exePath)
	return exeDir
}

// ExeDir is the directory the .exe file is ran from.
var ExeDir = getExecDirectory()

// LogDirectory is the directory for log and chunk files.
var LogDirectory = filepath.Join(ExeDir, "logs")
