package comm

import (
	"net"
	"sync"
)

// Connection responder that manages the net.Conn created by the server.
// To be used throughout message brokerage so no routing components need to own
// the conn object.
type ConnResponder struct {
	C  net.Conn
	mu sync.Mutex
}

func NewConnResponder(conn net.Conn) *ConnResponder {
	return &ConnResponder{
		C: conn,
	}
}

// Send the given payload back to the connection's return address.
func (r *ConnResponder) Write(b []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, err := r.C.Write(b)
	return err
}
