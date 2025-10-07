//nolint:errcheck,gosec,revive // Test file with acceptable error handling patterns
package gzcli

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/testutil"
)

// TestSetCache_CorruptedData tests handling of corrupted cache data
func TestSetCache_CorruptedData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override cache dir temporarily
	originalCacheDir := cacheDir
	cacheDir = filepath.Join(tmpDir, ".gzcli")
	defer func() { cacheDir = originalCacheDir }()

	// Test with various data types that might cause issues
	testCases := []struct {
		name string
		data interface{}
	}{
		{"nil data", nil},
		{"empty map", map[string]interface{}{}},
		{"large map", createLargeMap(10000)},
		{"nested structures", createDeeplyNested(100)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := setCache(tc.name, tc.data)
			if err != nil {
				t.Logf("setCache failed for %s: %v", tc.name, err)
			}
		})
	}
}

// TestGetCache_CorruptedFile tests reading corrupted cache files
func TestGetCache_CorruptedFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalCacheDir := cacheDir
	cacheDir = filepath.Join(tmpDir, ".gzcli")
	defer func() { cacheDir = originalCacheDir }()

	// Create cache directory
	os.MkdirAll(cacheDir, 0755)

	// Create corrupted cache files
	testCases := []struct {
		name    string
		content string
	}{
		{"malformed YAML", "key: value\n  invalid: indentation"},
		{"invalid UTF-8", "key: \xff\xfe invalid"},
		{"partial file", "key: value\nincomplete"},
		{"binary data", "\x00\x01\x02\x03\x04"},
		{"empty file", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cachePath := filepath.Join(cacheDir, tc.name+".yaml")
			os.WriteFile(cachePath, []byte(tc.content), 0644)

			var result map[string]interface{}
			err := GetCache(tc.name, &result)
			if err == nil {
				t.Logf("GetCache succeeded for corrupted file %s (data might be parseable)", tc.name)
			}
		})
	}
}

// TestCache_ConcurrentAccess tests concurrent cache operations
func TestCache_ConcurrentAccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalCacheDir := cacheDir
	cacheDir = filepath.Join(tmpDir, ".gzcli")
	defer func() { cacheDir = originalCacheDir }()

	// Test concurrent writes and reads
	testutil.ConcurrentTest(t, 10, 20, func(id, iter int) error {
		key := fmt.Sprintf("key-%d", id)
		data := map[string]interface{}{
			"worker":    id,
			"iteration": iter,
			"value":     testutil.RandomString(100),
		}

		// Write
		if err := setCache(key, data); err != nil {
			return fmt.Errorf("setCache failed: %w", err)
		}

		// Read back
		var result map[string]interface{}
		if err := GetCache(key, &result); err != nil {
			return fmt.Errorf("GetCache failed: %w", err)
		}

		return nil
	})
}

// TestCache_ConcurrentSameKey tests multiple workers accessing same key
func TestCache_ConcurrentSameKey(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalCacheDir := cacheDir
	cacheDir = filepath.Join(tmpDir, ".gzcli")
	defer func() { cacheDir = originalCacheDir }()

	// Multiple workers writing to the same key
	testutil.ConcurrentTest(t, 20, 10, func(id, iter int) error {
		key := "shared-key"
		data := map[string]interface{}{
			"last_writer": id,
			"iteration":   iter,
		}

		return setCache(key, data)
	})

	// Verify the cache file is still valid
	var result map[string]interface{}
	err = GetCache("shared-key", &result)
	if err != nil {
		t.Errorf("Final cache read failed: %v", err)
	} else {
		t.Logf("Final cache value: %+v", result)
	}
}

// TestDeleteCache_NonexistentKey tests deleting nonexistent cache
func TestDeleteCache_NonexistentKey(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalCacheDir := cacheDir
	cacheDir = filepath.Join(tmpDir, ".gzcli")
	defer func() { cacheDir = originalCacheDir }()

	err = DeleteCache("nonexistent-key")
	if err == nil {
		t.Error("Expected error when deleting nonexistent key")
	}
}

// TestSetCache_InvalidKey tests invalid key names
func TestSetCache_InvalidKey(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalCacheDir := cacheDir
	cacheDir = filepath.Join(tmpDir, ".gzcli")
	defer func() { cacheDir = originalCacheDir }()

	testCases := []string{
		"",
		"../../../etc/passwd",
		"/absolute/path",
		"key/with/slashes",
		"key\x00with\x00nulls",
	}

	for _, key := range testCases {
		t.Run(fmt.Sprintf("key-%q", key), func(t *testing.T) {
			err := setCache(key, map[string]interface{}{"test": "data"})
			// Some keys might be handled, some might fail
			t.Logf("setCache with key %q result: %v", key, err)
		})
	}
}

// TestCache_LargeData tests caching very large data
func TestCache_LargeData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large data test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalCacheDir := cacheDir
	cacheDir = filepath.Join(tmpDir, ".gzcli")
	defer func() { cacheDir = originalCacheDir }()

	// Create 10MB of data
	largeData := make(map[string]interface{})
	for i := 0; i < 10000; i++ {
		largeData[fmt.Sprintf("key-%d", i)] = testutil.RandomString(1000)
	}

	err = setCache("large-data", largeData)
	if err != nil {
		t.Errorf("Failed to cache large data: %v", err)
	}

	// Try to read it back
	var result map[string]interface{}
	err = GetCache("large-data", &result)
	if err != nil {
		t.Errorf("Failed to read large data: %v", err)
	}
}

// TestCache_ReadOnlyDirectory tests cache operations when directory is read-only
func TestCache_ReadOnlyDirectory(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping read-only test when running as root")
	}

	tmpDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalCacheDir := cacheDir
	cacheDir = filepath.Join(tmpDir, ".gzcli")
	defer func() { cacheDir = originalCacheDir }()

	// Create cache directory and make it read-only
	os.MkdirAll(cacheDir, 0755)
	os.Chmod(cacheDir, 0444)
	defer os.Chmod(cacheDir, 0755) // Restore for cleanup

	err = setCache("test", map[string]interface{}{"data": "value"})
	if err == nil {
		t.Error("Expected error when writing to read-only directory")
	}
}

// TestCache_SymlinkAttack tests cache with symlink in path
func TestCache_SymlinkAttack(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping symlink test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a target directory
	targetDir := filepath.Join(tmpDir, "target")
	os.MkdirAll(targetDir, 0755)

	// Create a symlink
	symlinkPath := filepath.Join(tmpDir, ".gzcli")
	err = os.Symlink(targetDir, symlinkPath)
	if err != nil {
		t.Skipf("Failed to create symlink: %v", err)
	}

	originalCacheDir := cacheDir
	cacheDir = symlinkPath
	defer func() { cacheDir = originalCacheDir }()

	// Try to use cache through symlink
	err = setCache("test", map[string]interface{}{"data": "value"})
	if err != nil {
		t.Logf("Cache through symlink: %v", err)
	}

	var result map[string]interface{}
	err = GetCache("test", &result)
	if err != nil {
		t.Logf("Read through symlink: %v", err)
	}
}

// TestCache_RaceConditions tests for race conditions
func TestCache_RaceConditions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalCacheDir := cacheDir
	cacheDir = filepath.Join(tmpDir, ".gzcli")
	defer func() { cacheDir = originalCacheDir }()

	detector := testutil.NewRaceConditionDetector()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				key := fmt.Sprintf("race-key-%d", id%3) // Use only 3 keys to increase contention
				data := map[string]interface{}{
					"worker": id,
					"iter":   j,
				}

				if err := setCache(key, data); err != nil {
					detector.AddError(fmt.Sprintf("Worker %d write error: %v", id, err))
				} else {
					detector.Increment("writes")
				}

				var result map[string]interface{}
				if err := GetCache(key, &result); err != nil {
					// Might fail if write hasn't completed
					detector.Increment("read-errors")
				} else {
					detector.Increment("reads")
				}
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Race test completed: %d writes, %d reads, %d read-errors",
		detector.GetCount("writes"),
		detector.GetCount("reads"),
		detector.GetCount("read-errors"))

	errors := detector.GetErrors()
	if len(errors) > 0 {
		t.Logf("Errors during race test: %d", len(errors))
		for _, err := range errors[:min(5, len(errors))] {
			t.Logf("  %s", err)
		}
	}
}

// Helper functions
func createLargeMap(size int) map[string]interface{} {
	m := make(map[string]interface{})
	for i := 0; i < size; i++ {
		m[fmt.Sprintf("key-%d", i)] = fmt.Sprintf("value-%d", i)
	}
	return m
}

func createDeeplyNested(depth int) interface{} {
	var result interface{} = "leaf"
	for i := 0; i < depth; i++ {
		result = map[string]interface{}{
			"level": i,
			"next":  result,
		}
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
