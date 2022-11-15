package taskq

import (
	"context"
	"runtime"
)

type workersPool struct {
	size int
}

func newWorkersPool(size int) *workersPool {
	if size < 1 {
		size = runtime.NumCPU()
	}
	return &workersPool{
		size: size,
	}
}

func (wp *workersPool) Len() int {
	return wp.size
}

func (wp *workersPool) Run(ctx context.Context, task Task) {
	processTask(ctx, task)
}

func (wp *workersPool) Shutdown(ctx context.Context) error {
	return nil
}
