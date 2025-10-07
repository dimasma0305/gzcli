package gzcli

import (
	"testing"
)

// Note: Cache tests require refactoring cache.go to accept a working directory parameter
// instead of using global state. Skipping detailed cache tests for now.
// The cache functionality is tested indirectly through integration tests.

func TestDeleteCache_NotExists(t *testing.T) {
	// Test that DeleteCache returns appropriate error for non-existent cache
	err := DeleteCache("definitely-does-not-exist-12345")
	if err == nil {
		t.Error("DeleteCache() should return error for non-existent cache")
	}
}
