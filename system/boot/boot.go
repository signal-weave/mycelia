package boot

import (
	"fmt"
	"os"

	"mycelia/system"
)

// -----------------------------------------------------------------------------
// Herein is the starup process related functions, all neatly placed in one
// file.
// This is the top of the cli + pre-init/recovery stack.
// -----------------------------------------------------------------------------

// Startup the program...
func Boot(argv []string) {
	parseCli(argv)
	fmt.Println("Recovery value:", system.DoRecovery)

	// If parseShutdownReport detected a crash or suspicious shutodwn, the we
	// want to use the recovery data instead of the PreInit data.
	if system.DoRecovery {
		parseShutdownReport()
	}
	// Second DoRecovery check because if the recovery process encounters and
	// error, stopping the recovery from happening, then it may switch off.
	if !system.DoRecovery {
		parsePreInitFile()
	}
}

// Parses and stores the runtime flags in public vars.
func parseCli(argv []string) {
	err := parseRuntimeArgs(argv)
	if err != nil {
		// We do not make a Mycelia Error here because main hands this in stdout
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

// Check for a Mycelia_Config.json file in the .exe directory.
// If found -> load values.
func parsePreInitFile() {
	_, err := os.Stat(system.PreInitFile)
	if err == nil {
		getPreInitData()
	}
}

// Check for a shutdown report.
// If found -> load values.
func parseShutdownReport() {
	_, err := os.Stat(system.ShutdownReportFile)
	if err != nil {
		makeInitialShutdownFile()
		return
	}
	recover()
}
