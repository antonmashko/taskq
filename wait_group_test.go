package taskq

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestWaitGroupEnqueue(t *testing.T) {
	wg := NewWaitGroup(0)
	var result int64
	for i := 0; i < 100; i++ {
		wg.Enqueue(context.Background(), TaskFunc(func(ctx context.Context) error {
			atomic.AddInt64(&result, 1)
			return nil
		}))
	}
	wg.Start()
	wg.Wait()

	if result != 100 {
		t.Fatal("invalid result number")
	}
}
