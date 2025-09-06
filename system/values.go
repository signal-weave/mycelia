package system

import (
	"fmt"
	"os"
	"path/filepath"

	"mycelia/errgo"
	"mycelia/globals"
	"mycelia/protocol"
)

// -----------------------------------------------------------------------------
// Herein are the shared values that system sub-packages can reference.
// -----------------------------------------------------------------------------

// -------Directory + Files-----------------------------------------------------

// The directory the program is running from.
func GetExecDirectory() string {
	exePath := errgo.ValueOrPanic(os.Executable())
	exeDir := filepath.Dir(exePath)
	return exeDir
}

var exeDir = GetExecDirectory()

var ConfigFile = fmt.Sprintf("%s/Mycelia_Config.json", exeDir)

// This is a list of commands used for booting up and pre-configuring the broker
// based on the Mycelia_Config.json file.
var ObjectList = []*protocol.Object{}

// -------System Runtime Data Structures----------------------------------------

// Proxy struct for unmarshalling the Mycelia_Config.json runtime data into
// cleanly.
// This handles type conversion - Go marshals json integers to float64 by
// default for whatever fucking reason.
type ParamData struct {
	Address          *string   `json:"address"`
	Port             *int      `json:"port"`
	Verbosity        *int      `json:"verbosity"`
	PrintTree        *bool     `json:"print-tree"`
	TransformTimeout *string   `json:"xform-timeout"`
	AutoConsolidate  *bool     `json:"consolidate"`
	SecurityToken    *[]string `json:"security-tokens"`
}

func NewParamData() *ParamData {
	timeoutStr := globals.TransformTimeout.String()

	return &ParamData{
		Address:          &globals.Address,
		Port:             &globals.Port,
		Verbosity:        &globals.Verbosity,
		PrintTree:        &globals.PrintTree,
		TransformTimeout: &timeoutStr,
		AutoConsolidate:  &globals.AutoConsolidate,
		SecurityToken:    &globals.SecurityTokens,
	}
}

// Values detailing how the broker shutdown last.
type ShutdownReport struct {
	GracefulShutdown *bool `json:"graceful-shutdown"`
}

// Any global dynamic values, shutdown deatils, or pre-defined routes.
type SystemData struct {
	ShutdownReport *ShutdownReport   `json:"shutdown-report"`
	Parameters     *ParamData        `json:"parameters"`
	Routes         *[]map[string]any `json:"routes"`
}

func NewSystemData() *SystemData {
	shutdownStatus := false
	report := &ShutdownReport{GracefulShutdown: &shutdownStatus}

	routes := []map[string]any{}

	return &SystemData{
		ShutdownReport: report,
		Parameters:     NewParamData(),
		Routes:         &routes,
	}
}
