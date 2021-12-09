package taskq

import (
	"context"
	"testing"
)

func TestBlockingQueueEnqueueOK(t *testing.T) {
	q := NewConcurrentQueue()
	for i := 0; i < 10; i++ {
		q.Enqueue(context.Background(), &testTask{})
	}
	if len(q.queue) != 10 {
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
	_, err := q.Enqueue(context.Background(), t1)
	if err != nil {
		t.Fatal("failed to enqueue. err=", err)
	}
	t2, err := q.Dequeue(context.Background())
	if err != nil {
		t.Fatal("failed to dequeue. err=", err)
	}
	err = t2.Do(context.Background())
	if err != nil || result != 1 {
		t.Fail()
	}
}

func BenchmarkBlockingQueue(b *testing.B) {
	q := NewConcurrentQueue()
	// q.queue = append(q.queue, &itask{})
	for i := 0; i < b.N; i++ {
		q.Enqueue(context.Background(), TaskFunc(func(ctx context.Context) error { return nil }))
		q.Dequeue(context.Background())
	}
}

func TestTaskqQueueImplementation(t *testing.T) {
	var _ Queue = NewConcurrentQueue()
	var _ Queue = NewLimitedConcurrentQueue(0)
}
