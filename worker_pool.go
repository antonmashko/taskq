package taskq

import (
	"context"
	"runtime"
)

type workerPool struct {
	size int
}

func newWorkerPool(size int) *workerPool {
	if size < 1 {
		size = runtime.NumCPU()
	}
	return &workerPool{
		size: size,
	}
}

func (wp *workerPool) Len() int {
	return wp.size
}

func (wp *workerPool) Run(ctx context.Context, task Task) {
	processTask(ctx, task)
}

func (wp *workerPool) Shutdown(ctx context.Context) error {
	return nil
}
