package taskq

import (
	"context"
	"sync"
)

type TaskQInterface interface {
	Enqueue(Task) int64
	Start() error
	Close()
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

func (wg *WaitGroup) Enqueue(task Task) int64 {
	wg.wg.Add(1)
	return wg.TaskQInterface.Enqueue(TaskFunc(func(ctx context.Context) error {
		err := task.Do(ctx)
		wg.wg.Done()
		return err
	}))
}

func (wg *WaitGroup) Wait() {
	wg.wg.Wait()
}
