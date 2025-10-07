package gzcli

import (
	"os"
	"path/filepath"
	"testing"
)

type benchmarkData struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Tags        []string          `yaml:"tags"`
	Metadata    map[string]string `yaml:"metadata"`
}

// BenchmarkSetCache measures cache write performance
func BenchmarkSetCache(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "cache-bench-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	originalCacheDir := cacheDir
	cacheDir = filepath.Join(tmpDir, ".gzcli")
	defer func() { cacheDir = originalCacheDir }()

	data := benchmarkData{
		Name:        "Test Challenge",
		Description: "This is a test challenge for benchmarking",
		Tags:        []string{"web", "crypto", "forensics"},
		Metadata: map[string]string{
			"author":     "test",
			"difficulty": "medium",
			"points":     "500",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = setCache("benchmark", data)
	}
}

// BenchmarkGetCache measures cache read performance
func BenchmarkGetCache(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "cache-bench-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	originalCacheDir := cacheDir
	cacheDir = filepath.Join(tmpDir, ".gzcli")
	defer func() { cacheDir = originalCacheDir }()

	data := benchmarkData{
		Name:        "Test Challenge",
		Description: "This is a test challenge for benchmarking",
		Tags:        []string{"web", "crypto", "forensics"},
		Metadata: map[string]string{
			"author":     "test",
			"difficulty": "medium",
			"points":     "500",
		},
	}

	// Setup: Write cache once
	_ = setCache("benchmark", data)

	var result benchmarkData
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetCache("benchmark", &result)
	}
}

// BenchmarkSetCache_Large measures cache write performance with large data
func BenchmarkSetCache_Large(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "cache-bench-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	originalCacheDir := cacheDir
	cacheDir = filepath.Join(tmpDir, ".gzcli")
	defer func() { cacheDir = originalCacheDir }()

	// Create large dataset
	largeData := make(map[string]benchmarkData)
	for i := 0; i < 100; i++ {
		largeData[string(rune(i))] = benchmarkData{
			Name:        "Challenge " + string(rune(i)),
			Description: "This is a test challenge for benchmarking with id " + string(rune(i)),
			Tags:        []string{"web", "crypto", "forensics", "pwn", "reverse"},
			Metadata: map[string]string{
				"author":     "test",
				"difficulty": "hard",
				"points":     "1000",
				"category":   "misc",
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = setCache("benchmark-large", largeData)
	}
}
