package benchmarks

import (
	"context"
	"encoding/json"
	"testing"
)

func SmallJSONUnmarshalF(_ context.Context) {
	var jsonB = []byte(`{"key1":"val1","key2":"val2"}`)
	result := make(map[string]interface{})
	_ = json.Unmarshal(jsonB, &result)
}

func BenchmarkTaskq_SmallJSONUnmarshal(b *testing.B) {
	TaskqTestF(b, SmallJSONUnmarshalF)
}

func BenchmarkSpawningGoroutines_SmallJSONUnmarshal(b *testing.B) {
	SpawningGoroutinesTestF(b, SmallJSONUnmarshalF)
}
