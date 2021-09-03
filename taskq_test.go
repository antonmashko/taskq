package taskq

import (
	"context"
	"errors"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func TestTaskStatusOk(t *testing.T) {
	if TaskStatus(&itask{status: Pending}) != Pending {
		t.Fail()
	}

}

func TestTaskStatusErr(t *testing.T) {
	if TaskStatus(nil) != None {
		t.Fail()
	}
}

func TestNilITaskStatusErr(t *testing.T) {
	var it *itask
	if TaskStatus(it) != None {
		t.Fail()
	}
}

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
	if id, err := tq.Enqueue(context.Background(), nil); err == nil || id != -1 {
		t.Fail()
	}
}

func TestEnqueueTaskOk(t *testing.T) {
	tq := New(1)
	var wg sync.WaitGroup
	var i int32
	wg.Add(1)
	tq.Enqueue(context.Background(), TaskFunc(func(ctx context.Context) error {
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
		id, err := tq.Enqueue(context.Background(), TaskFunc(func(ctx context.Context) error {
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

func TestDoneCallbackOk(t *testing.T) {
	tq := New(1)
	var ok bool
	task := TaskFunc(func(ctx context.Context) error { return nil })
	tq.TaskDone = func(_ int64, t Task) {
		if reflect.ValueOf(t).Pointer() == reflect.ValueOf(task).Pointer() {
			ok = true
		}
	}
	tq.process(context.Background(), &itask{id: 0, task: task})
	if !ok {
		t.Fail()
	}
}

func TestFailedCallbackOk(t *testing.T) {
	tq := New(1)
	var ok bool
	var cErr = errors.New("error")
	task := TaskFunc(func(ctx context.Context) error { return cErr })
	tq.TaskFailed = func(_ int64, t Task, err error) {
		if reflect.ValueOf(t).Pointer() == reflect.ValueOf(task).Pointer() && err == cErr {
			ok = true
		}
	}
	tq.process(context.Background(), &itask{id: 0, task: task})
	if !ok {
		t.Fail()
	}
}

type testTask struct {
	fOnError func(context.Context, int64, error)
	fdone    func(context.Context, int64)

	resultErr error
}

func (t *testTask) OnError(ctx context.Context, id int64, err error) {
	t.fOnError(ctx, id, err)
}

func (t *testTask) Done(ctx context.Context, id int64) {
	t.fdone(ctx, id)
}

func (t *testTask) Do(_ context.Context) error {
	return t.resultErr
}

func TestTaskDoneEvent(t *testing.T) {
	rch := make(chan bool)
	tt := &testTask{
		fdone: func(ctx context.Context, i int64) {
			rch <- true
		},
	}

	tq := New(1)
	tq.Start()
	tq.Enqueue(context.Background(), tt)

	if res := <-rch; !res {
		t.Fail()
	}
}

func TestTaskOnErrorEvent(t *testing.T) {
	rch := make(chan error)
	err := errors.New("task failed")
	tt := &testTask{
		resultErr: err,
		fOnError: func(c context.Context, i int64, e error) {
			rch <- e
		},
		fdone: func(ctx context.Context, i int64) {
			panic("done event invoked")
		},
	}

	tq := New(1)
	tq.Start()
	tq.Enqueue(context.Background(), tt)

	if tErr := <-rch; tErr != err {
		t.Fail()
	}
}
