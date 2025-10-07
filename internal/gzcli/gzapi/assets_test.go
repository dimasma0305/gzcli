package gzapi

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

func TestGZAPI_CreateAssets(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "asset-*.zip")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	content := []byte("test asset content")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/assets": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}

			if err := r.ParseMultipartForm(32 << 20); err != nil {
				t.Errorf("Failed to parse multipart form: %v", err)
			}

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode([]FileInfo{
				{Hash: "abc123hash", Name: "asset.zip"},
			}); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	fileInfo, err := api.CreateAssets(tmpFile.Name())
	if err != nil {
		t.Errorf("CreateAssets() failed: %v", err)
	}

	if len(fileInfo) != 1 {
		t.Errorf("Expected 1 file info, got %d", len(fileInfo))
	}

	if fileInfo[0].Hash != "abc123hash" {
		t.Errorf("Expected hash 'abc123hash', got %s", fileInfo[0].Hash)
	}
}

func TestGZAPI_GetAssets(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/admin/files": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []FileInfo{
					{Hash: "hash1", Name: "file1.zip"},
					{Hash: "hash2", Name: "file2.tar.gz"},
				},
			}); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	assets, err := api.GetAssets()
	if err != nil {
		t.Errorf("GetAssets() failed: %v", err)
	}

	if len(assets) != 2 {
		t.Errorf("Expected 2 assets, got %d", len(assets))
	}

	if assets[0].Hash != "hash1" {
		t.Errorf("Expected hash 'hash1', got %s", assets[0].Hash)
	}
}

// Helper functions are in common_test.go
