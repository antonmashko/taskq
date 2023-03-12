package taskq

import (
	"context"
	"errors"
	"sync"
)

var (
	EmptyQueue = errors.New("empty queue")
)

type Queue interface {
	Enqueue(context.Context, Task) (int64, error)
	// Dequeue Task from queue
	// if queue empty return `EmptyQueue` as error
	Dequeue(context.Context) (Task, error)
}

type ConcurrentQueue struct {
	lock    sync.Mutex
	lastInc int64
	queue   []Task
}

func NewConcurrentQueue() *ConcurrentQueue {
	return &ConcurrentQueue{}
}

func (q *ConcurrentQueue) Enqueue(_ context.Context, t Task) (int64, error) {
	var result int64
	q.lock.Lock()
	q.queue = append(q.queue, t)
	q.lastInc++
	result = q.lastInc
	q.lock.Unlock()
	return result, nil
}

func (q *ConcurrentQueue) Dequeue(_ context.Context) (Task, error) {
	q.lock.Lock()
	if len(q.queue) == 0 {
		q.lock.Unlock()
		return nil, EmptyQueue
	}
	it := q.queue[0]
	q.queue = q.queue[1:]
	q.lock.Unlock()
	return it, nil
}

func (q *ConcurrentQueue) Len(_ context.Context) int {
	return len(q.queue)
}
