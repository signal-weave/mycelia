package boot

import "os"

// -----------------------------------------------------------------------------
// Herein is the starup process related functions, all neatly placed in one
// file.
// This is the top of the cli + pre-init stack.
// -----------------------------------------------------------------------------

// Parses and stores the runtime flags in public var.
func ParseRuntimeArgs(argv []string) error {
	// -------CLI values--------------------------------------------------------
	err := parseRuntimeArgs(argv)
	if err != nil {
		// We do not make a Mycelia Error here because main hands this in stdout
		return err
	}

	// -------PreInit.json values-----------------------------------------------
	_, err = os.Stat(preInitFile)
	if err == nil {
		getPreInitData()
	}

	return nil
}
