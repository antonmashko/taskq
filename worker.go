package taskq

import (
	"sync/atomic"
)

type workerStatus int32

const (
	registered workerStatus = iota
	idle
	live
	stopped
)

type worker struct {
	id     uint64
	status workerStatus
	stop   func()
}

func (w *worker) isStatus(s workerStatus) bool {
	return workerStatus(atomic.LoadInt32((*int32)(&w.status))) == s
}

func (w *worker) setStatus(s workerStatus) {
	atomic.StoreInt32((*int32)(&w.status), int32(s))
}
