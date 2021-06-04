package taskq

import (
	"testing"
)

func TestBlockingQueueEnqueueOK(t *testing.T) {
	q := NewConcurrentQueue()
	for i := 0; i < 10; i++ {
		q.Enqueue(&itask{})
	}
	if len(q.(*ConcurrentQueue).queue) != 10 {
		t.Fail()
	}
}

func TestBlockingQueueDequeueOK(t *testing.T) {
	q := NewConcurrentQueue()
	t1 := &itask{}
	q.Enqueue(t1)
	t2 := q.Dequeue()
	if t1 != t2 {
		t.Fail()
	}
}

func BenchmarkBlockingQueue(b *testing.B) {
	q := NewConcurrentQueue()
	// q.queue = append(q.queue, &itask{})
	for i := 0; i < b.N; i++ {
		q.Enqueue(&itask{})
		q.Dequeue()
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
