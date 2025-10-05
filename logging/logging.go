package logging

import (
	"github.com/signal-weave/siglog"
)

// ----------Object Logging----------

func LogObjectAction(msg, uid string) {
	siglog.LogEntry(msg, uid, siglog.LL_INFO)
}

func LogObjectWarning(msg, uid string) {
	siglog.LogEntry(msg, uid, siglog.LL_WARN)
}

func LogObjectError(msg, uid string) {
	siglog.LogEntry(msg, uid, siglog.LL_ERROR)
}

// ----------System Logging----------

func LogSystemAction(msg string) {
	siglog.LogEntry(msg, "SYSTEM", siglog.LL_INFO)
}

func LogSystemWarning(msg string) {
	siglog.LogEntry(msg, "SYSTEM", siglog.LL_WARN)
}

func LogSystemError(msg string) {
	siglog.LogEntry(msg, "SYSTEM", siglog.LL_ERROR)
}
