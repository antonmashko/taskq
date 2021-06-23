package taskq

import (
	"context"
	"testing"
)

func TestBlockingQueueEnqueueOK(t *testing.T) {
	q := NewConcurrentQueue()
	for i := 0; i < 10; i++ {
		q.Enqueue(TaskFunc(func(ctx context.Context) error { return nil }))
	}
	if len(q.(*ConcurrentQueue).queue) != 10 {
		t.Fail()
	}
}

func TestBlockingQueueDequeueOK(t *testing.T) {
	q := NewConcurrentQueue()
	var result int
	t1 := TaskFunc(func(ctx context.Context) error {
		result++
		return nil
	})
	q.Enqueue(t1)
	t2 := q.Dequeue()
	err := t2.Do(context.Background())
	if err != nil || result != 1 {
		t.Fail()
	}
}

func BenchmarkBlockingQueue(b *testing.B) {
	q := NewConcurrentQueue()
	// q.queue = append(q.queue, &itask{})
	for i := 0; i < b.N; i++ {
		q.Enqueue(TaskFunc(func(ctx context.Context) error { return nil }))
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
