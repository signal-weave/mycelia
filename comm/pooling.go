package comm

import (
	"sync"

	"mycelia/globals"
)

// -----------------------------------------------------------------------------
// Herein are pooling helpers for connections and message buffers.
// -----------------------------------------------------------------------------

// BufPool is a reusable sync.Pool to reduce the number of byte buffer creations.
// Will reduce GC calls and big allocations.
//
// - Use the pool inside hot-path I/O functions.
//
// - Borrow buffer → use it → copy out → return buffer.
//
// - Don’t store or hold pooled buffers in broker; return them right after the
// read/write.
var BufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 256*globals.BytesInKilobyte) // 256 KB scratch buffer
		return &b
	},
}
