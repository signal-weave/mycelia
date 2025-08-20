package environ

import (
	"os"
	"strconv"
)

const (
	VERBOSITY_ENV     string = "VERBOSITY"
	XFORM_TIMEOUT_ENV string = "XFORM_TIMEOUT"
)

// A map of verbosity level to its corresponding string status.
var VerbosityStatusMap = map[int]string{
	0: "NONE",
	1: "ERROR",
	2: "WARNING",
	3: "ACTION",
}

// A map of verbosity status to its corresponding integer level.
var VerbosityLevelMap = map[string]int{
	"NONE":    0,
	"ERROR":   1,
	"WARNING": 2,
	"ACTION":  3,
}

// Gets the integer verbosity level from the status in the environ var.
func GetVerbosityLevel() int {
	verbosity := os.Getenv(VERBOSITY_ENV)
	return VerbosityLevelMap[verbosity]
}

// Gets the transformer timeout time in seconds.
func GetXformTimeout() int {
	string_val := os.Getenv(XFORM_TIMEOUT_ENV)
	int_val, err := strconv.Atoi(string_val)
	if err != nil {
		return 5
	}
	return int_val
}
