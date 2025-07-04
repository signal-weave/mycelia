package error_test

import (
	"errors"
	"testing"

	myerror "mycelia/error" // alias to avoid conflict with built-in "error"
)

func TestValueOrPanic_ReturnsValue(t *testing.T) {
	val := myerror.ValueOrPanic("ok", nil)
	if val != "ok" {
		t.Errorf("Expected 'ok', got: %v", val)
	}
}

func TestValueOrPanic_PanicsOnError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but function did not panic")
		}
	}()

	_ = myerror.ValueOrPanic("fail", errors.New("some error"))
}

func TestPanicIfError_NoPanicOnNil(t *testing.T) {
	// should not panic
	myerror.PanicIfError(nil)
}

func TestPanicIfError_PanicsOnError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but function did not panic")
		}
	}()

	myerror.PanicIfError(errors.New("fatal"))
}
