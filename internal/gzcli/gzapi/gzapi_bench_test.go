package gzapi

import (
	"strings"
	"testing"
)

// Benchmark URL construction with string concatenation vs strings.Builder
func BenchmarkURLConstruction_Concat(b *testing.B) {
	baseURL := "https://ctf.example.com"
	path := "/api/challenges/list"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = baseURL + path
	}
}

func BenchmarkURLConstruction_Builder(b *testing.B) {
	baseURL := "https://ctf.example.com"
	path := "/api/challenges/list"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var builder strings.Builder
		builder.Grow(len(baseURL) + len(path))
		builder.WriteString(baseURL)
		builder.WriteString(path)
		_ = builder.String()
	}
}

// Benchmark for multiple URL constructions (simulating batch requests)
func BenchmarkURLConstruction_Batch_Concat(b *testing.B) {
	baseURL := "https://ctf.example.com"
	paths := []string{
		"/api/challenges/list",
		"/api/challenges/1",
		"/api/challenges/2",
		"/api/teams/list",
		"/api/games/current",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			_ = baseURL + path
		}
	}
}

func BenchmarkURLConstruction_Batch_Builder(b *testing.B) {
	baseURL := "https://ctf.example.com"
	paths := []string{
		"/api/challenges/list",
		"/api/challenges/1",
		"/api/challenges/2",
		"/api/teams/list",
		"/api/games/current",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var builder strings.Builder
		for _, path := range paths {
			builder.Reset()
			builder.Grow(len(baseURL) + len(path))
			builder.WriteString(baseURL)
			builder.WriteString(path)
			_ = builder.String()
		}
	}
}

