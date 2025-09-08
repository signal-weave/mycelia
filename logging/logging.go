package logging

import (
	"mycelia/globals"
)

func logMessage(msg, uid string, verbosity int) {
	if globals.Verbosity == globals.VERB_NIL {
		return
	}

	go func() {
		ml := &messageLog{
			uid:       uid,
			msg:       msg,
			verbosity: verbosity,
		}
		GlobalLogger.in <- ml
	}()
}

// ----------Object Logging----------

func LogObjectAction(msg, uid string) {
	logMessage(msg, uid, globals.VERB_ACT)
}

func LogObjectWarning(msg, uid string) {
	logMessage(msg, uid, globals.VERB_WRN)
}

func LogObjectError(msg, uid string) {
	logMessage(msg, uid, globals.VERB_ERR)
}

// ----------System Logging----------

func LogSystemAction(msg string) {
	logMessage(msg, "SYSTEM", globals.VERB_ACT)
}

func LogSystemWarning(msg string) {
	logMessage(msg, "SYSTEM", globals.VERB_WRN)
}

func LogSystemError(msg string) {
	logMessage(msg, "SYSTEM", globals.VERB_ERR)
}
