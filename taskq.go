package taskq

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
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

type TaskQ struct {
	lastInc int64
	isClosed chan struct{}
	queue   chan struct{}
	pending *blockingQueue
	//tasksMaxRetry int
	size int

	lock sync.Mutex

	TaskDone   func(int64, Task)
	TaskFailed func(int64, Task, error)
}

func New(size int) *TaskQ {
	if size < 1 {
		size = runtime.NumCPU()
	}
	return &TaskQ{
		size:  size,
		isClosed: make(chan struct{}),
		queue: make(chan struct{}, size),
		pending: &blockingQueue{
			queue: make([]*itask, 0, size),
		},
	}
}

func (t *TaskQ) enqueue(task Task) *itask {
	if task == nil {
		return nil
	}
	it := &itask{
		id:     atomic.AddInt64(&t.lastInc, 1),
		status: Pending,
		task:   task,
	}
	t.pending.enqueue(it)
	// notify worker about pending task
	select {
	case <- t.isClosed:
	default:
		select {
		case t.queue <- struct{}{}:
			// we have free worker
		default:
			// all worker's are busy
		}
	}
	return it
}

func (t *TaskQ) Enqueue(task Task) int64 {
	it := t.enqueue(task)
	if it == nil {
		return -1
	}
	return it.id
}

func (t *TaskQ) Start() error {
	// run process workers
	for i := 0; i < t.size; i++ {
		// each worker will make task.Do
		go func(workerID int) {
			for range t.queue {
				for {
					task := t.pending.dequeue()
					if task == nil {
						break
					}
					t.process(task)
				}
			}
		}(i)
	}
	return nil
}

func (t *TaskQ) process(it *itask) {
	it.status = InProgress
	err := it.task.Do(context.Background())
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

func (t *TaskQ) Close() error {
	select {
	case <-t.isClosed:
	default:
		close(t.isClosed)
		close(t.queue)
	}
	return nil
}
