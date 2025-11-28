package routing

import (
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/signal-weave/rhizome"
)

// -------Connections-----------------------------------------------------------

// a Conn whose Write always fails (to test MarkBroken path)
type errWriteConn struct{ net.Conn }

func (e errWriteConn) Write([]byte) (int, error) {
	return 0, errors.New("boom")
}

// -------Borrowers-------------------------------------------------------------

type fakeBorrowed struct {
	conn             net.Conn
	putCalled        bool
	markBrokenCalled bool
}

func (f *fakeBorrowed) Put()                               { f.putCalled = true }
func (f *fakeBorrowed) MarkBroken()                        { f.markBrokenCalled = true }
func (f *fakeBorrowed) Conn() net.Conn                     { return f.conn }
func (f *fakeBorrowed) SetReadDeadline(t time.Time) error  { return nil }
func (f *fakeBorrowed) SetWriteDeadline(t time.Time) error { return nil }

// -------Rhizome Objects-------------------------------------------------------

// makeFakeObj creates a fake object with a random UUID.
func makeFakeObj(payload string) *rhizome.Object {
	return rhizome.NewObject(
		0, 0, 0,
		uuid.NewString(),
		"", "", "", "",
		0, []byte(payload),
	)
}

// -------Subscribers-----------------------------------------------------------

// mkSubs returns a slice of subscribers with addresses A, B, C, ...
func mkSubs(n int) []subscriber {
	out := make([]subscriber, n)
	for i := 0; i < n; i++ {
		out[i] = subscriber{Address: string(rune('A' + i))}
	}
	return out
}

// -------Partition-------------------------------------------------------------

// Start a partition with a tiny buffered input and stop cleanly
func newTestPartition() *partition {
	p := &partition{
		route:   &route{},   // not used directly thanks to seams
		channel: &channel{}, // not used directly thanks to seams
	}
	p.in = make(chan *rhizome.Object, 1)
	p.start()
	return p
}
