package taskq

import (
	"context"
	"sync"
)

type TaskQInterface interface {
	Enqueue(context.Context, Task) (int64, error)
	Start() error
	Shutdown(ctx context.Context) error
	Close() error
}

type WaitGroup struct {
	TaskQInterface
	wg sync.WaitGroup
}

func NewWaitGroup(size int) *WaitGroup {
	return &WaitGroup{
		TaskQInterface: New(size),
	}
}

func ConvertToWaitGroup(taskq TaskQInterface) *WaitGroup {
	return &WaitGroup{
		TaskQInterface: taskq,
	}
}

func (wg *WaitGroup) Enqueue(ctx context.Context, task Task) (int64, error) {
	wg.wg.Add(1)
	return wg.TaskQInterface.Enqueue(ctx, TaskFunc(func(ctx context.Context) error {
		err := task.Do(ctx)
		wg.wg.Done()
		return err
	}))
}

func (wg *WaitGroup) Wait() {
	wg.wg.Wait()
}
