package comm

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"mycelia/test"
)

func TestConnResponder_WriteAndReadFrame_RoundTrip(t *testing.T) {
	test.WithTimeout(t, 2*time.Second, func(t *testing.T) {
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()

		r := NewConnResponder(c1)
		want := []byte("hello responder")

		// Start reader FIRST so writer won't block on Pipe backpressure.
		gotCh := make(chan []byte, 1)
		errCh := make(chan error, 1)
		go func() {
			b, err := ReadFrameU32(c2)
			if err != nil {
				errCh <- err
				return
			}
			gotCh <- b
		}()

		if err := r.Write(want); err != nil {
			t.Fatalf("r.Write error: %v", err)
		}

		select {
		case err := <-errCh:
			t.Fatalf("ReadFrameU32 error: %v", err)
		case got := <-gotCh:
			if !bytes.Equal(got, want) {
				t.Fatalf("payload mismatch: got %q, want %q", got, want)
			}
		case <-time.After(1500 * time.Millisecond):
			t.Fatal("timeout waiting for reader")
		}
	})
}

func TestConnResponder_ConcurrentWrites_Serialized(t *testing.T) {
	test.WithTimeout(t, 3*time.Second, func(t *testing.T) {
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()

		r := NewConnResponder(c1)

		msgs := [][]byte{
			[]byte("a"),
			[]byte("bbb"),
			[]byte("cccc"),
			[]byte("dd"),
			[]byte("eeeee"),
		}

		// Start a reader that will consume exactly len(msgs) frames.
		type rec struct {
			b   []byte
			err error
		}
		out := make(chan rec, len(msgs))
		go func() {
			for i := 0; i < len(msgs); i++ {
				b, err := ReadFrameU32(c2)
				out <- rec{b: b, err: err}
				if err != nil {
					return
				}
			}
		}()

		// Now fire concurrent writes.
		var wg sync.WaitGroup
		wg.Add(len(msgs))
		for _, m := range msgs {
			m := m
			go func() {
				defer wg.Done()
				if err := r.Write(m); err != nil {
					t.Errorf("concurrent Write error: %v", err)
				}
			}()
		}
		wg.Wait()

		// Collect what the reader got.
		gotCounts := map[string]int{}
		for i := 0; i < len(msgs); i++ {
			select {
			case rec := <-out:
				if rec.err != nil {
					t.Fatalf("ReadFrameU32 error: %v", rec.err)
				}
				gotCounts[string(rec.b)]++
			case <-time.After(1500 * time.Millisecond):
				t.Fatalf("reader timed out after %d/%d frames", i, len(msgs))
			}
		}

		// Verify boundaries preserved (counts match).
		wantCounts := map[string]int{}
		for _, m := range msgs {
			wantCounts[string(m)]++
		}
		if len(gotCounts) != len(wantCounts) {
			t.Fatalf("unique payload count mismatch: got %d, want %d", len(gotCounts), len(wantCounts))
		}
		for k, want := range wantCounts {
			if got := gotCounts[k]; got != want {
				t.Fatalf("payload %q count = %d, want %d", k, got, want)
			}
		}
	})
}

func TestConnResponder_Close_ThenWriteErrors(t *testing.T) {
	test.WithTimeout(t, 2*time.Second, func(t *testing.T) {
		c1, c2 := net.Pipe()
		defer c2.Close()

		r := NewConnResponder(c1)

		if err := r.Close(); err != nil {
			t.Fatalf("Close error: %v", err)
		}
		err := r.Write([]byte("should fail"))
		if err == nil {
			t.Fatalf("expected error writing after Close, got nil")
		}
		// Accept any non-nil error; commonly io.ErrClosedPipe or net.ErrClosed.
		if !errors.Is(err, io.ErrClosedPipe) && !errors.Is(err, net.ErrClosed) {
			t.Logf("write after Close error: %v", err)
		}
	})
}
