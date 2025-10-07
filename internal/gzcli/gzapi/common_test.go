package gzapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockServer creates a test HTTP server that simulates GZCTF API
// This is shared across all test files in the gzapi package
func mockServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()

	// Add default login handler if not provided
	if handlers == nil {
		handlers = make(map[string]http.HandlerFunc)
	}
	if _, ok := handlers["/api/account/login"]; !ok {
		handlers["/api/account/login"] = func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		}
	}

	for path, handler := range handlers {
		mux.HandleFunc(path, handler)
	}
	return httptest.NewServer(mux)
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

// containsSubstring is a helper for contains
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
