package challenge

import (
	"testing"
	"time"
)

// Benchmark script value parsing
func BenchmarkScriptValue_GetCommand(b *testing.B) {
	testCases := []struct {
		name  string
		value ScriptValue
	}{
		{
			name:  "StringCommand",
			value: ScriptValue{Simple: "docker compose up -d"},
		},
		{
			name: "MapCommand",
			value: ScriptValue{
				Complex: &ScriptConfig{
					Execute: "docker compose up -d",
				},
			},
		},
		{
			name: "MapWithInterval",
			value: ScriptValue{
				Complex: &ScriptConfig{
					Execute:  "docker compose up -d",
					Interval: 5 * time.Minute,
				},
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = tc.value.GetCommand()
			}
		})
	}
}

// Benchmark interval parsing
func BenchmarkScriptValue_GetInterval(b *testing.B) {
	value := ScriptValue{
		Complex: &ScriptConfig{
			Execute:  "docker compose up -d",
			Interval: 5 * time.Minute,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = value.GetInterval()
	}
}

// Benchmark interval validation
func BenchmarkValidateInterval(b *testing.B) {
	intervals := []struct {
		name     string
		duration time.Duration
	}{
		{"Valid1m", time.Minute},
		{"Valid5m", 5 * time.Minute},
		{"Valid1h", time.Hour},
		{"Valid30s", 30 * time.Second},
	}

	for _, iv := range intervals {
		b.Run(iv.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ValidateInterval(iv.duration, "test-script")
			}
		})
	}
}
