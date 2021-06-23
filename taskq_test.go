package taskq

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNewTaskQOk(t *testing.T) {
	if tq := New(1); tq == nil {
		t.Fail()
	}
}

func TestNewTaskQNumCPUOk(t *testing.T) {
	if tq := New(0); tq == nil || len(tq.workers) != runtime.NumCPU() {
		t.Fail()
	}
}

func TestEnqueueNilTaskErr(t *testing.T) {
	tq := New(1)
	if tq.Enqueue(nil) != -1 {
		t.Fail()
	}
}

func TestEnqueueTaskOk(t *testing.T) {
	tq := New(1)
	var wg sync.WaitGroup
	var i int32
	wg.Add(1)
	tq.Enqueue(TaskFunc(func(ctx context.Context) error {
		atomic.AddInt32(&i, 1)
		wg.Done()
		return nil
	}))
	tq.Start()
	tq.Close()
	wg.Wait()
	if i != 1 {
		t.Fail()
	}
}

func TestEnqueueUniqueIDOk(t *testing.T) {
	tq := New(1)
	var wg sync.WaitGroup
	unique := make(map[int64]struct{})
	const count = 100
	wg.Add(count)
	for i := 0; i < count; i++ {
		id := tq.Enqueue(TaskFunc(func(ctx context.Context) error {
			wg.Done()
			return nil
		}))
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

func TestTaskDoneEvent(t *testing.T) {
	rch := make(chan bool)
	tt := &testTask{
		fdone: func(ctx context.Context) {
			rch <- true
		},
	}

	tq := New(1)
	tq.Start()
	tq.Enqueue(tt)

	if res := <-rch; !res {
		t.Fail()
	}
}

func TestTaskOnErrorEvent(t *testing.T) {
	rch := make(chan error)
	err := errors.New("task failed")
	tt := &testTask{
		resultErr: err,
		fOnError: func(c context.Context, e error) {
			rch <- e
		},
		fdone: func(ctx context.Context) {
			panic("done event invoked")
		},
	}

	tq := New(1)
	tq.Start()
	tq.Enqueue(tt)

	if tErr := <-rch; tErr != err {
		t.Fail()
	}
}
