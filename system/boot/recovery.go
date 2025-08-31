package boot

import (
	"encoding/json"
	"os"

	"mycelia/errgo"
	"mycelia/str"
	"mycelia/system"
)

// Check if a shutdown report exists, if it does then check if the broker
// gracefully shut down previously.
// If it did not, then perform a recovery.
func recover() {
	data, err := os.ReadFile(system.ShutdownReportFile)
	if err != nil {
		str.ErrorPrint("Could not import ShutdownReport JSON data")
		str.ErrorPrint("Skipping shutdown check.")
		return
	}

	var rd system.SystemData // Recovery Data
	err = json.Unmarshal(data, &rd)

	if rd.ShutdownReport == nil {
		str.ErrorPrint(
			"Shutdown report is missing shutdown status, skipping check",
		)
		system.DoRecovery = false
	}

	// Broker shut down in expected manner, do not recover.
	if *rd.ShutdownReport.GracefulShutdown {
		system.DoRecovery = false
		return
	}

	// Broker shut down unexpectedly, recover data.
	system.DoRecovery = true

	if rd.Parameters != nil {
		parseRuntimeConfigurable(*rd.Parameters)
	}
	if rd.Routes != nil {
		parseRouteCmds(*rd.Routes)
	}
}

func makeInitialShutdownFile() {
	data := map[string]any{
		"shutdown-report": map[string]bool{
			"graceful-shutdown": false,
		},
		"routes":     []map[string]any{},
		"parameters": *system.NewParamData(),
	}

	jsonData := errgo.ValueOrPanic(json.MarshalIndent(data, "", "    "))
	file := errgo.ValueOrPanic(os.Create(system.ShutdownReportFile))
	defer file.Close()

	_, err := file.WriteString(string(jsonData))
	errgo.PanicIfError(err)
	str.ActionPrint("ShutdownReport json created in exe directory.")
}
