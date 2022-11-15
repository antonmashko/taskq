package taskq_test

import (
	"context"
	"sync"
	"testing"

	"github.com/antonmashko/taskq"
)

func TestNew_Ok(t *testing.T) {
	if tq := taskq.New(0); tq == nil {
		t.Fail()
	}
}

func TestPool_Ok(t *testing.T) {
	if tq := taskq.Pool(0); tq == nil {
		t.Fail()
	}
}

func TestEnqueueUniqueIDOk(t *testing.T) {
	tq := taskq.New(1)
	var wg sync.WaitGroup
	unique := make(map[int64]struct{})
	const count = 100
	wg.Add(count)
	for i := 0; i < count; i++ {
		id, err := tq.Enqueue(context.Background(), taskq.TaskFunc(func(ctx context.Context) error {
			wg.Done()
			return nil
		}))
		if err != nil {
			t.Fatalf("failed to enqueue. err=%v", err)
		}
		if _, ok := unique[id]; ok {
			t.Fail()
		} else {
			unique[id] = struct{}{}
		}
	}

	tq.Start()
	wg.Wait()
	tq.Close()
}
