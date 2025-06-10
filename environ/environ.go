package environ

import (
	"os"
)

const VERBOSITY_ENV string = "VERBOSITY"

// A map of verbosity level to its corresponding string status.
var VerbosityStatusMap = map[int]string{
	0: "NONE",
	1: "ACTION",
	2: "WARNING",
	3: "ERROR",
}

// A map of verbosity status to its corresponding integer level.
var VerbosityLevelMap = map[string]int{
	"NONE":    0,
	"ACTION":  1,
	"WARNING": 2,
	"ERROR":   3,
}

// Gets the integer verbosity level from the status in the environ var.
func GetVerbosityLevel() int {
	verbosity := os.Getenv("VERBOSITY")
	return VerbosityLevelMap[verbosity]
}
