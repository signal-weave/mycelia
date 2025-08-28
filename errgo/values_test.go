package errgo

import (
	"errors"
	"testing"
)

func TestValueOrPanic_ReturnsValue(t *testing.T) {
	val := ValueOrPanic("ok", nil)
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

	_ = ValueOrPanic("fail", errors.New("some error"))
}

func TestPanicIfError_NoPanicOnNil(t *testing.T) {
	// should not panic
	PanicIfError(nil)
}

func TestPanicIfError_PanicsOnError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but function did not panic")
		}
	}()

	PanicIfError(errors.New("fatal"))
}
