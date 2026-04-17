package daemon

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dimasma0305/gzcli/internal/log"
	tail "github.com/hpcloud/tail"
)

// maxRecentLogLines is the number of trailing lines to display from a log file.
const maxRecentLogLines = 5

// ShowRecentLogs displays recent log entries if the log file exists.
// The last maxRecentLogLines lines are read in-process to avoid spawning a
// subprocess with a user-controlled path (gosec G204).
func ShowRecentLogs(logFile string) {
	if _, err := os.Stat(logFile); err != nil {
		return // Log file doesn't exist
	}

	log.Info("")
	log.Info("📋 Recent Activity (last %d lines from log):", maxRecentLogLines)

	lines, err := readLastLines(logFile, maxRecentLogLines)
	if err != nil {
		log.Info("   (Unable to read log file)")
		return
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			log.Info("   %s", trimmed)
		}
	}
}

// readLastLines returns up to n trailing lines from path. Lines are returned
// in file order (oldest first).
func readLastLines(path string, n int) ([]string, error) {
	//nolint:gosec // G304: logFile path is provided by the operator via CLI flag.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	ring := make([]string, 0, n)
	scanner := bufio.NewScanner(f)
	// Allow long log lines (default buffer is 64KiB).
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		if len(ring) == n {
			ring = ring[1:]
		}
		ring = append(ring, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return ring, nil
}

// FollowLogs follows a log file and displays new content in real-time
func FollowLogs(logFile string) error {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Tail the log file with re-open and follow options to handle rotations
	t, err := tail.TailFile(logFile, tail.Config{
		ReOpen:    true,
		Follow:    true,
		MustExist: false,
		Poll:      true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: io.SeekEnd},
	})
	if err != nil {
		return fmt.Errorf("failed to tail log file: %w", err)
	}
	defer t.Cleanup()

	// Print a header and the last few lines for context
	ShowRecentLogs(logFile)
	fmt.Println()

	for {
		select {
		case <-sigChan:
			fmt.Println("\n📋 Log following stopped.")
			return nil
		case line, ok := <-t.Lines:
			if !ok {
				return fmt.Errorf("log tail channel closed")
			}
			if line == nil {
				continue
			}
			text := line.Text
			if strings.TrimSpace(text) == "" {
				continue
			}
			if strings.Contains(text, "[x]") || strings.Contains(text, "INFO") || strings.Contains(text, "ERROR") {
				fmt.Println(text)
			} else {
				fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), text)
			}
		}
	}
}
