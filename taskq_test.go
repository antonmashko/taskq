package taskq

import (
	"testing"
)

func BenchmarkChannel(b *testing.B) {
	ch := make(chan *itask, 1)
	for i := 0; i < b.N; i++ {
		ch <- &itask{}
		<-ch
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

// func TestBlockingQueue(t *testing.T) {
// 	q := &blockingQueue{queue: make([]*itask, 0, WorkersPollSize)}
// 	for i := 0; i < 10; i++ {
// 		q.enqueue(&itask{})
// 	}
// 	if len(q.queue) != 10 {
// 		t.Error("queue not woking")
// 	}
// }
