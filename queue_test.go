package taskq

import (
	"testing"
)

func TestBlockingQueueEnqueueOK(t *testing.T) {
	q := &blockingQueue{queue: make([]*itask, 0, 10)}
	for i := 0; i < 10; i++ {
		q.enqueue(&itask{})
	}
	if len(q.queue) != 10 {
		t.Fail()
	}
}

func TestBlockingQueueDequeueOK(t *testing.T) {
	q := &blockingQueue{queue: make([]*itask, 0, 1)}
	t1 := &itask{}
	q.enqueue(t1)
	t2 := q.dequeue()
	if t1 != t2 {
		t.Fail()
	}

}

func BenchmarkBlockingQueue(b *testing.B) {
	q := &blockingQueue{queue: make([]*itask, 0, 1)}
	q.queue = append(q.queue, &itask{})
	for i := 0; i < b.N; i++ {
		q.enqueue(&itask{})
		q.dequeue()
	}
}

// func BenchmarkChannel(b *testing.B) {
// 	ch := make(chan *itask, 1)
// 	for i := 0; i < b.N; i++ {
// 		ch <- &itask{}
// 		<-ch
// 	}
// }

// func BenchmarkRingBlockingQueue(b *testing.B) {
// 	q := &ringBlockingQueue{queue: make([]*itask, 0, 1)}
// 	q.queue = append(q.queue, &itask{})
// 	for i := 0; i < b.N; i++ {
// 		q.enqueue(&itask{})
// 		q.dequeue()
// 	}
// }

// func BenchmarkLinkedBlockingQueue(b *testing.B) {
// 	q := &linkedBlockedQueue{}
// 	for i := 0; i < b.N; i++ {
// 		q.enqueue(&itask{})
// 		q.dequeue()
// 	}
// }
