package taskq

import (
	"context"
	"runtime"
	"sync/atomic"
	"time"
)

type workersPool struct {
	workers        []*worker
	queue          Queue
	update         chan struct{}
	onDequeueError func(context.Context, int, error)
	closed         int32
}

func newWorkersPool(size int, q Queue) *workersPool {
	if size < 1 {
		size = runtime.NumCPU()
	}
	return &workersPool{
		workers: make([]*worker, size),
		update:  make(chan struct{}, size),
		queue:   q,
		closed:  0,
	}
}

func (wp *workersPool) Enqueue(t Task) bool {
	// notify worker about pending task
	select {
	case wp.update <- struct{}{}:
		// available worker
		return true
	default:
		// all workers are busy
		return false
	}
}

func (wp *workersPool) Start(ctx context.Context) {
	for i := range wp.workers {
		innerCtx, cancel := context.WithCancel(ctx)
		wp.workers[i] = &worker{
			id:     uint64(i),
			status: registered,
			stop:   cancel,
		}

		go func(ctx context.Context, w *worker) {
			wp.Enqueue(nil) // trigger processing elements from queue
			for {
				w.setStatus(idle)
				select {
				case <-ctx.Done():
					w.setStatus(stopped)
					return
				case _, ok := <-wp.update:
					if !ok {
						w.setStatus(stopped)
						return
					}
					w.setStatus(live)
					for atomic.LoadInt32(&wp.closed) != 1 {
						task, err := wp.queue.Dequeue(ctx)
						if err != nil {
							if err == EmptyQueue {
								break
							}
							if wp.onDequeueError == nil {
								panic(err)
							}
							break
						}
						processTask(ctx, task)
					}
				}
			}
		}(innerCtx, wp.workers[i])
	}

	var ready bool
	for !ready {
		ready = true
		for _, w := range wp.workers {
			if w.isStatus(registered) {
				ready = false
				break
			}
		}
	}
}

func (wp *workersPool) Shutdown(ctx context.Context) error {
	defer close(wp.update)

	if wait, ok := ctx.Value(ctxWaitKey{}).(bool); !ok || !wait {
		atomic.StoreInt32(&wp.closed, 1)
	}

	for _, worker := range wp.workers {
		worker.stop()
	}

	var pollDuration = time.Millisecond * 500
	timer := time.NewTimer(pollDuration)
	for {
		exit := true
		for _, w := range wp.workers {
			if !w.isStatus(stopped) {
				exit = false
				break
			}
		}

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
