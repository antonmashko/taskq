package taskq

import "sync"

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
