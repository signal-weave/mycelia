package routing

import (
	"context"
	"net"
	"testing"
)

func TestSubscriberDeliver_HappyPath(t *testing.T) {
	server, client := net.Pipe()

	fb := &fakeBorrowed{conn: client}

	origDial := dialBorrowed
	dialBorrowed = func(ctx context.Context, addr string) (borrower, error) {
		return fb, nil
	}
	defer func() { dialBorrowed = origDial }()

	// Save the original subWrite function for restoration later.
	// Test will inject a fake that writes to fake borrowed Conn.
	origWrite := subWrite
	subWrite = func(b borrower, data []byte) error {
		return origWrite(b, data)
	}
	defer func() { subWrite = origWrite }()

	// server reads what deliver wrote
	readDone := make(chan []byte, 1)
	go func() {
		defer close(readDone)
		buf := make([]byte, 1<<16)
		n, _ := server.Read(buf)
		readDone <- append([]byte(nil), buf[:n]...)
		_ = server.Close()
	}()

	s := &subscriber{Address: "ignored"}
	payload := "hello-subscriber"
	obj := makeFakeObj(payload)

	s.deliver(obj)

	got := <-readDone
	if string(got) != payload {
		t.Fatalf("payload mismatch: got %s, want %s", string(got), payload)
	}
	if !fb.putCalled {
		t.Fatalf("expected Put() to be called")
	}
	if fb.markBrokenCalled {
		t.Fatalf("did not expect MarkBroken()")
	}
}

func TestSubscriberDeliver_WriteErrorMarksBroken(t *testing.T) {
	fb := &fakeBorrowed{conn: errWriteConn{}}

	origDial := dialBorrowed
	dialBorrowed = func(ctx context.Context, addr string) (borrower, error) { return fb, nil }
	defer func() { dialBorrowed = origDial }()

	s := &subscriber{Address: "ignored"}
	obj := makeFakeObj("hello")

	s.deliver(obj)

	if !fb.markBrokenCalled {
		t.Fatalf("expected MarkBroken() to be called on write error")
	}
	if !fb.putCalled {
		t.Fatalf("expected Put() to be called")
	}
}
