package comm

import (
	"sync"
	"testing"

	"mycelia/globals"
)

func TestBufPool_Get_SizeAndType(t *testing.T) {
	v := BufPool.Get()
	defer BufPool.Put(v)

	p, ok := v.(*[]byte)
	if !ok {
		t.Fatalf("BufPool.Get returned %T, want *[]byte", v)
	}
	if p == nil || *p == nil {
		t.Fatalf("BufPool.Get returned nil pointer or nil slice")
	}

	wantBase := 64 * globals.BytesInKilobyte
	if len(*p) < wantBase {
		t.Fatalf("len(*p) = %d, want at least %d", len(*p), wantBase)
	}
	if cap(*p) < wantBase {
		t.Fatalf("cap(*p) = %d, want at least %d", cap(*p), wantBase)
	}
}

func TestBufPool_PutGet_TypeSafety(t *testing.T) {
	first := BufPool.Get()
	p1, ok := first.(*[]byte)
	if !ok || p1 == nil {
		t.Fatalf("first BufPool.Get type invalid: %T", first)
	}
	if len(*p1) > 0 {
		(*p1)[0] ^= 0xFF
	}
	BufPool.Put(first)

	second := BufPool.Get()
	defer BufPool.Put(second)
	if _, ok := second.(*[]byte); !ok {
		t.Fatalf("second BufPool.Get type invalid: %T", second)
	}
}

func TestBufPool_ConcurrentGetPut(t *testing.T) {
	const goroutines = 32
	const iters = 512

	var wg sync.WaitGroup
	wg.Add(goroutines)
	errCh := make(chan error, goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				v := BufPool.Get()
				p, ok := v.(*[]byte)
				if !ok || p == nil || *p == nil || len(*p) == 0 {
					errCh <- errTypeOrSize{}
					return
				}
				(*p)[0] ^= 0x01
				BufPool.Put(v)
			}
		}()
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("concurrent BufPool use failed: %v", err)
	}
}

type errTypeOrSize struct{}

func (e errTypeOrSize) Error() string { return "unexpected type or size from BufPool" }
