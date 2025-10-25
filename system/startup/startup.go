package startup

import (
	"fmt"
	"mycelia/str"
	"os"

	"mycelia/globals"
	"mycelia/logging"
	"mycelia/system"

	"github.com/signal-weave/siglog"
)

// -----------------------------------------------------------------------------
// Herein is the startup process related functions, all neatly placed in one
// file.
// This is the top of the cli + pre-init/recovery stack.
// -----------------------------------------------------------------------------

// Startup reads cli / config values and initializes systems like logging, etc.
func Startup(argv []string) {
	initializeLogger()
	logging.LogSystemAction("Starting startup Process!")

	str.PrintStartupText(system.BuildMetadata.String())
	parseCli(argv)
	parseConfigFile()

	logging.LogSystemAction("Ending startup Process!")
}

// initializeLogger sets all the logging values including log level, output
// directory, and batch mode.
func initializeLogger() {
	err := siglog.SetLogDirectory(globals.LogDirectory)
	if err != nil {
		fmt.Println(fmt.Sprintf("Could not create log dir: %s", err))
		return
	}
	err = siglog.SetOutput(siglog.OUT_FILE)
	if err != nil {
		fmt.Println(fmt.Sprintf("Could not set log out file: %s", err))
		return
	}
	err = siglog.SetLogLevel(siglog.LL_DEBUG)
	if err != nil {
		fmt.Println(fmt.Sprintf("Unable to set log level environ var: %s", err))
		return
	}

	if err := siglog.SetBatchMode(siglog.BATCH_NONE); err != nil {
		fmt.Println(fmt.Sprintf("Could not set log batch mode: %s", err))
	}
}

// Parses and stores the runtime flags in public vars.
func parseCli(argv []string) {
	err := parseRuntimeArgs(argv)
	if err != nil {
		fmt.Println(fmt.Sprintf("Could not parse runtime args: %s", err))
		// We do not make a Mycelia Error here because main hands this to stdout
		_, err = fmt.Fprintln(os.Stderr, err)
		fmt.Println(fmt.Sprintf("Could not write to stdout: %s", err))
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
