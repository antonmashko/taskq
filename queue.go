package taskq

import (
	"sync"
	"sync/atomic"
)

type Queue interface {
	Enqueue(Task) int64
	Dequeue() Task
}
type ConcurrentQueue struct {
	lock   sync.Mutex
	queue  []Task
	lastId int64
}

func NewConcurrentQueue() Queue {
	return &ConcurrentQueue{}
}

func (q *ConcurrentQueue) Enqueue(t Task) int64 {
	q.lock.Lock()
	q.queue = append(q.queue, t)
	q.lock.Unlock()
	return atomic.AddInt64(&q.lastId, 1)
}

func (q *ConcurrentQueue) Dequeue() Task {
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

type LimitedConcurrentQueue struct {
	ch     chan Task
	lastId int64
}

func NewLimitedConcurrentQueue(size int) Queue {
	return &LimitedConcurrentQueue{ch: make(chan Task, size)}
}

func (q *LimitedConcurrentQueue) Enqueue(t Task) int64 {
	q.ch <- t
	return atomic.AddInt64(&q.lastId, 1)
}

func (q *LimitedConcurrentQueue) Dequeue() Task {
	return <-q.ch
}
