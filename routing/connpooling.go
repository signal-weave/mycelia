package routing

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

// Pool is a per-address connection pool. It reuses TCP connections and
// redials only when the pool is empty or a connection is unhealthy/expired.
type Pool struct {
	pools sync.Map // map[string]*addrPool

	MaxPerAddr   int
	IdleTimeout  time.Duration
	DialTimeout  time.Duration
	KeepAlive    time.Duration
	ReadTimeout  time.Duration // optional default on borrow; set 0 to skip
	WriteTimeout time.Duration // optional default on borrow; set 0 to skip
}

type addrPool struct {
	addr   string
	ch     chan net.Conn // The pool
	dialer *net.Dialer
	idleTO time.Duration
	mu     sync.Mutex

	// track last-used to evict stale conns when returned
	lastUsed map[net.Conn]time.Time
}

type Borrowed struct {
	conn   net.Conn
	owner  *addrPool
	broken bool
}

// NewPool creates a pool with reasonable defaults.
func NewPool() *Pool {
	return &Pool{
		MaxPerAddr:   64,
		IdleTimeout:  60 * time.Second,
		DialTimeout:  5 * time.Second,
		KeepAlive:    30 * time.Second,
		ReadTimeout:  0,
		WriteTimeout: 0,
	}
}

// globalConnPool is the routing package singleton that manages all connections,
// tracking the per-address pools.
var globalConnPool = NewPool()

func (p *Pool) getAddrPool(addr string) *addrPool {
	v, ok := p.pools.Load(addr)
	if ok {
		return v.(*addrPool)
	}
	ap := &addrPool{
		addr:     addr,
		ch:       make(chan net.Conn, p.MaxPerAddr),
		dialer:   &net.Dialer{Timeout: p.DialTimeout, KeepAlive: p.KeepAlive},
		idleTO:   p.IdleTimeout,
		lastUsed: make(map[net.Conn]time.Time),
	}
	// publish-or-use the winner
	actual, _ := p.pools.LoadOrStore(addr, ap)
	return actual.(*addrPool)
}

// Get borrows a connection for addr (reused if available, else it dials).
func (p *Pool) Get(ctx context.Context, addr string) (*Borrowed, error) {
	ap := p.getAddrPool(addr)

	select {
	case c := <-ap.ch:
		ap.mu.Lock()
		lastUsed := ap.lastUsed[c]
		delete(ap.lastUsed, c)
		ap.mu.Unlock()
		if ap.idleTO > 0 && time.Since(lastUsed) > ap.idleTO {
			_ = c.Close()
			// fallthrough to dial...
		} else {
			return &Borrowed{conn: c, owner: ap}, nil
		}
	default:
		// no cached connection available, fall through to dial
	}

	type dialRes struct {
		c   net.Conn
		err error
	}

	done := make(chan dialRes, 1)
	go func() {
		c, e := ap.dialer.DialContext(ctx, "tcp", addr)
		done <- dialRes{c: c, err: e}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-done:
		if r.err != nil {
			return nil, r.err
		}
		return &Borrowed{conn: r.c, owner: ap}, nil
	}
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
