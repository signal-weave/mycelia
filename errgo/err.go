package errgo

import (
	"mycelia/logging"

	"github.com/signal-weave/siglog"
)

type MyceliaError interface {
	Error() string
	Verbosity() siglog.LogLevel
}

type myError struct {
	msg       string
	verbosity siglog.LogLevel
}

func (me myError) Verbosity() siglog.LogLevel {
	return me.verbosity
}

func (e myError) Error() string {
	return e.msg
}

func NewError(msg string, verbosity siglog.LogLevel) error {
	e := myError{
		verbosity: verbosity,
		msg:       msg,
	}
	LogError(e)

	return e
}

func LogError(e MyceliaError) {
	switch e.Verbosity() {

	case siglog.LL_ERROR:
		logging.LogSystemError(e.Error())

	case siglog.LL_WARN:
		logging.LogSystemWarning(e.Error())

	case siglog.LL_INFO:
		logging.LogSystemAction(e.Error())
	}
}
