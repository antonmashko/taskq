package taskq

import "sync"

type Queue interface {
	Enqueue(Task)
	Dequeue() Task
}
type ConcurrentQueue struct {
	lock  sync.Mutex
	queue []Task
}

func NewConcurrentQueue() Queue {
	return &ConcurrentQueue{}
}

func (q *ConcurrentQueue) Enqueue(t Task) {
	q.lock.Lock()
	q.queue = append(q.queue, t)
	q.lock.Unlock()
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
	ch chan Task
}

func NewLimitedConcurrentQueue(size int) Queue {
	return &LimitedConcurrentQueue{ch: make(chan Task, size)}
}

func (q *LimitedConcurrentQueue) Enqueue(t Task) {
	q.ch <- t
}

func (q *LimitedConcurrentQueue) Dequeue() Task {
	return <-q.ch
}
