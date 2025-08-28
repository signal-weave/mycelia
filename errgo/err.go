// * Error handling utilities

package errgo

import (
	"mycelia/global"
	"mycelia/str"
)

type MyceliaError interface {
	Msg() string
	Verbosity() int
}

type myError struct {
	msg       string
	verbosity int
}

func (me myError) Msg() string {
	return me.msg
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

	case global.VERB_ERR:
		str.ErrorPrint(e.Msg())

	case global.VERB_WRN:
		str.WarningPrint(e.Msg())

	case global.VERB_ACT:
		str.ActionPrint(e.Msg())

	}
}
