package taskq

import (
	"context"
	"sync/atomic"
	"time"
)

type workerSpawner struct {
	workerCounter int64
	limitCh       chan struct{}
}

func newWorkerSpawner(limit int) *workerSpawner {
	var limitCh chan struct{}
	if limit > 0 {
		limitCh = make(chan struct{}, limit)
	}
	return &workerSpawner{
		limitCh: limitCh,
	}
}

func (ws *workerSpawner) Len() int {
	return 2
}

func (ws *workerSpawner) Run(ctx context.Context, t Task) {
	atomic.AddInt64(&ws.workerCounter, 1)
	if ws.limitCh != nil {
		ws.limitCh <- struct{}{}
	}

	go func(ctx context.Context, task Task) {
		processTask(ctx, task)
		if ws.limitCh != nil {
			<-ws.limitCh
		}
		atomic.AddInt64(&ws.workerCounter, -1)
	}(context.Background(), t)
}

func (ws *workerSpawner) Shutdown(ctx context.Context) error {
	var pollDuration = time.Millisecond * 500
	timer := time.NewTimer(pollDuration)
	for atomic.LoadInt64(&ws.workerCounter) != 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			const delta = 1.1
			timer.Reset(time.Duration(float64(pollDuration) * delta))
		}
	}
	return nil
}
