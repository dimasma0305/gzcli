package gzcli

import (
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

// cacheDir caches the working directory to avoid repeated lookups
var cacheDir = func() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, ".gzcli", "cache")
}()

// Cache configuration constants
const (
	maxMemoryCacheSize = 100             // Maximum number of entries in memory cache
	defaultCacheTTL    = 5 * time.Minute // Default TTL for cache entries
)

// cacheEntry represents a cached item with metadata
type cacheEntry struct {
	data      []byte
	timestamp time.Time
	key       string
}

// lruCache implements an LRU cache with TTL support
type lruCache struct {
	mu       sync.RWMutex
	capacity int
	entries  map[string]*list.Element
	lruList  *list.List
	ttl      time.Duration
}

// Global in-memory cache instance
var memoryCache = newLRUCache(maxMemoryCacheSize, defaultCacheTTL)

// Global file write mutex map for serializing writes to same key (especially needed on Windows)
var (
	fileWriteMutexes   = make(map[string]*sync.Mutex)
	fileWriteMutexesMu sync.Mutex
)

// getFileWriteMutex returns a file-specific mutex to prevent race conditions during file writes,
// which is particularly important on operating systems like Windows.
func getFileWriteMutex(key string) *sync.Mutex {
	fileWriteMutexesMu.Lock()
	defer fileWriteMutexesMu.Unlock()

	if mu, exists := fileWriteMutexes[key]; exists {
		return mu
	}

	mu := &sync.Mutex{}
	fileWriteMutexes[key] = mu
	return mu
}

// newLRUCache initializes a new LRUCache with a specified capacity and time-to-live (TTL).
func newLRUCache(capacity int, ttl time.Duration) *lruCache {
	return &lruCache{
		capacity: capacity,
		entries:  make(map[string]*list.Element, capacity),
		lruList:  list.New(),
		ttl:      ttl,
	}
}

// get retrieves a value from the LRU cache. It returns the data and a boolean indicating
// whether the key was found and the data is still valid.
func (c *lruCache) get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	entry := element.Value.(*cacheEntry)

	// Check if entry has expired
	if time.Since(entry.timestamp) > c.ttl {
		c.removeElement(element)
		return nil, false
	}

	// Move to front (most recently used)
	c.lruList.MoveToFront(element)
	return entry.data, true
}

// put adds a key-value pair to the LRU cache. If the key already exists, it updates the value
// and moves it to the front of the cache. If the cache is full, it evicts the least recently used item.
func (c *lruCache) put(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key exists, update and move to front
	if element, exists := c.entries[key]; exists {
		c.lruList.MoveToFront(element)
		entry := element.Value.(*cacheEntry)
		entry.data = data
		entry.timestamp = time.Now()
		return
	}

	// Add new entry
	entry := &cacheEntry{
		data:      data,
		timestamp: time.Now(),
		key:       key,
	}
	element := c.lruList.PushFront(entry)
	c.entries[key] = element

	// Evict least recently used if over capacity
	if c.lruList.Len() > c.capacity {
		c.evictOldest()
	}
}

// remove deletes a key-value pair from the LRU cache.
func (c *lruCache) remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, exists := c.entries[key]; exists {
		c.removeElement(element)
	}
}

// removeElement is an internal helper to remove an element from the cache. It must be called
// with the mutex held.
func (c *lruCache) removeElement(element *list.Element) {
	c.lruList.Remove(element)
	entry := element.Value.(*cacheEntry)
	delete(c.entries, entry.key)
}

// evictOldest removes the least recently used item from the cache.
func (c *lruCache) evictOldest() {
	element := c.lruList.Back()
	if element != nil {
		c.removeElement(element)
	}
}

// setCache atomically writes data to a two-tier cache (memory and disk). It handles serialization
// to YAML and ensures file writes are safe across different operating systems.
func setCache(key string, data any) error {
	// Serialize writes to the same key to prevent Windows file locking issues
	mu := getFileWriteMutex(key)
	mu.Lock()
	defer mu.Unlock()

	// Encode data to bytes using YAML
	buf, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("encoding failed: %w", err)
	}

	// Store in memory cache
	memoryCache.put(key, buf)

	// Write to disk cache
	cachePath := filepath.Join(cacheDir, key+".yaml")

	// Create cache directory with proper permissions
	if err = os.MkdirAll(filepath.Dir(cachePath), 0750); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Atomic write pattern using temp file
	tmpFile, err := os.CreateTemp(cacheDir, "tmp-")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	// Write data
	if _, err := tmpFile.Write(buf); err != nil {
		return fmt.Errorf("write failed: %w", err)
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

// renameWithRetry attempts to rename a file with retry logic, which is particularly useful on Windows
// where file locking can cause transient errors.
func renameWithRetry(src, dst string) error {
	const maxRetries = 10
	const retryDelay = 50 * time.Millisecond

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		// On Windows, we need to remove the destination file first if it exists
		// because os.Rename won't overwrite an existing file atomically
		if runtime.GOOS == "windows" {
			// Add a small delay before removing to let any file handles close
			if i > 0 {
				time.Sleep(retryDelay * time.Duration(i))
			}
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

// GetCache retrieves cached data from a two-tier cache, checking the in-memory cache first
// before falling back to the disk cache.
func GetCache(key string, data any) error {
	// Try memory cache first
	if cachedData, found := memoryCache.get(key); found {
		// Decode from cached bytes
		if err := yaml.Unmarshal(cachedData, data); err != nil {
			// If unmarshal fails, remove from cache and fall through to disk
			memoryCache.remove(key)
		} else {
			return nil
		}
	}

	// Fall back to disk cache
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

	// Read file content
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat error: %w", err)
	}

	buf := make([]byte, fileInfo.Size())
	if _, err := file.Read(buf); err != nil {
		return fmt.Errorf("read error: %w", err)
	}

	// Decode data
	if err := yaml.Unmarshal(buf, data); err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	// Store in memory cache for future access
	memoryCache.put(key, buf)

	return nil
}

// DeleteCache removes a cache entry from both the in-memory and disk caches.
func DeleteCache(key string) error {
	// Remove from memory cache
	memoryCache.remove(key)

	// Remove from disk cache
	cachePath := filepath.Join(cacheDir, key+".yaml")

	if err := os.Remove(cachePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cache not found: %s", key)
		}
		return fmt.Errorf("deletion error: %w", err)
	}

	return nil
}
