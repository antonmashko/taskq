package taskq

import "errors"

var ErrRetryTask = errors.New("error occurred during task execution. retry task")
var ErrFailTask = errors.New("task failed")
var ErrMaxRetry = errors.New("max retries exceeded for task")
