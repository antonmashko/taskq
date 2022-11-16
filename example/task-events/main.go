package main

import (
	"context"
	"errors"
	"log"

	"github.com/antonmashko/taskq"
)

type task struct {
	result int
	err    error
}

func (t *task) Do(ctx context.Context) error {
	return t.err
}

func (t *task) Done(ctx context.Context) {
	log.Println("[Done] task result:", t.result)
}

func (t *task) OnError(ctx context.Context, err error) {
	log.Println("[OnError] task err:", err)
}

/*
	$ go run main.go
	Output:
	[Done] task result: 1
	[OnError] task err: foo err
*/
func main() {
	tq := taskq.New(0)
	tq.Start()

	tq.Enqueue(context.Background(), &task{result: 1})
	tq.Enqueue(context.Background(), &task{err: errors.New("foo err")})

	tq.Shutdown(taskq.ContextWithWait(context.Background()))
}
