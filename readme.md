# TaskQ
[![Go Report](https://goreportcard.com/badge/github.com/antonmashko/taskq)](https://goreportcard.com/report/github.com/antonmashko/taskq)
[![GoDoc](http://godoc.org/github.com/antonmashko/taskq?status.svg)](http://godoc.org/github.com/antonmashko/taskq)
[![Build Status](https://travis-ci.org/antonmashko/taskq.svg)](https://travis-ci.org/antonmashko/taskq)
[![Codecov](https://img.shields.io/codecov/c/github/antonmashko/taskq.svg)](https://codecov.io/gh/antonmashko/taskq)

Simple and powerful goroutine manager.

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

## Graceful shutdown
[Shutdown](https://pkg.go.dev/github.com/antonmashko/taskq#TaskQ.Shutdown) and [Close](https://pkg.go.dev/github.com/antonmashko/taskq#TaskQ.Close) gracefully shuts down the TaskQ without interrupting any active tasks. If TaskQ need to finish all tasks in queue, use context [ContextWithWait](https://pkg.go.dev/github.com/antonmashko/taskq#ContextWithWait) as `Shutdown` method argument.

## Benchmark results
[Benchmarks](benchmarks/readme.md)
