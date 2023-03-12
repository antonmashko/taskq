package taskq_test

import (
	"context"
	"testing"

	"github.com/antonmashko/taskq"
)

func TestBlockingQueueEnqueueOK(t *testing.T) {
	q := taskq.NewConcurrentQueue()
	for i := 0; i < 10; i++ {
		q.Enqueue(context.Background(), &testTask{})
	}
	if q.Len(context.Background()) != 10 {
		t.Fail()
	}
}

func TestBlockingQueueDequeueOK(t *testing.T) {
	q := taskq.NewConcurrentQueue()
	var result int
	t1 := taskq.TaskFunc(func(ctx context.Context) error {
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

func TestTaskqQueueImplementation(t *testing.T) {
	var _ taskq.Queue = taskq.NewConcurrentQueue()
}
