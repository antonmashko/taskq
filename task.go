package taskq

import (
	"context"
)

// Task for TaskQ
type Task interface {
	Do(ctx context.Context) error
}

type TaskDone interface {
	Done(context.Context)
}

type TaskOnError interface {
	OnError(context.Context, error)
}

type TaskFunc func(ctx context.Context) error

func (t TaskFunc) Do(ctx context.Context) error {
	return t(ctx)
}

func processTask(ctx context.Context, task Task) {
	err := task.Do(ctx)
	if err != nil {
		if event, ok := task.(TaskOnError); ok && event != nil {
			event.OnError(ctx, err)
		}
		return
	}

	if event, ok := task.(TaskDone); ok && event != nil {
		event.Done(ctx)
	}
}
