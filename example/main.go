package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

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
		log.Print("added task with id:", tq.Enqueue(&printer{}))
	}
	tq.Close()
	time.Sleep(time.Second)
}
