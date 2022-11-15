package taskq

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type workersSpawner struct {
	workerCounter  uint64
	update         chan struct{}
	queue          Queue
	limitCh        chan struct{}
	onDequeueError func()

	queueChecked int32
	closed       int32
	masters      []*worker

	lock    sync.Mutex
	workers map[uint64]*worker
}

func newWorkersSpawner(limit int, queue Queue) *workersSpawner {
	var limitCh chan struct{}
	if limit > 0 {
		limitCh = make(chan struct{}, limit)
	}
	return &workersSpawner{
		workerCounter: 0,
		update:        make(chan struct{}, runtime.NumCPU()),
		workers:       make(map[uint64]*worker),
		limitCh:       limitCh,
		queue:         queue,
	}
}

func (ws *workersSpawner) Enqueue(_ Task) bool {
	// notify worker about pending task
	select {
	case ws.update <- struct{}{}:
		// we have free worker
		return true
	default:
		// all workers are busy
		return false
	}
}

func (ws *workersSpawner) Start(ctx context.Context) {
	const count = 2

	ws.masters = make([]*worker, count)
	for i := 0; i < count; i++ {
		innerCtx, cancel := context.WithCancel(ctx)
		ws.masters[i] = &worker{
			id:     uint64(i),
			stop:   cancel,
			status: registered,
		}

		go func(ctx context.Context, w *worker) {
			for {
				w.setStatus(idle)
				select {
				case <-ctx.Done():
					w.setStatus(stopped)
					return
				case <-ws.update:
					w.setStatus(live)
					for atomic.LoadInt32(&ws.closed) != 1 {
						task, err := ws.queue.Dequeue(ctx)
						if err != nil {
							if err == EmptyQueue {
								atomic.StoreInt32(&ws.queueChecked, 1)
								break
							}
							if ws.onDequeueError == nil {
								panic(err)
							}
							break
						}
						ws.run(ctx, task)
					}
				}
			}
		}(innerCtx, ws.masters[i])
	}

	var ready bool
	for !ready {
		ready = true
		// waiting until all masters in idle state
		for _, w := range ws.masters {
			if w.isStatus(registered) {
				ready = false
				break
			}
		}
	}
}

func (ws *workersSpawner) run(ctx context.Context, t Task) {
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
		select {
		case <-ctx.Done():
			w.setStatus(stopped)
		default:
			processTask(ctx, task)
			ws.lock.Lock()
			w.stop()
			w.setStatus(stopped)
			delete(ws.workers, w.id)
			ws.lock.Unlock()
			if ws.limitCh != nil {
				<-ws.limitCh
			}
		}
	}(innerCtx, w, t)
}

func (ws *workersSpawner) hasActiveWorkers() bool {
	ws.lock.Lock()
	for _, w := range ws.workers {
		w.stop()
		if !w.isStatus(stopped) {
			ws.lock.Unlock()
			return true
		}
	}
	for _, w := range ws.masters {
		//if w.isStatus(idle) {
		w.stop()
		//}
		if !w.isStatus(stopped) {
			ws.lock.Unlock()
			return true
		}
	}
	ws.lock.Unlock()
	return false
}

func (ws *workersSpawner) Shutdown(ctx context.Context) error {
	defer close(ws.update)

	if wait, ok := ctx.Value(ctxWaitKey{}).(bool); !ok || !wait {
		atomic.StoreInt32(&ws.closed, 1)
	}

	var pollDuration = time.Millisecond * 500
	timer := time.NewTimer(pollDuration)
	for !ws.hasActiveWorkers() {
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
