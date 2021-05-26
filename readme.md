# TaskQ
[![Go Report](https://goreportcard.com/badge/github.com/antonmashko/taskq)](https://goreportcard.com/report/github.com/antonmashko/taskq)
[![GoDoc](http://godoc.org/github.com/antonmashko/taskq?status.svg)](http://godoc.org/github.com/antonmashko/taskq)
[![Build Status](https://travis-ci.org/antonmashko/taskq.svg)](https://travis-ci.org/antonmashko/taskq)
[![Codecov](https://img.shields.io/codecov/c/github/antonmashko/taskq.svg)](https://codecov.io/gh/antonmashko/taskq)

Goroutine manager. 

---
## Installing 
```bash
go get github.com/antonmashko/taskq
```

## TaskQ
### Initializing
Use TaskQ for managing you goroutines. 
```golang
taskq := New(<size>)
```
size - parameter will control goroutines count. In case if all goroutines are busy, task will be added to queue and will wait for a free worker (goroutine) from pool.

### Add task to TaskQ
```golang
taskID := taskq.Enqueue(<task>)
```

### Start TaskQ
```golang
err := taskq.Start()
```
run all added and future tasks in taskq.
