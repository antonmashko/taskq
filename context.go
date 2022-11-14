package taskq

import "context"

type ctxWaitKey struct{}

var ContextWithWait = context.WithValue(context.Background(), ctxWaitKey{}, true)
