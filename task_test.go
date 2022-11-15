package taskq_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/antonmashko/taskq"
)

func TestTaskDoneEvent(t *testing.T) {
	rch := make(chan bool)
	tt := &testTask{
		fdone: func(ctx context.Context) {
			rch <- true
		},
	}

	tq := taskq.New(1)
	tq.Start()
	tq.Enqueue(context.Background(), tt)

	select {
	case <-rch:
		break
	case <-time.After(time.Second):
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

	tq := taskq.New(1)
	tq.Start()
	tq.Enqueue(context.Background(), tt)

	select {
	case tErr := <-rch:
		if tErr != err {
			t.Fail()
		}
		break
	case <-time.After(time.Second):
		t.Fail()
	}
}
