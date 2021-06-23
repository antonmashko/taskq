package taskq

import (
	"context"
	"errors"
)

var ErrRetryTask = errors.New("error occurred during task execution. retry task")

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

type RetryableTask struct {
	task       Task
	maxRetries int
}

func NewRetryableTask(task Task, maxRetries int) Task {
	return &RetryableTask{task: task, maxRetries: maxRetries}
}

func (t *RetryableTask) Do(ctx context.Context) error {
	for i := 0; i < t.maxRetries; i++ {
		err := t.task.Do(context.Background())
		if err == nil {
			return nil
		}
		if err == ErrRetryTask {
			continue
		}
		return err
	}
	return nil
}
