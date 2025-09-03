package comm

import (
	"net"
	"sync"
)

type Responder interface {
	Write(data []byte) error
	Close() error
}

// Connection responder that manages the net.Conn created by the server.
// To be used throughout message brokerage so no routing components need to own
// the conn object.
type ConnResponder struct {
	c  net.Conn
	mu sync.Mutex
}

func NewConnResponder(conn net.Conn) *ConnResponder {
	return &ConnResponder{
		c: conn,
	}
}

// Send the given payload back to the connection's return address.
func (r *ConnResponder) Write(b []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return WriteFrameU32(r.c, b)
}

// Closes connection...
func (r *ConnResponder) Close() error {
	return r.c.Close()
}
