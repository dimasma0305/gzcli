//nolint:revive // Package name kept as "log" for stable internal imports.
package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var debugMode = false

// SetDebugMode enables or disables debug logging
func SetDebugMode(enabled bool) {
	debugMode = enabled
}

// Debug logs debug messages when debug mode is enabled
func Debug(format string, elem ...any) {
	if debugMode {
		fmt.Println(color.CyanString("[DEBUG] ") + fmt.Sprintf(format, elem...))
	}
}

// DebugH2 logs indented debug messages when debug mode is enabled
func DebugH2(format string, elem ...any) {
	if debugMode {
		fmt.Println(color.CyanString("  [DEBUG] ") + fmt.Sprintf(format, elem...))
	}
}

// DebugH3 logs more indented debug messages when debug mode is enabled
func DebugH3(format string, elem ...any) {
	if debugMode {
		fmt.Println(color.CyanString("    [DEBUG] ") + fmt.Sprintf(format, elem...))
	}
}

// Fatal logs an error message and exits the program
func Fatal(args ...interface{}) {
	var message string

	// Handle different argument combinations
	switch len(args) {
	case 0:
		message = "fatal error occurred"
	case 1:
		switch v := args[0].(type) {
		case error:
			message = v.Error()
		case string:
			message = v
		default:
			message = fmt.Sprintf("%v", v)
		}
	default:
		// If first argument is a string, use as format
		if format, ok := args[0].(string); ok {
			message = fmt.Sprintf(format, args[1:]...)
		} else {
			// Otherwise, just print all arguments
			message = fmt.Sprint(args...)
		}
	}

	// Format and print the error message
	lines := strings.Split(strings.TrimSpace(message), "\n")
	for _, line := range lines {
		fmt.Fprintln(os.Stderr, color.RedString("[x] ")+line)
	}
	os.Exit(1)
}

// Error logs an error message to stderr
func Error(str string, elem ...any) {
	fmt.Fprintln(os.Stderr, color.RedString("[x] ")+fmt.Sprintf(str, elem...))
}

// ErrorH2 logs an indented error message to stderr
func ErrorH2(format string, elem ...any) {
	fmt.Fprintln(os.Stderr, color.RedString("  [x] ")+fmt.Sprintf(format, elem...))
}

// Info logs an informational message
func Info(format string, elem ...any) {
	fmt.Println(color.BlueString("[x] ") + fmt.Sprintf(format, elem...))
}

// InfoH2 logs an indented informational message
func InfoH2(format string, elem ...any) {
	fmt.Println(color.GreenString("  [x] ") + fmt.Sprintf(format, elem...))
}

// InfoH3 logs a double-indented informational message
func InfoH3(format string, elem ...any) {
	fmt.Println(color.YellowString("    [x] ") + fmt.Sprintf(format, elem...))
}

// SuccessDownload logs a successful challenge download
func SuccessDownload(challName string, challCategory string) {
	Info("success downloading: %s (%s)", challName, challCategory)
}
