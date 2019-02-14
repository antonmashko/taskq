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
	queue         chan struct{}
	pending       *blockingQueue
	tasksMaxRetry int
	workersCount  int

	lock sync.Mutex

	TaskDone   func(Task)
	TaskFailed func(Task, error)
}

func New() *TaskQ {
	return &TaskQ{
		workersCount:  WorkersPollSize,
		tasksMaxRetry: TaskMaxRetry,
		queue:         make(chan struct{}, WorkersPollSize),
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
	t.pending.enqueue(it)
	select {
	case t.queue <- struct{}{}:
		// successfully sent to the workers
	default:
		// if we can't send task to directly to the workers
		// we add it to pending queue
	}
	return it.id
}

func (t *TaskQ) Start() error {
	// run process workers
	for i := 0; i < t.workersCount-1; i++ {
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
	var try int
	for {
		if err := it.task.Do(); err != nil {
			if err == ErrRetryTask {
				try++
			}
			if try < t.tasksMaxRetry {
				continue
			} else {
				err = ErrMaxRetry
			}
			if t.TaskFailed != nil {
				t.TaskFailed(it.task, err)
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
	close(t.queue)
	return nil
}

type blockingQueue struct {
	lock  sync.Mutex
	queue []*itask
}

func (q *blockingQueue) enqueue(it *itask) {
	q.lock.Lock()
	q.queue = append(q.queue, it)
	q.lock.Unlock()
}

func (q *blockingQueue) dequeue() *itask {
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

type ringBlockingQueue struct {
	lock     sync.Mutex
	queue    []*itask
	writeIdx int
	readIdx  int
}

func (q *ringBlockingQueue) enqueue(it *itask) {
	q.lock.Lock()
	q.writeIdx = q.writeIdx % len(q.queue)
	q.queue[q.writeIdx] = it
	q.writeIdx++
	q.lock.Unlock()
}

func (q *ringBlockingQueue) dequeue() *itask {
	q.lock.Lock()
	if len(q.queue) == 0 {
		q.lock.Unlock()
		return nil
	}
	q.readIdx = q.readIdx % len(q.queue)
	it := q.queue[q.readIdx]
	if it == nil {
		q.lock.Unlock()
		return nil
	}
	q.readIdx++
	q.lock.Unlock()
	return it
}

type qitem struct {
	it   *itask
	next *qitem
}

type linkedBlockedQueue struct {
	lock  sync.Mutex
	first *qitem
	last  *qitem
}

func (q *linkedBlockedQueue) enqueue(it *itask) {
	q.lock.Lock()
	if q.last == nil {
		q.last = &qitem{it: it}
	} else {
		q.last.next = &qitem{it: it}
		q.last = q.last.next
	}
	if q.first == nil {
		q.first = q.last
	}
	q.lock.Unlock()
}

func (q *linkedBlockedQueue) dequeue() *itask {
	q.lock.Lock()
	r := q.first
	if r == nil {
		q.lock.Unlock()
		return nil
	}
	q.first = q.first.next
	q.lock.Unlock()
	return r.it
}
