package taskq

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestWaitGroupEnqueue(t *testing.T) {
	wg := NewWaitGroup(2)
	var result int64
	const expected = 10
	for i := 0; i < expected; i++ {
		wg.Enqueue(context.Background(), TaskFunc(func(ctx context.Context) error {
			atomic.AddInt64(&result, 1)
			return nil
		}))
	}
	wg.Wait()

	if res := atomic.LoadInt64(&result); res != expected {
		t.Fatalf("invalid result number. expected=%d got=%d", expected, res)
	}
}
