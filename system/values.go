package system

import (
	"path/filepath"

	"mycelia/globals"
	"mycelia/protocol"
)

// -----------------------------------------------------------------------------
// Herein are the shared values that system sub-packages can reference.
// -----------------------------------------------------------------------------

var ConfigFile = filepath.Join(globals.ExeDir, "Mycelia_Config.json")

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
	LogOutput        *int      `json:"log-output"`
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
