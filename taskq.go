package taskq

import (
	"sync"
	"sync/atomic"
)

const (
	Pending byte = iota
	InProgress
	Done
	Failed
)

var WorkersPollSize = 10
var TaskMaxRetry = 3

type Task interface {
	Do() error
}

type itask struct {
	id    int64
	state byte
	task  Task
}

type TaskQ struct {
	lastInc       int64
	queue         chan *itask
	pending       *blockingQueue
	tasksMaxRetry int
	workersCount  int
	done          bool

	lock sync.Mutex

	TaskFailed func(Task)
	TaskDone   func(Task)
}

func New() *TaskQ {
	return &TaskQ{
		workersCount:  WorkersPollSize,
		tasksMaxRetry: TaskMaxRetry,
		queue:         make(chan *itask, WorkersPollSize),
		pending: &blockingQueue{
			queue: make([]*itask, 0, WorkersPollSize),
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
	select {
	case t.queue <- it:
		// if we can't send task to directly on workers
	default:
		// we add it to pending queue
		t.pending.enqeue(it)
	}
	return it.id
}

func (t *TaskQ) Start() error {
	// run process workers
	for i := 0; i < t.workersCount-1; i++ {
		// each worker will make task.Do
		go func(workerID int) {
			for task := range t.queue {
				for task != nil {
					t.process(task)
					task = t.pending.deqeue()
				}
			}
		}(i)
	}
	return nil
}

func (t *TaskQ) process(it *itask) {
	it.state = InProgress
	var try int
	for {
		if err := it.task.Do(); err != nil {
			if err == ErrRetryTask {
				try++
			}
			if try < t.tasksMaxRetry {
				continue
			}
			if t.TaskFailed != nil {
				t.TaskFailed(it.task)
			}
			break
		}
		it.state = Done
		if t.TaskDone != nil {
			t.TaskDone(it.task)
		}
		break
	}
}

func (t *TaskQ) Close() error {
	t.done = true
	close(t.queue)
	return nil
}

type blockingQueue struct {
	lock  sync.Mutex
	queue []*itask
}

func (q *blockingQueue) enqeue(it *itask) {
	q.lock.Lock()
	q.queue = append(q.queue, it)
	q.lock.Unlock()
}

func (q *blockingQueue) deqeue() *itask {
	q.lock.Lock()
	if len(q.queue) == 0 {
		q.lock.Unlock()
		return nil
	}
	it := q.queue[0]
	q.queue = q.queue[1:]
	q.lock.Unlock()
	return it
}
