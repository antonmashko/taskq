package benchmarks

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func SleepAndSmallJSONUnmarshalF(_ context.Context) {
	// using time.Sleep with 50ms for simulating http request and response executing time
	time.Sleep(50 * time.Millisecond)
	var jsonB = []byte(`{"key1":"val1","key2":"val2"}`)
	result := make(map[string]interface{})
	_ = json.Unmarshal(jsonB, &result)
}

func BenchmarkTaskq_SleepAndSmallJSONUnmarshalF(b *testing.B) {
	TaskqTestF(b, SleepAndSmallJSONUnmarshalF)
}

func BenchmarkSpawningGoroutines_SleepAndSmallJSONUnmarshalF(b *testing.B) {
	SpawningGoroutinesTestF(b, SleepAndSmallJSONUnmarshalF)
}

func BenchmarkTaskqPool_SleepAndSmallJSONUnmarshalF(b *testing.B) {
	TaskqPoolTestF(b, SleepAndSmallJSONUnmarshalF)
}
