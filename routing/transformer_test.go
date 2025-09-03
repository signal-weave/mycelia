package routing

import (
	"strings"
	"testing"
	"time"

	"mycelia/globals"
	"mycelia/protocol"
	"mycelia/test"
)

func newMsg(body string) *protocol.Object {
	return &protocol.Object{
		ObjType:      1,
		CmdType:      2,
		UID:          "uid-123",
		ReturnAdress: "127.0.0.1:9999",
		Arg1:         "routeA",
		Arg2:         "chanA",
		Arg3:         "",
		Arg4:         "",
		Payload:      []byte(body),
	}
}

func TestTransformer_transformDelivery_Success(t *testing.T) {
	// MockTwoWayServer echoes back "XFORM:" + body
	addr, stop := test.MockTwoWayServer(t, "XFORM:")
	defer stop()

	tr := newTransformer(addr)
	in := newMsg("hello")
	// Ensure timeout is comfortably long for CI
	globals.TransformTimeout = 2 * time.Second

	out, err := tr.apply(in)
	if err != nil {
		t.Fatalf("transformDelivery error: %v", err)
	}
	if out == nil {
		t.Fatalf("expected non-nil *object")
	}
	if string(out.Payload) != "XFORM:hello" {
		t.Fatalf("unexpected transformed payload: %q", string(out.Payload))
	}

	// Headers/fields should be preserved
	if out.ObjType != in.ObjType || out.CmdType != in.CmdType ||
		out.UID != in.UID || out.ReturnAdress != in.ReturnAdress ||
		out.Arg1 != in.Arg1 || out.Arg2 != in.Arg2 ||
		out.Arg3 != in.Arg3 || out.Arg4 != in.Arg4 {
		t.Fatalf("fields not preserved across transform")
	}
}

func TestTransformer_transformDelivery_DialFailure_ReturnsOriginal(t *testing.T) {
	tr := newTransformer("127.0.0.1:1") // very likely closed
	in := newMsg("ignored")

	out, err := tr.apply(in)
	if err == nil {
		t.Fatalf("expected dial error, got nil")
	}
	if out != in {
		t.Fatalf("expected original object to be returned on error")
	}
	if !strings.Contains(err.Error(), "Could not dial transformer") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransformer_transformDelivery_ReadError_ReturnsOriginal(t *testing.T) {
	// One-way server reads body then closes without responding → client's Read
	// fails.
	addr, gotBody, stop := test.MockOneWayServer(t)
	defer stop()

	tr := newTransformer(addr)
	in := newMsg("payload-abc")
	globals.TransformTimeout = 1 * time.Second

	out, err := tr.apply(in)
	if err == nil {
		t.Fatalf("expected read error, got nil")
	}
	if out != in {
		t.Fatalf("expected original object to be returned on read error")
	}
	if !strings.Contains(err.Error(), "Error reading from transformer") {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert server received what we sent
	select {
	case got := <-gotBody:
		if got != "payload-abc" {
			t.Fatalf("server saw wrong body: %q", got)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for mock server to receive payload")
	}
}
