package taskq

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type workerSpawner struct {
	lock          sync.Mutex
	workerCounter uint64
	limitCh       chan struct{}
	workers       map[uint64]*worker
}

func newWorkerSpawner(limit int) *workerSpawner {
	var limitCh chan struct{}
	if limit > 0 {
		limitCh = make(chan struct{}, limit)
	}
	return &workerSpawner{
		workers: make(map[uint64]*worker),
		limitCh: limitCh,
	}
}

func (ws *workerSpawner) Len() int {
	return 2
}

func (ws *workerSpawner) Run(ctx context.Context, t Task) {
	innerCtx, cancel := context.WithCancel(ctx)
	w := &worker{
		status: registered,
		stop:   cancel,
	}

	ws.lock.Lock()
	for {
		id := atomic.AddUint64(&ws.workerCounter, 1)
		if _, ok := ws.workers[id]; !ok {
			w.id = id
			ws.workers[id] = w
			break
		}

	}
	ws.lock.Unlock()

	if ws.limitCh != nil {
		ws.limitCh <- struct{}{}
	}

	go func(ctx context.Context, w *worker, task Task) {
		w.setStatus(live)
		processTask(ctx, task)
		ws.lock.Lock()
		w.stop()
		w.setStatus(stopped)
		delete(ws.workers, w.id)
		ws.lock.Unlock()
		if ws.limitCh != nil {
			<-ws.limitCh
		}
	}(innerCtx, w, t)
}

func (ws *workerSpawner) Shutdown(ctx context.Context) error {
	var pollDuration = time.Millisecond * 500
	timer := time.NewTimer(pollDuration)
	var exit bool
	for {
		ws.lock.Lock()
		exit = len(ws.workers) <= 0
		ws.lock.Unlock()
		if exit {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			const delta = 1.1
			timer.Reset(time.Duration(float64(pollDuration) * delta))
		}
	}
}
