package taskq_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/antonmashko/taskq"
)

type testTask struct {
	fOnError func(context.Context, error)
	fdone    func(context.Context)

	resultErr error
}

func (t *testTask) OnError(ctx context.Context, err error) {
	t.fOnError(ctx, err)
}

func (t *testTask) Done(ctx context.Context) {
	t.fdone(ctx)
}

func (t *testTask) Do(_ context.Context) error {
	return t.resultErr
}

func TestTaskqStartNoError_Ok(t *testing.T) {
	tq := taskq.New(0)
	if err := tq.Start(); err != nil {
		t.Fail()
	}
}

func TestTaskqDoubleStart_Err(t *testing.T) {
	tq := taskq.New(0)
	if err := tq.Start(); err != nil {
		t.Fail()
	}
	if err := tq.Start(); err == nil {
		t.Fail()
	}
}

func TestTaskqStartOnNotEmptyQueue_Ok(t *testing.T) {
	q := taskq.NewConcurrentQueue()
	var wg sync.WaitGroup
	const count = 100
	wg.Add(count)
	for i := 0; i < count; i++ {
		_, err := q.Enqueue(context.Background(), taskq.TaskFunc(func(ctx context.Context) error {
			wg.Done()
			return nil
		}))
		if err != nil {
			t.Fatalf("got error on enqueue. err:%s", err)
		}
	}

	tq := taskq.NewWithQueue(0, q)
	tq.Start()
	wg.Wait()
}

func TestPoolStartOnNotEmptyQueue_Ok(t *testing.T) {
	q := taskq.NewConcurrentQueue()
	var wg sync.WaitGroup
	const count = 100
	wg.Add(count)
	for i := 0; i < count; i++ {
		_, err := q.Enqueue(context.Background(), taskq.TaskFunc(func(ctx context.Context) error {
			wg.Done()
			return nil
		}))
		if err != nil {
			t.Fatalf("got error on enqueue. err:%s", err)
		}
	}
	tq := taskq.PoolWithQueue(0, q)
	tq.Start()
	wg.Wait()
}

func TestTaskqSequentialExecution_Ok(t *testing.T) {
	tq := taskq.New(1)
	tq.Start()
	count := 100
	counter := int32(0)
	ch := make(chan int)
	for i := 0; i < count; i++ {
		tq.Enqueue(context.Background(), taskq.TaskFunc(func(ctx context.Context) error {
			ch <- int(atomic.AddInt32(&counter, 1))
			return nil
		}))
	}
	tq.Enqueue(context.Background(), taskq.TaskFunc(func(ctx context.Context) error {
		close(ch)
		return nil
	}))

	curr := 0
	for i := range ch {
		curr++
		if i != curr {
			t.Fatalf("invalid value from channel. expected:%d actual:%d", curr, i)
		}
	}
}

func TestPoolSequentialExecution_Ok(t *testing.T) {
	tq := taskq.Pool(1)
	tq.Start()
	count := 100
	counter := int32(0)
	ch := make(chan int)
	for i := 0; i < count; i++ {
		tq.Enqueue(context.Background(), taskq.TaskFunc(func(ctx context.Context) error {
			ch <- int(atomic.AddInt32(&counter, 1))
			return nil
		}))
	}
	tq.Enqueue(context.Background(), taskq.TaskFunc(func(ctx context.Context) error {
		close(ch)
		return nil
	}))

	curr := 0
	for i := range ch {
		curr++
		if i != curr {
			t.Fatalf("invalid value from channel. expected:%d actual:%d", curr, i)
		}
	}
}
