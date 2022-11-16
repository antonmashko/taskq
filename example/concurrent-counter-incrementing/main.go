package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/antonmashko/taskq"
)

var counter = int32(0)

type printer struct{}

func (p *printer) Do(_ context.Context) error {
	fmt.Println("global counter:", atomic.AddInt32(&counter, 1))
	return nil
}

func main() {
	tq := taskq.New(0)
	tq.Start()
	for i := 0; i < 1000; i++ {
		id, err := tq.Enqueue(context.Background(), &printer{})
		log.Printf("added task with id: %d; err=%v\n", id, err)
	}
	tq.Shutdown(taskq.ContextWithWait(context.Background()))
	log.Println("Result:", counter)
}
