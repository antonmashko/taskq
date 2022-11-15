# TaskQ
[![Go Report](https://goreportcard.com/badge/github.com/antonmashko/taskq)](https://goreportcard.com/report/github.com/antonmashko/taskq)
[![GoDoc](http://godoc.org/github.com/antonmashko/taskq?status.svg)](http://godoc.org/github.com/antonmashko/taskq)
[![Build Status](https://travis-ci.org/antonmashko/taskq.svg)](https://travis-ci.org/antonmashko/taskq)
[![Codecov](https://img.shields.io/codecov/c/github/antonmashko/taskq.svg)](https://codecov.io/gh/antonmashko/taskq)

Simple and powerful goroutines manager.

---
## Installing 
```bash
go get github.com/antonmashko/taskq
```

## TaskQ when and how to use
Usually you need `TaskQ` for controlling resources of service (goroutines spawning).
Example:
```golang
taskq := New(10)
taskq.Start()
taskID, err := taskq.Enqueue(context.Background(), taskq.TaskFunc(func(ctx context.Context) error {
        fmt.Println("hello world")
		return nil
	}))
```
In this example, we created TaskQ with limiting max active goroutines count to 10. Tasks for executing we're adding to TaskQ with `Enqueue`. Task will be executed when TaskQ will be able to create new worker (goroutine). Goroutine will dispose and next task will be executed in a new goroutine.

Use `limit=0` for not limiting goroutines number.


## Graceful shutdown
`Shutdown` and `Close` gracefully shuts down the TaskQ without interrupting any active tasks. If TaskQ need to finish all tasks in queue, use context `ContextWithWait` with `Shutdown` method.
