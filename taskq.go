package taskq

import (
	"context"
	"errors"
	"sync/atomic"
)

type TaskqType byte

const (
	Pool TaskqType = iota + 1
	Spawn
)

var (
	ErrClosed  = errors.New("taskq closed")
	ErrNilTask = errors.New("nil task")
)

type workersManager interface {
	Enqueue(Task) bool
	Start(context.Context)
	Shutdown(context.Context) error
}

type TaskQ struct {
	queue Queue

	isRunning      int32
	closed         int32
	workersManager workersManager

	OnDequeueError func(ctx context.Context, workerID int, err error)
}

func New(size int) *TaskQ {
	return NewWithQueue(size, NewConcurrentQueue())
}

func NewWithType(size int, tp TaskqType) *TaskQ {
	return NewWithTypeAndQueue(size, tp, NewConcurrentQueue())
}

func NewWithQueue(size int, q Queue) *TaskQ {
	return NewWithTypeAndQueue(size, Spawn, q)
}

func NewWithTypeAndQueue(size int, tp TaskqType, q Queue) *TaskQ {
	var workersManager workersManager
	switch tp {
	case Pool:
		workersManager = newWorkersPool(size, q)
	default:
		workersManager = newWorkersSpawner(size, q)
	}
	return &TaskQ{
		workersManager: workersManager,
		queue:          q,
	}
}

func (t *TaskQ) Enqueue(ctx context.Context, task Task) (int64, error) {
	if task == nil {
		return -1, ErrNilTask
	}

	if atomic.LoadInt32(&t.closed) != 0 {
		return -1, ErrClosed
	}

	id, err := t.queue.Enqueue(ctx, task)
	if err != nil {
		return -1, err
	}

	t.workersManager.Enqueue(task)
	return id, nil
}

func (t *TaskQ) Start() error {
	if !atomic.CompareAndSwapInt32(&t.isRunning, 0, 1) {
		return errors.New("taskq is running")
	}
	ctx := context.Background()
	t.workersManager.Start(ctx)
	return nil
}

func (t *TaskQ) Shutdown(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&t.closed, 0, 1) {
		return errors.New("taskq is closing")
	}
	return t.workersManager.Shutdown(ctx)
}

func (t *TaskQ) Close() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return t.Shutdown(ctx)
}
