package taskq_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/antonmashko/taskq"
)

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

func TestPoolEnqueueNilTask_Err(t *testing.T) {
	tq := taskq.Pool(1)
	id, err := tq.Enqueue(context.Background(), nil)
	if err == nil || id != -1 {
		t.Fail()
	}
}

func TestPoolEnqueueTaskAndStart_Ok(t *testing.T) {
	tq := taskq.Pool(1)
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

func TestPoolStartAndEnqueueTask_Ok(t *testing.T) {
	tq := taskq.Pool(1)
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

func TestPoolEnqueueFromMultipleGourotines_Ok(t *testing.T) {
	tq := taskq.Pool(0)
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
