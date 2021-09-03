package taskq

import (
	"context"
	"sync"
)

type WaitGroup struct {
	*TaskQ
	wg sync.WaitGroup
}

func NewWaitGroup(size int) *WaitGroup {
	return &WaitGroup{
		TaskQ: New(size),
	}
}

func NewWaitGroupFromTaskq(taskq *TaskQ) *WaitGroup {
	return &WaitGroup{
		TaskQ: taskq,
	}
}

func (wg *WaitGroup) Enqueue(ctx context.Context, task Task) (int64, error) {
	wg.wg.Add(1)
	return wg.TaskQ.Enqueue(ctx, TaskFunc(func(ctx context.Context) error {
		err := task.Do(ctx)
		wg.wg.Done()
		return err
	}))
}

func (wg *WaitGroup) Wait() {
	wg.wg.Wait()
}
