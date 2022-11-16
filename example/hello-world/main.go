package main

import (
	"context"
	"fmt"
	"log"

	"github.com/antonmashko/taskq"
)

type Task struct{}

func (Task) Do(ctx context.Context) error {
	fmt.Println("hello world")
	return nil
}

func main() {
	// Initializing new TaskQ instance with a limit of 10 max active goroutines
	// Use `limit=0` for not limiting goroutines number.
	tq := taskq.New(10)

	// Starting reading and executing tasks from queue
	err := tq.Start()
	if err != nil {
		log.Fatal(err)
	}

	// Enqueue new task for execution.
	// TaskQ will run Do method of Task when it will have an available worker (goroutine)
	_, err = tq.Enqueue(context.Background(), Task{})
	if err != nil {
		log.Fatal(err)
	}

	// Gracefully shutting down
	err = tq.Close()
	if err != nil {
		log.Fatal(err)
	}
}
