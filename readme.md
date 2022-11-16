# TaskQ
[![Go Report](https://goreportcard.com/badge/github.com/antonmashko/taskq)](https://goreportcard.com/report/github.com/antonmashko/taskq)
[![GoDoc](http://godoc.org/github.com/antonmashko/taskq?status.svg)](http://godoc.org/github.com/antonmashko/taskq)
[![Build Status](https://travis-ci.org/antonmashko/taskq.svg)](https://travis-ci.org/antonmashko/taskq)
[![Codecov](https://img.shields.io/codecov/c/github/antonmashko/taskq.svg)](https://codecov.io/gh/antonmashko/taskq)

Simple and powerful goroutine manager.

* [Installing](#installing)
* [Purpose](#purpose)
* [Example](#example)
* [Persistence and Queues](#persistence-and-queues)
* [Task Events](#task-events)
* [Benchmark results](#benchmark-results)

---
## Installing 
```bash
go get github.com/antonmashko/taskq
```

## Purpose
Usually, you need `TaskQ` for managing your goroutines. This will allow you to control service resource consumption and graceful shutdown.

## Example

```golang
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

```
More examples [here](example)

## Persistence and Queues
By default TaskQ stores all tasks in memory using [ConcurrencyQueue](https://pkg.go.dev/github.com/antonmashko/taskq#ConcurrentQueue). For creating custom queue you need to implement interface [Queue](https://pkg.go.dev/github.com/antonmashko/taskq#Queue) and pass it as argument on creating [NewWithQueue](https://pkg.go.dev/github.com/antonmashko/taskq#NewWithQueue) or [PoolWithQueue](https://pkg.go.dev/github.com/antonmashko/taskq#PoolWithQueue).
See [example](example/redis-custom-queue) of how to adapt redis queue into TaskQ

## Task Events
Task support two type of events:
1. Done - completion of the task. https://pkg.go.dev/github.com/antonmashko/taskq#TaskDone
2. OnError - error handling event. https://pkg.go.dev/github.com/antonmashko/taskq#TaskOnError 
For invoking event implement interface on your task ([example](example/task-events)).

## Graceful shutdown
[Shutdown](https://pkg.go.dev/github.com/antonmashko/taskq#TaskQ.Shutdown) and [Close](https://pkg.go.dev/github.com/antonmashko/taskq#TaskQ.Close) gracefully shuts down the TaskQ without interrupting any active tasks. If TaskQ need to finish all tasks in queue, use context [ContextWithWait](https://pkg.go.dev/github.com/antonmashko/taskq#ContextWithWait) as `Shutdown` method argument.

## Benchmark results
[Benchmarks](benchmarks/readme.md)
