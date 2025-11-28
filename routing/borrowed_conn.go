package routing

import (
	"context"
	"errors"
	"net"
	"time"
)

// The Borrowed connection from the pool.
type Borrowed struct {
	conn   net.Conn
	owner  *addrPool
	broken bool
}

type borrower interface {
	Conn() net.Conn
	MarkBroken()
	Put()
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

func (b *Borrowed) Conn() net.Conn { return b.conn }
func (b *Borrowed) MarkBroken()    { b.broken = true }

// Put releases the borrowed connection back into the pool (unless broken).
func (b *Borrowed) Put() {
	if b.conn == nil || b.owner == nil {
		return
	}
	if b.broken {
		_ = b.conn.Close()
		return
	}

	select {
	case b.owner.ch <- b.conn:
		b.owner.mu.Lock()
		b.owner.lastUsed[b.conn] = time.Now()
		b.owner.mu.Unlock()
	default:
		// Close if full
		_ = b.conn.Close()
	}
}

func (b *Borrowed) SetReadDeadline(t time.Time) error {
	if b.conn == nil {
		return errors.New("nil conn")
	}
	return b.conn.SetReadDeadline(t)
}

func (b *Borrowed) SetWriteDeadline(t time.Time) error {
	if b.conn == nil {
		return errors.New("nil conn")
	}
	return b.conn.SetWriteDeadline(t)
}

// Function to dial the borrowed connection from the pool.
// Override for testing.
// The default is globalConnPool.Get
var dialBorrowed = func(ctx context.Context, addr string) (borrower, error) {
	return globalConnPool.Get(ctx, addr)
}
