package taskq

import (
	"context"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

type Status byte

const (
	None Status = iota
	Pending
	InProgress
	Done
	Failed
)

type itask struct {
	id     int64
	status Status
	task   Task
}

func (t *itask) Do(ctx context.Context) error {
	return t.task.Do(ctx)
}

func TaskStatus(task Task) Status {
	if it, ok := task.(*itask); ok && it != nil {
		return it.status
	}
	return None
}

type workerStatus int32

const (
	registered workerStatus = iota
	idle
	live
	stopped
)

type worker struct {
	id     int
	status workerStatus
	stop   func()
}

func (w *worker) isStatus(s workerStatus) bool {
	return workerStatus(atomic.LoadInt32((*int32)(&w.status))) == s
}

func (w *worker) setStatus(s workerStatus) {
	atomic.StoreInt32((*int32)(&w.status), int32(s))
}

type adaptedQueue struct {
	Queue
}

func (q *adaptedQueue) enqueue(it *itask) {
	q.Queue.Enqueue(it)
}

func (q *adaptedQueue) dequeue() *itask {
	t := q.Queue.Dequeue()
	if t == nil {
		return nil
	}
	return t.(*itask)
}

type TaskQ struct {
	lastInc int64

	closed     int32
	hasUpdates chan struct{}
	pending    *adaptedQueue

	workers []*worker

	TaskDone   func(int64, Task)
	TaskFailed func(int64, Task, error)
}

func New(size int) *TaskQ {
	return NewWithQueue(size, NewConcurrentQueue())
}

func NewWithQueue(size int, q Queue) *TaskQ {
	if size < 1 {
		size = runtime.NumCPU()
	}
	return &TaskQ{
		hasUpdates: make(chan struct{}, size),
		workers:    make([]*worker, size),
		pending:    &adaptedQueue{Queue: q},
	}
}

func (t *TaskQ) enqueue(task Task) *itask {
	if atomic.LoadInt32(&t.closed) != 0 || task == nil {
		return nil
	}
	it := &itask{
		id:     atomic.AddInt64(&t.lastInc, 1),
		status: Pending,
		task:   task,
	}
	t.pending.enqueue(it)
	t.triggerUpdateNotification()
	return it
}

func (t *TaskQ) Enqueue(task Task) int64 {
	it := t.enqueue(task)
	if it == nil {
		return -1
	}
	return it.id
}

func (t *TaskQ) triggerUpdateNotification() bool {
	// notify worker about pending task
	select {
	case t.hasUpdates <- struct{}{}:
		// we have free worker
		return true
	default:
		// all workers are busy
		return false
	}
}

func (t *TaskQ) Start() error {
	parentCtx := context.Background()
	started := make(chan struct{})
	defer close(started)

	// run process workers
	for i := 0; i < len(t.workers); i++ {
		ctx, cancel := context.WithCancel(parentCtx)
		t.workers[i] = &worker{id: i, stop: cancel}

		go func(ctx context.Context, w *worker) {
			started <- struct{}{}
			for {
				w.setStatus(idle)
				select {
				case <-ctx.Done():
					w.setStatus(stopped)
					return
				case _, ok := <-t.hasUpdates:
					if !ok {
						w.setStatus(stopped)
						return
					}
					w.setStatus(live)
					for {
						task := t.pending.dequeue()
						if task == nil {
							break
						}
						// if some of workers in idle state we will trigger it for processing tasks from queue
						t.triggerUpdateNotification()
						t.process(ctx, task)
					}
				}
			}
		}(ctx, t.workers[i])

		<-started
		t.triggerUpdateNotification()
	}

	return nil
}

func (t *TaskQ) process(ctx context.Context, it *itask) {
	it.status = InProgress
	err := it.task.Do(ctx)
	if err != nil {
		it.status = Failed
		if t.TaskFailed != nil {
			t.TaskFailed(it.id, it.task, err)
		}
	}
	it.status = Done
	if t.TaskDone != nil {
		t.TaskDone(it.id, it.task)
	}
}

func (t *TaskQ) Shutdown(ctx context.Context) error {
	const pollDuration = time.Millisecond * 500
	atomic.StoreInt32(&t.closed, 1) // no more accepting tasks
	defer close(t.hasUpdates)

	for _, worker := range t.workers {
		worker.stop()
	}
	srv := http.Server{}
	srv.Shutdown(ctx)

	timer := time.NewTimer(pollDuration)
	for {
		exit := true
		for _, w := range t.workers {
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
			// TODO: increase polling time
			// time.Reset(pollDuration * delta)
		}
	}
}

func (t *TaskQ) Close() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return t.Shutdown(ctx)
}
