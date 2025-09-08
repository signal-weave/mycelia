// * Error handling utilities

package errgo

import (
	"mycelia/globals"
	"mycelia/logging"
)

type MyceliaError interface {
	Error() string
	Verbosity() int
}

type myError struct {
	msg       string
	verbosity int
}

func (me myError) Verbosity() int {
	return me.verbosity
}

func (e myError) Error() string {
	return e.msg
}

func NewError(msg string, verbosity int) error {
	e := myError{
		verbosity: verbosity,
		msg:       msg,
	}
	LogError(e)

	return e
}

func LogError(e MyceliaError) {
	switch e.Verbosity() {

	case globals.VERB_ERR:
		logging.LogSystemError(e.Error())

	case globals.VERB_WRN:
		logging.LogSystemWarning(e.Error())

	case globals.VERB_ACT:
		logging.LogSystemAction(e.Error())

	}
}
