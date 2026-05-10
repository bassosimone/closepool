//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/rbmk-project/rbmk/blob/v0.18.0/pkg/common/closepool/closepool.go
//

// Package closepool allows pooling [io.Closer] instances
// and closing them in a single operation.
package closepool

import (
	"errors"
	"io"
	"slices"
	"sync"
)

// CloserFunc allows to turn a suitable function into an [io.Closer].
type CloserFunc func() error

// Ensure that [CloserFunc] implements [io.Closer].
var _ io.Closer = CloserFunc(nil)

// Close implements io.Closer.
func (fx CloserFunc) Close() error {
	return fx()
}

// Pool allows pooling a set of [io.Closer].
//
// The zero value is ready to use.
type Pool struct {
	// handles contains the [io.Closer] to close.
	handles []io.Closer

	// mu provides mutual exclusion.
	mu sync.Mutex
}

// Add adds a given [io.Closer] to the pool.
//
// This method is concurrency safe.
func (p *Pool) Add(closer io.Closer) {
	p.mu.Lock()
	p.handles = append(p.handles, closer)
	p.mu.Unlock()
}

// Move transfers ownership of the pooled [io.Closer] instances
// to a new [Pool] and returns it. The source pool is reset to
// empty, so a deferred [Pool.Close] on it becomes a no-op. The
// destination pool inherits the LIFO close order.
//
// This enables the "armed defer + commit on success" pattern: a
// caller defers [Pool.Close] for cleanup on the error paths, then
// calls Move on the success path to hand ownership to the result.
//
// This method is concurrency safe.
func (p *Pool) Move() *Pool {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := &Pool{handles: p.handles}
	p.handles = nil
	return out
}

// Close closes all the [io.Closer] inside the pool iterating
// in backward order. Therefore, if one registers a TCP connection
// and then the corresponding TLS connection, the TLS connection
// is closed first. The returned error is the join of all the
// errors that occurred when closing connections.
//
// This method is concurrency safe.
func (p *Pool) Close() error {
	// Lock and copy the [io.Closer] to close.
	p.mu.Lock()
	handles := p.handles
	p.handles = nil
	p.mu.Unlock()

	// Close all the [io.Closer].
	var errv []error
	for _, handle := range slices.Backward(handles) {
		if err := handle.Close(); err != nil {
			errv = append(errv, err)
		}
	}
	return errors.Join(errv...)
}
