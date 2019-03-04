package taskq

import (
	"sync"
	"sync/atomic"
)

const (
	Ignored byte = iota
	Pending
	InProgress
	Done
	Failed
)

type itask struct {
	id    int64
	state byte
	task  Task
}

type TaskQ struct {
	lastInc int64
	queue   chan struct{}
	pending *blockingQueue
	//tasksMaxRetry int
	size int

	lock sync.Mutex

	TaskDone   func(Task)
	TaskFailed func(Task, error)
}

func New(size int) *TaskQ {
	if size < 0 {
		return nil
	}
	return &TaskQ{
		size:  size,
		queue: make(chan struct{}, size),
		pending: &blockingQueue{
			queue: make([]*itask, 0, size),
		},
	}
}

func (t *TaskQ) Enqueue(task Task) int64 {
	if task == nil {
		return -1
	}
	it := &itask{
		id:    atomic.AddInt64(&t.lastInc, 1),
		state: Pending,
		task:  task,
	}
	t.pending.enqueue(it)
	// notify worker about pending task
	select {
	case t.queue <- struct{}{}:
		// we have free worker
	default:
		// all worker's are busy
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
	it.state = InProgress
	err := it.task.Do()
	if err != nil {
		it.state = Failed
		if t.TaskFailed != nil {
			t.TaskFailed(it.task, err)
		}
	}
	it.state = Done
	if t.TaskDone != nil {
		t.TaskDone(it.task)
	}
}

func (t *TaskQ) Close() error {
	close(t.queue)
	return nil
}
