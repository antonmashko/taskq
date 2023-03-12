package taskq

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"time"
)

var (
	ErrClosed  = errors.New("taskq closed")
	ErrNilTask = errors.New("nil task")
)

type TaskQ struct {
	queue Queue

	isRunning int32
	isClosing int32
	isStopped int32

	workerCount int32
	workers     chan *worker

	OnDequeueError func(ctx context.Context, workerID uint64, err error)
}

func New(limit int) *TaskQ {
	return NewWithQueue(limit, NewConcurrentQueue())
}

func NewWithQueue(limit int, q Queue) *TaskQ {
	if limit <= 0 {
		limit = runtime.NumCPU()
	}
	// init worker pool
	workers := make(chan *worker, limit)
	for i := 1; i < limit+1; i++ {
		workers <- &worker{
			id:     uint64(i),
			status: registered,
		}
	}
	return &TaskQ{
		queue:          q,
		isRunning:      0,
		isClosing:      0,
		isStopped:      0,
		workers:        workers,
		OnDequeueError: nil,
	}
}

func (t *TaskQ) triggerDequeue(ctx context.Context) bool {
	if atomic.LoadInt32(&t.isRunning) != 1 {
		return false
	}
	select {
	// trying to fetch free worker for immediate task execution
	case w, ok := <-t.workers:
		if !ok {
			return false
		}
		atomic.AddInt32(&t.workerCount, 1)
		go func(ctx context.Context, w *worker) {
			w.setStatus(live)
			for atomic.LoadInt32(&t.isStopped) != 1 {
				task, err := t.queue.Dequeue(ctx)
				if err != nil {
					if err == EmptyQueue {
						break
					}
					if t.OnDequeueError == nil {
						panic(err)
					}
					t.OnDequeueError(ctx, w.id, err)
					break
				}
				processTask(ctx, task)
			}
			w.setStatus(idle)
			t.workers <- w // return worker to pool
			atomic.AddInt32(&t.workerCount, -1)
		}(ctx, w)
		return true
	default:
		return false
	}
}

func (t *TaskQ) Enqueue(ctx context.Context, task Task) (int64, error) {
	if task == nil {
		return -1, ErrNilTask
	}

	if atomic.LoadInt32(&t.isClosing) != 0 {
		return -1, ErrClosed
	}

	id, err := t.queue.Enqueue(ctx, task)
	if err != nil {
		return -1, err
	}

	t.triggerDequeue(ctx)
	return id, nil
}

func (t *TaskQ) triggerFreeWorkers(ctx context.Context) {
	count := len(t.workers)
	for i := 0; i < count; i++ {
		if !t.triggerDequeue(ctx) {
			return
		}
	}
}

func (t *TaskQ) Start() error {
	if atomic.LoadInt32(&t.isClosing) != 0 {
		return ErrClosed
	}
	if !atomic.CompareAndSwapInt32(&t.isRunning, 0, 1) {
		return errors.New("taskq is running")
	}
	t.triggerFreeWorkers(context.Background())
	return nil
}

func (t *TaskQ) Shutdown(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&t.isClosing, 0, 1) {
		return errors.New("taskq is closing")
	}
	defer close(t.workers)
	if wait, _ := ctx.Value(ctxWaitKey{}).(bool); !wait {
		atomic.StoreInt32(&t.isStopped, 1)
	} else {
		t.triggerFreeWorkers(ctx)
	}

	var pollDuration = time.Millisecond * 500
	timer := time.NewTimer(pollDuration)
	for atomic.LoadInt32(&t.workerCount) > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			const delta = 1.1
			timer.Reset(time.Duration(float64(pollDuration) * delta))
		}
	}
	return nil
}

func (t *TaskQ) Close() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return t.Shutdown(ctx)
}
