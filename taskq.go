package taskq

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

type taskqType byte

const (
	pool taskqType = iota + 1
	spawn
)

var (
	ErrClosed  = errors.New("taskq closed")
	ErrNilTask = errors.New("nil task")
)

type taskManager interface {
	// Len return number of requires workers
	Len() int
	// Run task from queue
	Run(context.Context, Task)
	// Shutdown finish active tasks and dispose workers
	Shutdown(context.Context) error
}

type TaskQ struct {
	queue Queue

	isRunning int32
	isClosing int32
	isStopped int32

	OnDequeueError func(ctx context.Context, workerID uint64, err error)

	taskManager taskManager
	// workers
	workers []*worker
	update  chan struct{}
}

func New(limit int) *TaskQ {
	return NewWithQueue(limit, NewConcurrentQueue())
}

func NewWithQueue(limit int, q Queue) *TaskQ {
	return newWithTypeAndQueue(limit, spawn, q)
}

func Pool(size int) *TaskQ {
	return PoolWithQueue(size, NewConcurrentQueue())
}

func PoolWithQueue(size int, q Queue) *TaskQ {
	return newWithTypeAndQueue(size, pool, q)
}

func newWithTypeAndQueue(size int, tp taskqType, q Queue) *TaskQ {
	var tm taskManager
	switch tp {
	case pool:
		tm = newWorkerPool(size)
	default:
		tm = newWorkerSpawner(size)
	}
	return &TaskQ{
		taskManager: tm,
		queue:       q,
		isRunning:   0,
		isClosing:   0,
		isStopped:   0,
		workers:     make([]*worker, tm.Len()),
		// update channel should always have length more than the workers count for avoiding deadlock
		update:         make(chan struct{}, tm.Len()+1),
		OnDequeueError: nil,
	}
}

func (t *TaskQ) triggerDequeue() bool {
	// notify worker about pending task (without blocking)
	select {
	case t.update <- struct{}{}:
		return true
	default:
		return false
	}
}

func (t *TaskQ) Enqueue(ctx context.Context, task Task) (int64, error) {
	if task == nil {
		return -1, ErrNilTask
	}

	if atomic.LoadInt32(&t.isClosing) != 0 {
		return -1, ErrClosed
	}

	id, err := t.queue.Enqueue(ctx, task)
	if err != nil {
		return -1, err
	}

	t.triggerDequeue()
	return id, nil
}

func (t *TaskQ) Start() error {
	if !atomic.CompareAndSwapInt32(&t.isRunning, 0, 1) {
		return errors.New("taskq is running")
	}

	ctx := context.Background()
	for i := range t.workers {
		innerCtx, cancel := context.WithCancel(ctx)
		t.workers[i] = &worker{
			id:     uint64(i),
			status: registered,
			stop:   cancel,
		}

		go func(ctx context.Context, w *worker) {
			// Triggering dequeue in case if Taskq was created with not empty queue
			t.triggerDequeue()
			for {
				w.setStatus(idle)
				select {
				case <-ctx.Done():
					w.setStatus(stopped)
					return
				case _, ok := <-t.update:
					if !ok {
						w.setStatus(stopped)
						return
					}
					w.setStatus(live)
					for atomic.LoadInt32(&t.isStopped) != 1 {
						task, err := t.queue.Dequeue(ctx)
						if err != nil {
							if err == EmptyQueue {
								break
							}
							if t.OnDequeueError == nil {
								panic(err)
							}

							t.OnDequeueError(ctx, w.id, err)
							break
						}
						t.taskManager.Run(ctx, task)
					}
				}
			}
		}(innerCtx, t.workers[i])
	}

	var ready bool
	for !ready {
		ready = true
		for _, w := range t.workers {
			if w.isStatus(registered) {
				ready = false
				break
			}
		}
	}

	return nil
}

func (t *TaskQ) hasActive() bool {
	if len(t.update) != 0 {
		return true
	}

	// checking if we still have active workers
	for _, w := range t.workers {
		if w.isStatus(live) {
			return true
		}
		// worker is ready to be stopped
		// nothing in update channel and no active tasks inside
		w.stop()
	}

	for _, w := range t.workers {
		if !w.isStatus(stopped) {
			return true
		}
	}

	return false
}

func (t *TaskQ) Shutdown(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&t.isClosing, 0, 1) {
		return errors.New("taskq is closing")
	}

	defer close(t.update)
	if wait, _ := ctx.Value(ctxWaitKey{}).(bool); !wait {
		atomic.StoreInt32(&t.isStopped, 1)
	} else {
		t.triggerDequeue()
	}

	err := t.taskManager.Shutdown(ctx)
	if err != nil {
		return err
	}

	var pollDuration = time.Millisecond * 500
	timer := time.NewTimer(pollDuration)
	for t.hasActive() {
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

func (t *TaskQ) Close() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return t.Shutdown(ctx)
}
