// * Error handling utilities

package errgo

import (
	"mycelia/globals"
	"mycelia/str"
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
	AnnounceError(e)

	return e
}

func AnnounceError(e MyceliaError) {
	switch e.Verbosity() {

	case globals.VERB_ERR:
		str.ErrorPrint(e.Error())

	case globals.VERB_WRN:
		str.WarningPrint(e.Error())

	case globals.VERB_ACT:
		str.ActionPrint(e.Error())

	}
}
