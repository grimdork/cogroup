// Package cogroup provides a small wrapper around sync.WaitGroup that
// exposes a readable counter. The implementation uses an atomic counter
// for the fast path and is the default exported type. WARNING: values of
// types in this package must not be copied after first use because they
// contain sync primitives.
package cogroup

import (
	"sync"
	"sync/atomic"
)

// noCopy may be embedded into structs which must not be copied after first use.
// The Lock method is a no-op but is recognised by `go vet` to find erroneous
// copies of values that embed noCopy.
// See: https://pkg.go.dev/go.dev/src/sync/atomic#example_NoCopy
type noCopy struct{}

func (*noCopy) Lock() {}

// CoGroup is an atomic-backed wait group with a readable counter. Do not copy
// values of this type after first use; it contains sync primitives.
type CoGroup struct {
	noCopy noCopy
	wg     sync.WaitGroup
	c      int64 // accessed via atomic operations
}

// Add increases the counter.
func (co *CoGroup) Add(n int) {
	atomic.AddInt64(&co.c, int64(n))
	co.wg.Add(n)
}

// Count returns the current count.
func (co *CoGroup) Count() int {
	return int(atomic.LoadInt64(&co.c))
}

// Done decrements the CoGroup counter by one.
func (co *CoGroup) Done() {
	atomic.AddInt64(&co.c, -1)
	co.wg.Done()
}

// Wait blocks until the counter reaches zero.
func (co *CoGroup) Wait() {
	co.wg.Wait()
}
