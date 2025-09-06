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

// Read cli / config values...
func Boot(argv []string) {
	parseCli(argv)
	parseConfigFile()
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
func parseConfigFile() {
	_, err := os.Stat(system.ConfigFile)
	if err == nil {
		getConfigData()
	}
}
