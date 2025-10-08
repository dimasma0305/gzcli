package filesystem

import (
	"path/filepath"
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/watchertypes"
	"github.com/fsnotify/fsnotify"
)

// Benchmark file path matching
func BenchmarkShouldProcessEvent(b *testing.B) {
	event := fsnotify.Event{
		Name: "/home/user/ctf/web/challenge1/src/main.go",
		Op:   fsnotify.Write,
	}
	config := watchertypes.WatcherConfig{
		IgnorePatterns: []string{"*.tmp", "*.log", "*.swp", ".git/*"},
		WatchPatterns:  []string{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ShouldProcessEvent(event, config)
	}
}

// Benchmark update type determination
func BenchmarkDetermineUpdateType(b *testing.B) {
	filePath := "/home/user/ctf/web/challenge1/src/main.go"
	challengeCwd := "/home/user/ctf/web/challenge1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetermineUpdateType(filePath, challengeCwd)
	}
}

// Benchmark challenge category detection
func BenchmarkIsInChallengeCategory(b *testing.B) {
	filePath := "/home/user/ctf/Pwn/buffer-overflow/src/exploit.c"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsInChallengeCategory(filePath)
	}
}

// Benchmark ignore pattern matching with many patterns
func BenchmarkShouldProcessEvent_ManyPatterns(b *testing.B) {
	event := fsnotify.Event{
		Name: "/home/user/ctf/web/challenge1/src/main.go",
		Op:   fsnotify.Write,
	}
	config := watchertypes.WatcherConfig{
		IgnorePatterns: []string{
			"*.tmp", "*.log", "*.swp", ".git/*", "node_modules/*",
			"*.pyc", "__pycache__/*", ".vscode/*", ".idea/*",
			"dist/*", "build/*", "*.cache", "*.bak",
		},
		WatchPatterns: []string{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ShouldProcessEvent(event, config)
	}
}

// Benchmark path operations
func BenchmarkPathOperations(b *testing.B) {
	filePath := "/home/user/ctf/web/challenge1/src/main.go"
	challengeCwd := "/home/user/ctf/web/challenge1"

	b.Run("Abs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = filepath.Abs(filePath)
		}
	})

	b.Run("Rel", func(b *testing.B) {
		absFile, _ := filepath.Abs(filePath)
		absChallenge, _ := filepath.Abs(challengeCwd)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = filepath.Rel(absChallenge, absFile)
		}
	})

	b.Run("Match", func(b *testing.B) {
		pattern := "*.go"
		name := filepath.Base(filePath)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = filepath.Match(pattern, name)
		}
	})
}
