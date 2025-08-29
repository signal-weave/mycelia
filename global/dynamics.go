package global

import (
	"os"
	"strconv"
	"time"
)

// -----------------------------------------------------------------------------
// Shared, or "global", dynamic values that are referenced between packages.
// This is not meant to contain constant values.
// -----------------------------------------------------------------------------

var Address string = "127.0.0.1"
var Port int = 5000
var Verbosity int = 0 // 0=quiet, 1=info, 2=debug, 3=trace...
var PrintTree bool = false
var TransformTimeout time.Duration = time.Duration(5) * time.Second

func UpdateVerbosityEnvironVar() {
	os.Setenv("VERBOSITY", strconv.Itoa(Verbosity))
}
