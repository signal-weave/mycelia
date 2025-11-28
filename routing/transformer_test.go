package routing

import (
	"bytes"
	"context"
	"errors"
	"net"
	"testing"
)

func TestApply_HappyPath(t *testing.T) {
	// pipe simulates a real net.Conn pair
	server, client := net.Pipe()

	// stub dial to return our fake
	dialBorrowed = func(ctx context.Context, addr string) (borrower, error) {
		return &fakeBorrowed{conn: client}, nil
	}
	// restore after test
	defer func() {
		dialBorrowed = func(ctx context.Context, addr string) (borrower, error) {
			return globalConnPool.Get(ctx, addr)
		}
	}()

	// server goroutine: transform payload with a prefix
	go func() {
		defer server.Close()
		buf := make([]byte, 64)
		n, _ := server.Read(buf)
		_, _ = server.Write(append([]byte("X:"), buf[:n]...))
	}()

	tr := &transformer{Address: "ignored-in-test"}
	originalPayload := "hello"
	transformedPayload := "X:" + originalPayload

	obj := makeFakeObj(originalPayload)

	out, err := tr.apply(obj)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !bytes.Equal(out.Payload, []byte(transformedPayload)) {
		t.Fatalf("got %q, want %q", out.Payload, transformedPayload)
	}
}

func TestApply_DialError(t *testing.T) {
	dialBorrowed = func(ctx context.Context, addr string) (borrower, error) {
		return nil, errors.New("nope")
	}
	defer func() {
		dialBorrowed = func(ctx context.Context, addr string) (borrower, error) {
			return globalConnPool.Get(ctx, addr)
		}
	}()

	tr := &transformer{Address: "bad"}
	originalPayload := "hello"

	obj := makeFakeObj(originalPayload)
	out, err := tr.apply(obj)
	if err == nil {
		t.Fatal("expected error")
	}
	// contract: returns an original object on failure
	if string(out.Payload) != originalPayload {
		t.Fatalf("expected original payload, got %q", out.Payload)
	}
}
