package daemon

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"os/exec"

	"github.com/dimasma0305/gzcli/internal/log"
	tail "github.com/hpcloud/tail"
)

// ShowRecentLogs displays recent log entries if the log file exists
func ShowRecentLogs(logFile string) {
	if _, err := os.Stat(logFile); err != nil {
		return // Log file doesn't exist
	}

	log.Info("")
	log.Info("📋 Recent Activity (last 5 lines from log):")

	// Use tail command to get last few lines
	cmd := exec.Command("tail", "-n", "5", logFile)
	output, err := cmd.Output()
	if err != nil {
		log.Info("   (Unable to read log file)")
		return
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			log.Info("   %s", strings.TrimSpace(line))
		}
	}
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
