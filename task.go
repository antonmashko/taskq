package taskq

import "errors"

var ErrRetryTask = errors.New("error occurred during task execution. retry task")

// Task for TaskQ
type Task interface {
	Do() error
}

type TaskFunc func() error

func (t TaskFunc) Do() error {
	return t()
}

type RetryableTask struct {
	task       Task
	maxRetries int
}

func NewRetryableTask(task Task, maxRetries int) Task {
	return &RetryableTask{task: task, maxRetries: maxRetries}
}

func (t *RetryableTask) Do() error {
	for i := 0; i < t.maxRetries; i++ {
		err := t.task.Do()
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
