package taskq

import (
	"context"
	"errors"
	"sync/atomic"
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

	closed         int32
	workersManager workersManager

	OnDequeueError func(ctx context.Context, workerID int, err error)
}

func New(size int) *TaskQ {
	return NewWithQueue(size, NewConcurrentQueue())
}

func NewWithQueue(size int, q Queue) *TaskQ {
	return &TaskQ{
		workersManager: newWorkersPool(size, q),
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
	ctx := context.Background()
	t.workersManager.Start(ctx)
	return nil
}

func (t *TaskQ) Shutdown(ctx context.Context) error {
	atomic.StoreInt32(&t.closed, 1) // no more accepting tasks
	return t.workersManager.Shutdown(ctx)
}

func (t *TaskQ) Close() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return t.Shutdown(ctx)
}
