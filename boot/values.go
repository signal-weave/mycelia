package boot

import (
	"fmt"
	"os"
	"path/filepath"

	"mycelia/protocol"
	"mycelia/errgo"
)

// -----------------------------------------------------------------------------
// Herein are the shared values that other packages can reference or each file
// in the boot package can reference and some helper funcs to populate them.
// -----------------------------------------------------------------------------

// ------Pre-defined Structure--------------------------------------------------

func getExecDirectory() string {
	exePath := errgo.ValueOrPanic(os.Executable())
	exeDir := filepath.Dir(exePath)
	return exeDir
}

var exeDir = getExecDirectory()
var preInitFile = fmt.Sprintf("%s/PreInit.json", exeDir)

// Parse command type funcs append their command to this list.
var CommandList = []*protocol.Command{}
