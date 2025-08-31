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

var PreInitFile = fmt.Sprintf("%s/PreInit.json", exeDir)
var ShutdownReportFile = fmt.Sprintf("%s/ShutdownReport.json", exeDir)

// Whether to read the shutdown report and perform a recovery if a crash status
// or unexpected shutdown was logged.
var DoRecovery = true

// Parse command type funcs append their command to this list.
var CommandList = []*protocol.Command{}

// -------System Runtime Data Structures----------------------------------------

// Proxy struct for unmarshalling the PreInit.json runtime data into cleanly.
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
	DoRecovery       *bool     `json:"do-recovery"`
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
		DoRecovery:       &DoRecovery,
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
