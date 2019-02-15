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
