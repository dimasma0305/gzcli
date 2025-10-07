package gzcli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/yaml.v2"
)

// cacheDir caches the working directory to avoid repeated lookups
var cacheDir = func() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, ".gzcli")
}()

// setCache atomically writes data to cache with proper directory creation
func setCache(key string, data any) error {
	cachePath := filepath.Join(cacheDir, key+".yaml")

	// Create cache directory with proper permissions
	if err := os.MkdirAll(filepath.Dir(cachePath), 0750); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Atomic write pattern using temp file
	tmpFile, err := os.CreateTemp(cacheDir, "tmp-")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	// Use buffered writer with pre-allocated buffer
	bw := bufio.NewWriterSize(tmpFile, 32*1024) // 32KB buffer
	if err := yaml.NewEncoder(bw).Encode(data); err != nil {
		return fmt.Errorf("encoding failed: %w", err)
	}

	// Flush buffer before renaming
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("buffer flush failed: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("temp file close failed: %w", err)
	}

	// Atomic rename to final path with Windows-specific retry logic
	if err := renameWithRetry(tmpPath, cachePath); err != nil {
		return fmt.Errorf("failed to finalize cache: %w", err)
	}

	return nil
}

// renameWithRetry handles file renaming with retry logic for Windows
func renameWithRetry(src, dst string) error {
	const maxRetries = 5
	const retryDelay = 10 * time.Millisecond

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		// On Windows, we need to remove the destination file first if it exists
		// because os.Rename won't overwrite an existing file atomically
		if runtime.GOOS == "windows" {
			_ = os.Remove(dst) // Ignore error if file doesn't exist
		}

		err := os.Rename(src, dst)
		if err == nil {
			return nil
		}

		lastErr = err
		// Only retry on Windows for access denied errors
		if runtime.GOOS == "windows" {
			time.Sleep(retryDelay * time.Duration(i+1)) // Exponential backoff
			continue
		}
		// On Unix, fail immediately
		return err
	}

	return lastErr
}

// GetCache reads cached data using optimized file access
func GetCache(key string, data any) error {
	cachePath := filepath.Join(cacheDir, key+".yaml")

	//nolint:gosec // G304: Cache files are created by the application itself
	file, err := os.Open(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cache not found")
		}
		return fmt.Errorf("cache access error: %w", err)
	}
	defer func() { _ = file.Close() }()

	buffered := bufio.NewReader(file)
	if err := yaml.NewDecoder(buffered).Decode(data); err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	return nil
}

// DeleteCache removes cache files with minimal syscalls
func DeleteCache(key string) error {
	cachePath := filepath.Join(cacheDir, key+".yaml")

	if err := os.Remove(cachePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cache not found: %s", key)
		}
		return fmt.Errorf("deletion error: %w", err)
	}

	return nil
}
