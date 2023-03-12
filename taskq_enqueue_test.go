package taskq_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/antonmashko/taskq"
)

type testQueue struct {
	sync.Mutex
	t          taskq.Task
	enqueueErr error
	dequeueErr error
}

func (q *testQueue) Enqueue(_ context.Context, t taskq.Task) (int64, error) {
	q.Lock()
	defer q.Unlock()
	if q.enqueueErr != nil {
		return -1, q.enqueueErr
	}
	q.t = t
	return 1, nil
}

func (q *testQueue) Dequeue(_ context.Context) (taskq.Task, error) {
	q.Lock()
	defer q.Unlock()
	if q.dequeueErr != nil {
		err := q.dequeueErr
		q.dequeueErr = nil
		return nil, err
	}
	if q.t == nil {
		return nil, taskq.EmptyQueue
	}
	return q.t, nil
}

func TestEnqueueNilTask_Err(t *testing.T) {
	tq := taskq.New(1)
	id, err := tq.Enqueue(context.Background(), nil)
	if err == nil || id != -1 {
		t.Fail()
	}
}

func TestEnqueueTaskAndStart_Ok(t *testing.T) {
	tq := taskq.New(1)
	var wg sync.WaitGroup
	var i int32
	wg.Add(1)
	task := taskq.TaskFunc(func(ctx context.Context) error {
		atomic.AddInt32(&i, 1)
		wg.Done()
		return nil
	})
	tq.Enqueue(context.Background(), task)
	tq.Start()
	wg.Wait()
	if i != 1 {
		t.Fail()
	}
}

func TestStartAndEnqueueTask_Ok(t *testing.T) {
	tq := taskq.New(1)
	tq.Start()
	var wg sync.WaitGroup
	var i int32
	wg.Add(1)
	task := taskq.TaskFunc(func(ctx context.Context) error {
		atomic.AddInt32(&i, 1)
		wg.Done()
		return nil
	})
	tq.Enqueue(context.Background(), task)
	wg.Wait()
	if i != 1 {
		t.Fail()
	}
}

func TestEnqueueFromMultipleGourotines_Ok(t *testing.T) {
	tq := taskq.New(0)
	tq.Start()
	var wg sync.WaitGroup
	const (
		goroutinesCount = int32(100)
		delta           = int32(1)
	)
	var result int32
	wg.Add(int(goroutinesCount))
	for i := int32(0); i < goroutinesCount; i++ {
		go func() {
			tq.Enqueue(context.Background(), taskq.TaskFunc(func(ctx context.Context) error {
				atomic.AddInt32(&result, delta)
				wg.Done()
				return nil
			}))
		}()
	}
	wg.Wait()
	if result != goroutinesCount*delta {
		t.Fatalf("invalid result. expected:%d actual:%d", goroutinesCount*delta, result)
	}
}

func TestEnqueueToClosingTaskq_Ok(t *testing.T) {
	tq := taskq.New(0)
	tq.Start()
	tq.Close()
	id, err := tq.Enqueue(context.Background(), taskq.TaskFunc(func(ctx context.Context) error { return nil }))
	if id != -1 || err != taskq.ErrClosed {
		t.Fail()
	}
}

func TestErrorFromEnqueue_Err(t *testing.T) {
	q := &testQueue{}
	tq := taskq.NewWithQueue(0, q)
	q.enqueueErr = errors.New("enqueue error")
	_, err := tq.Enqueue(context.Background(), taskq.TaskFunc(func(ctx context.Context) error { return nil }))
	if err != q.enqueueErr {
		t.Fatalf("invalid error. expected=%s got=%s", q.enqueueErr, err)
	}
}

func TestErrorFromDequeue_Err(t *testing.T) {
	q := &testQueue{}
	expectedErr := errors.New("dequeue error")
	q.dequeueErr = expectedErr
	tq := taskq.NewWithQueue(0, q)
	var wg sync.WaitGroup
	wg.Add(1)
	tq.OnDequeueError = func(ctx context.Context, workerID uint64, err error) {
		if err != q.dequeueErr {
			t.Errorf("invalid error. expected=%s got=%s", expectedErr, err)
		}
		wg.Done()
	}
	tq.Enqueue(context.Background(), taskq.TaskFunc(func(ctx context.Context) error { return nil }))
	tq.Start()
	wg.Wait()
}
