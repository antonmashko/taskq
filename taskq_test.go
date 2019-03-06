package taskq

import (
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
	if tq := New(0); tq == nil || tq.size != runtime.NumCPU() {
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
	tq.Enqueue(TaskFunc(func() error {
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

func TestEnqueuUniqueIDOk(t *testing.T) {
	tq := New(1)
	var wg sync.WaitGroup
	unique := make(map[int64]struct{})
	const count = 100
	wg.Add(count)
	for i := 0; i < count; i++ {
		id := tq.Enqueue(TaskFunc(func() error {
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
	tq.Close()
	wg.Wait()
}

func TestDoneCallbackOk(t *testing.T) {
	tq := New(1)
	var ok bool
	task := TaskFunc(func() error { return nil })
	tq.TaskDone = func(t Task) {
		if reflect.ValueOf(t).Pointer() == reflect.ValueOf(task).Pointer() {
			ok = true
		}
	}
	tq.process(&itask{id: 0, task: task})
	if !ok {
		t.Fail()
	}
}

func TestFailedCallbackOk(t *testing.T) {
	tq := New(1)
	var ok bool
	var cErr = errors.New("error")
	task := TaskFunc(func() error { return cErr })
	tq.TaskFailed = func(t Task, err error) {
		if reflect.ValueOf(t).Pointer() == reflect.ValueOf(task).Pointer() && err == cErr {
			ok = true
		}
	}
	tq.process(&itask{id: 0, task: task})
	if !ok {
		t.Fail()
	}
}
