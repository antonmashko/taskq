package taskq

import (
	"context"
	"testing"
)

func TestTaskFromTaskMenegerOk(t *testing.T) {
	m := NewTaskManger(New(1))
	block := make(chan struct{})
	task := TaskFunc(func(ctx context.Context) error {
		<-block
		return nil
	})
	id := m.Enqueue(task)
	m.Start()
	mtask := m.Task(id)
	if mtask == nil {
		t.Error("task not found")
	}
	block <- struct{}{}
	m.Close()
}
