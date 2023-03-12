package benchmarks

import (
	"context"
	"sync"
	"testing"

	"github.com/antonmashko/taskq"
)

func TaskqTestF(b *testing.B, f func(context.Context)) {
	tq := taskq.New(0)
	var wg sync.WaitGroup
	_ = tq.Start()

	b.ResetTimer()
	wg.Add(b.N)

	for i := 0; i < b.N; i++ {
		tq.Enqueue(context.Background(),
			taskq.TaskFunc(func(ctx context.Context) error {
				f(ctx)
				wg.Done()
				return nil
			}))
	}

	wg.Wait()
}

func SpawningGoroutinesTestF(b *testing.B, f func(context.Context)) {
	wg := sync.WaitGroup{}
	b.ResetTimer()

	wg.Add(b.N)

	for i := 0; i < b.N; i++ {
		go func(ctx context.Context) error {
			f(ctx)
			wg.Done()
			return nil
		}(context.Background())
	}
	wg.Wait()
}
