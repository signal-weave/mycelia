package test

import (
	"testing"
	"time"
)

func WithTimeout(t *testing.T, d time.Duration, fn func(t *testing.T)) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn(t)
	}()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("test timed out after %s", d)
	}
}
