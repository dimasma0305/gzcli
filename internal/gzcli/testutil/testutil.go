package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// NetworkFailureServer creates a test server that simulates network failures
func NetworkFailureServer(t *testing.T, failureType string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch failureType {
		case "timeout":
			// Simulate timeout by hanging forever
			time.Sleep(time.Hour)
		case "connection_reset":
			// Close connection immediately
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				_ = conn.Close()
			}
		case "slow_response":
			// Send response very slowly
			w.WriteHeader(http.StatusOK)
			for i := 0; i < 100; i++ {
				_, _ = w.Write([]byte("a"))
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				time.Sleep(100 * time.Millisecond)
			}
		case "malformed_json":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"invalid": json}`))
		case "incomplete_response":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data": "incomplete`))
		case "rate_limit":
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error": "rate limit exceeded"}`))
		case "internal_error":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "internal server error"}`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success": true}`))
		}
	}))
}

// TimeoutDialer creates a dialer that times out immediately
func TimeoutDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   1 * time.Nanosecond,
		KeepAlive: 1 * time.Nanosecond,
	}
}

// CreateTimeoutClient creates an HTTP client with very short timeout
func CreateTimeoutClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: timeout,
			}).DialContext,
		},
	}
}

// ConcurrentTest runs a test function concurrently and reports any panics or errors
func ConcurrentTest(t *testing.T, concurrency int, iterations int, testFunc func(id int, iteration int) error) {
	var wg sync.WaitGroup
	errChan := make(chan error, concurrency*iterations)
	panicChan := make(chan interface{}, concurrency*iterations)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicChan <- fmt.Sprintf("Worker %d panicked: %v", workerID, r)
				}
			}()

			for iter := 0; iter < iterations; iter++ {
				if err := testFunc(workerID, iter); err != nil {
					errChan <- fmt.Errorf("worker %d, iteration %d: %w", workerID, iter, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errChan)
	close(panicChan)

	// Check for panics
	var panics []string
	for p := range panicChan {
		panics = append(panics, fmt.Sprint(p))
	}
	if len(panics) > 0 {
		t.Errorf("Panics occurred during concurrent test:\n%s", strings.Join(panics, "\n"))
	}

	// Check for errors
	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}
	if len(errors) > 0 {
		t.Errorf("Errors occurred during concurrent test:\n%s", strings.Join(errors, "\n"))
	}
}

// GenerateLargeJSON creates a large JSON payload for testing
func GenerateLargeJSON(sizeMB int) []byte {
	// Create array of objects to reach target size
	targetSize := sizeMB * 1024 * 1024
	var data []map[string]string

	itemSize := 100 // approximate size per item
	itemCount := targetSize / itemSize

	for i := 0; i < itemCount; i++ {
		data = append(data, map[string]string{
			"id":    fmt.Sprintf("item_%d", i),
			"data":  RandomString(50),
			"value": fmt.Sprintf("value_%d", i),
		})
	}

	result, _ := json.Marshal(data)
	return result
}

// GenerateLargeString creates a large string for testing
func GenerateLargeString(sizeKB int) string {
	size := sizeKB * 1024
	b := make([]byte, size)
	for i := range b {
		b[i] = byte('a' + (i % 26))
	}
	return string(b)
}

// RandomString generates a random string of specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// CreateMalformedYAML generates various types of malformed YAML
func CreateMalformedYAML(malformType string) string {
	switch malformType {
	case "invalid_syntax":
		return "key: value\n  invalid indentation\nkey2: value2"
	case "unclosed_quote":
		return `key: "unclosed value`
	case "invalid_utf8":
		return "key: \xff\xfe value"
	case "extremely_nested":
		nested := "key: "
		for i := 0; i < 1000; i++ {
			nested += "\n  " + strings.Repeat(" ", i) + "level: " + fmt.Sprint(i)
		}
		return nested
	case "duplicate_keys":
		return "key: value1\nkey: value2\nkey: value3"
	case "null_bytes":
		return "key: value\x00more"
	default:
		return "invalid: yaml: syntax:"
	}
}

// CreateMalformedCSV generates various types of malformed CSV
func CreateMalformedCSV(malformType string) string {
	switch malformType {
	case "missing_headers":
		return "John,john@test.com,Team1\nJane,jane@test.com,Team2"
	case "inconsistent_columns":
		return "Name,Email,Team\nJohn,john@test.com\nJane,jane@test.com,Team2,Extra"
	case "unclosed_quotes":
		return `Name,Email,Team
John,"john@test.com,Team1`
	case "bom":
		return "\xef\xbb\xbfName,Email,Team\nJohn,john@test.com,Team1"
	case "null_bytes":
		return "Name,Email,Team\nJohn\x00,john@test.com,Team1"
	case "sql_injection":
		return `Name,Email,Team
'; DROP TABLE users; --,john@test.com,Team1`
	case "xss_attempt":
		return `Name,Email,Team
<script>alert('xss')</script>,john@test.com,Team1`
	default:
		return "malformed,csv"
	}
}

// MockServerWithDelay creates a server that responds after a delay
func MockServerWithDelay(t *testing.T, delay time.Duration, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		handler(w, r)
	}))
}

// MockServerWithAuth creates a server that requires authentication
func MockServerWithAuth(t *testing.T, validToken string, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		cookie, _ := r.Cookie("session")

		if auth != "Bearer "+validToken && (cookie == nil || cookie.Value != validToken) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "unauthorized"}`))
			return
		}
		handler(w, r)
	}))
}

// RaceConditionDetector helps detect race conditions in tests
type RaceConditionDetector struct {
	mu       sync.Mutex
	counters map[string]int
	errors   []string
}

func NewRaceConditionDetector() *RaceConditionDetector {
	return &RaceConditionDetector{
		counters: make(map[string]int),
	}
}

func (r *RaceConditionDetector) Increment(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[key]++
}

func (r *RaceConditionDetector) AddError(err string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors = append(r.errors, err)
}

func (r *RaceConditionDetector) GetCount(key string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.counters[key]
}

func (r *RaceConditionDetector) GetErrors() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string{}, r.errors...)
}

// SimulateDiskFull creates a writer that fails after certain amount of data
type DiskFullWriter struct {
	written int
	limit   int
}

func NewDiskFullWriter(limit int) *DiskFullWriter {
	return &DiskFullWriter{limit: limit}
}

func (d *DiskFullWriter) Write(p []byte) (n int, err error) {
	if d.written+len(p) > d.limit {
		return 0, fmt.Errorf("no space left on device")
	}
	d.written += len(p)
	return len(p), nil
}

// FailingReader simulates read failures
type FailingReader struct {
	failAfter int
	read      int
}

func NewFailingReader(failAfter int) *FailingReader {
	return &FailingReader{failAfter: failAfter}
}

func (f *FailingReader) Read(p []byte) (n int, err error) {
	if f.read >= f.failAfter {
		return 0, fmt.Errorf("read error: connection reset")
	}
	n = len(p)
	if f.read+n > f.failAfter {
		n = f.failAfter - f.read
	}
	f.read += n
	return n, nil
}

// CreateChaosServer creates a server that randomly fails requests
func CreateChaosServer(t *testing.T, failureRate float64, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rand.Float64() < failureRate {
			// Random failure
			failures := []int{
				http.StatusInternalServerError,
				http.StatusBadGateway,
				http.StatusServiceUnavailable,
				http.StatusGatewayTimeout,
			}
			w.WriteHeader(failures[rand.Intn(len(failures))])
			_, _ = w.Write([]byte(`{"error": "random failure"}`))
			return
		}
		handler(w, r)
	}))
}

// WaitWithTimeout waits for a condition with timeout
func WaitWithTimeout(t *testing.T, timeout time.Duration, condition func() bool, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for condition: %s", message)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// AssertNoError fails the test if error is not nil
func AssertNoError(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

// AssertError fails the test if error is nil
func AssertError(t *testing.T, err error, message string) {
	if err == nil {
		t.Fatalf("%s: expected error but got nil", message)
	}
}

// AssertContains fails if the string doesn't contain the substring
func AssertContains(t *testing.T, str, substr string) {
	if !strings.Contains(str, substr) {
		t.Fatalf("Expected string to contain %q, got: %s", substr, str)
	}
}

// CaptureStderr captures stderr output during test
func CaptureStderr(t *testing.T, fn func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	fn()

	_ = w.Close()
	os.Stderr = old

	var buf strings.Builder
	_, _ = io.Copy(&buf, r)
	return buf.String()
}
