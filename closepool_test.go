//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/rbmk-project/rbmk/blob/v0.18.0/pkg/common/closepool/closepool_test.go
//

package closepool_test

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bassosimone/closepool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCloser implements io.Closer for testing.
type mockCloser struct {
	closed atomic.Int64
	err    error
}

// t0 is the time when we started running.
var t0 = time.Now()

func (m *mockCloser) Close() error {
	m.closed.Add(int64(time.Since(t0)))
	return m.err
}

func TestCloserFunc(t *testing.T) {
	var closed bool
	pool := &closepool.Pool{}
	pool.Add(closepool.CloserFunc(func() error {
		closed = true
		return nil
	}))
	pool.Close()
	require.True(t, closed)
}

func TestPool(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		pool := closepool.Pool{}
		m1 := &mockCloser{}
		m2 := &mockCloser{}

		pool.Add(m1)
		pool.Add(m2)

		err := pool.Close()
		require.NoError(t, err)
		assert.Greater(t, m1.closed.Load(), int64(0))
		assert.Greater(t, m2.closed.Load(), int64(0))
	})

	t.Run("close order", func(t *testing.T) {
		pool := closepool.Pool{}
		m1 := &mockCloser{}
		m2 := &mockCloser{}

		pool.Add(m1) // Added first
		pool.Add(m2) // Added second

		// Should close in reverse order
		err := pool.Close()
		require.NoError(t, err)
		assert.Greater(t, m1.closed.Load(), m2.closed.Load())
	})

	t.Run("error handling", func(t *testing.T) {
		pool := closepool.Pool{}
		expectedErr1 := errors.New("close error #1")
		expectedErr2 := errors.New("close error #2")

		m1 := &mockCloser{err: expectedErr1}
		m2 := &mockCloser{err: expectedErr2}

		pool.Add(m1)
		pool.Add(m2)

		err := pool.Close()
		require.Error(t, err)
		require.Equal(t, errors.Join(expectedErr2, expectedErr1).Error(), err.Error())
	})

	t.Run("concurrent usage", func(t *testing.T) {
		pool := closepool.Pool{}
		done := make(chan struct{})

		// Concurrently add closers
		go func() {
			for i := 0; i < 100; i++ {
				pool.Add(&mockCloser{})
			}
			close(done)
		}()

		// Add more closers from main goroutine
		for i := 0; i < 100; i++ {
			pool.Add(&mockCloser{})
		}

		<-done // Wait for goroutine to finish

		err := pool.Close()
		require.NoError(t, err)
	})
}
