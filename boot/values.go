package boot

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mycelia/commands"
	"mycelia/errgo"
)

// -----------------------------------------------------------------------------
// Herein are the shared values that other packages can reference or each file
// in the boot package can reference and some helper funcs to populate them.
// -----------------------------------------------------------------------------

// ------Pre-defined Structure--------------------------------------------------

func getExecDirectory() string {
	exePath := errgo.ValueOrPanic(os.Executable())
	exeDir := filepath.Dir(exePath)
	return exeDir
}

var exeDir = getExecDirectory()
var preInitFile = fmt.Sprintf("%s/PreInit.json", exeDir)

// Parse command type funcs append their command to this list.
var CommandList = []commands.Command{}

// ------Configurable Globals---------------------------------------------------

type runtimeConfig struct {
	Address          string
	Port             int
	Verbosity        int // 0=quiet, 1=info, 2=debug, 3=trace...
	PrintTree        bool
	TransformTimeout time.Duration
}

func defaultRuntimeConfig() runtimeConfig {
	return runtimeConfig{
		Address:          "127.0.0.1",
		Port:             5000,
		Verbosity:        0,
		PrintTree:        false,
		TransformTimeout: 5 * time.Second,
	}
}

// -----------------------------------------------------------------------------
// The primary struct for getting cli values. They are stored here rather than
// environment vars so that they can be stored as non-string types like nubmers
// or time durations.
// -----------------------------------------------------------------------------
var RuntimeCfg = defaultRuntimeConfig()
