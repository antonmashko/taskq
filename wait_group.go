package taskq

import "context"

type WaitGroup struct {
	tq *TaskQ
}

func NewWaitGroup(size int) *WaitGroup {
	tq := New(size)
	err := tq.Start()
	if err != nil {
		panic(err)
	}
	return &WaitGroup{
		tq: tq,
	}
}

func (wg *WaitGroup) Enqueue(ctx context.Context, t Task) (int64, error) {
	return wg.tq.Enqueue(ctx, t)
}

func (wg *WaitGroup) Wait() {
	wg.tq.Shutdown(ContextWithWait(context.Background()))
}
