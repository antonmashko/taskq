package benchmarks

import (
	"context"
	"testing"
	"time"
)

func BenchmarkTaskq_SleepF(b *testing.B) {
	sleeps := []time.Duration{
		time.Microsecond,
		50 * time.Microsecond,
		time.Millisecond,
		50 * time.Millisecond,
	}

	for _, s := range sleeps {
		b.Run(s.String(), func(b *testing.B) {
			TaskqTestF(b, func(ctx context.Context) {
				time.Sleep(s)
			})
		})
	}
}

func BenchmarkSpawningGoroutines_SleepF(b *testing.B) {
	sleeps := []time.Duration{
		time.Microsecond,
		50 * time.Microsecond,
		time.Millisecond,
		50 * time.Millisecond,
	}

	for _, s := range sleeps {
		b.Run(s.String(), func(b *testing.B) {
			SpawningGoroutinesTestF(b, func(ctx context.Context) {
				time.Sleep(s)
			})
		})
	}
}
