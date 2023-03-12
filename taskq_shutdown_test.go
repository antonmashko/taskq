package taskq_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/antonmashko/taskq"
)

func TestGracefulShutdownWithWait_Ok(t *testing.T) {
	res := int32(0)
	tq := taskq.New(0)
	if err := tq.Start(); err != nil {
		panic(err)
	}

	tf := taskq.TaskFunc(func(ctx context.Context) error {
		time.Sleep(35 * time.Millisecond)
		atomic.AddInt32(&res, 1)
		return nil
	})
	expected := 100
	for i := 0; i < expected; i++ {
		if _, err := tq.Enqueue(context.Background(), tf); err != nil {
			panic(err)
		}
	}
	if err := tq.Shutdown(taskq.ContextWithWait(context.Background())); err != nil {
		panic(err)
	}

	if atomic.LoadInt32(&res) != int32(expected) {
		t.Fail()
	}
}

func TestGracefulShutdownWithoutWait_Ok(t *testing.T) {
	res := int32(0)
	tq := taskq.New(10)
	tf := taskq.TaskFunc(func(ctx context.Context) error {
		time.Sleep(2 * time.Second)
		atomic.AddInt32(&res, 1)
		return nil
	})
	if err := tq.Start(); err != nil {
		panic(err)
	}

	expected := 100
	for i := 0; i < expected; i++ {
		if _, err := tq.Enqueue(context.Background(), tf); err != nil {
			panic(err)
		}
	}
	if err := tq.Close(); err != nil {
		panic(err)
	}

	// taskq should not complete all tasks from
	// if result equal to expected than graceful should working incorrect
	if atomic.LoadInt32(&res) == int32(expected) {
		t.Fail()
	}
}

func TestGracefulShutdownWithTimeout_Ok(t *testing.T) {
	var result bool
	tq := taskq.New(10)
	ch := make(chan struct{})
	tf := taskq.TaskFunc(func(ctx context.Context) error {
		ch <- struct{}{}
		time.Sleep(10 * time.Second)
		result = true
		return nil
	})
	if err := tq.Start(); err != nil {
		panic(err)
	}
	if _, err := tq.Enqueue(context.Background(), tf); err != nil {
		panic(err)
	}
	<-ch
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := tq.Shutdown(ctx)
	if err == nil {
		t.Fatal("Shutdown return nil err")
	}

	if err != context.DeadlineExceeded {
		t.Fatalf("invalid error. expected: %T actual %T", err, context.DeadlineExceeded)
	}

	if result {
		t.Fail()
	}
}

func TestCloseAfterClose_Err(t *testing.T) {
	tq := taskq.New(0)
	err := tq.Close()
	if err != nil {
		t.Fatal("error on first close:", err)
	}
	if err = tq.Close(); err != taskq.ErrClosed {
		t.Fatalf("invalid error. expected=%s got=%s", taskq.ErrClosed, err)
	}
}
