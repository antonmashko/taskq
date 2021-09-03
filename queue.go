package taskq

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
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

func NewConcurrentQueue() Queue {
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

type LimitedConcurrentQueue struct {
	lastInc int64
	ch      chan Task
}

func NewLimitedConcurrentQueue(size int) Queue {
	return &LimitedConcurrentQueue{ch: make(chan Task, size)}
}

func (q *LimitedConcurrentQueue) Enqueue(_ context.Context, t Task) (int64, error) {
	q.ch <- t
	return atomic.AddInt64(&q.lastInc, 1), nil
}

func (q *LimitedConcurrentQueue) Dequeue(ctx context.Context) (Task, error) {
	select {
	case t := <-q.ch:
		if t == nil {
			return nil, EmptyQueue
		}
		return t, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
