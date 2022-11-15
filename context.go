package taskq

import "context"

type ctxWaitKey struct{}

func ContextWithWait(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxWaitKey{}, true)
}
