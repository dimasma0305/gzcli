//nolint:errcheck,gosec,revive // Test file with acceptable error handling patterns
package challenge

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// mockGZAPI creates a mock GZCTF API for testing
func mockGZAPI(t *testing.T, handlers map[string]http.HandlerFunc) (*gzapi.GZAPI, func()) {
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
	server := httptest.NewServer(mux)

	creds := &gzapi.Creds{
		Username: "test",
		Password: "test",
	}

	// Use Init to properly initialize the client
	api, err := gzapi.Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Failed to initialize mock API: %v", err)
	}

	cleanup := func() {
		server.Close()
	}

	return api, cleanup
}

func TestHandleChallengeAttachments_NoAttachment(t *testing.T) {
	challengeConf := ChallengeYaml{
		Name:     "Test Challenge",
		Category: "Web",
		Provide:  nil, // No attachment
	}

	challengeData := &gzapi.Challenge{
		Id:         1,
		Title:      "Test Challenge",
		Attachment: nil, // No existing attachment
	}

	api, cleanup := mockGZAPI(t, nil)
	defer cleanup()

	err := HandleChallengeAttachments(challengeConf, challengeData, api)
	if err != nil {
		t.Errorf("HandleChallengeAttachments() with no attachment error = %v, want nil", err)
	}
}

func TestHandleChallengeAttachments_RemoteURL(t *testing.T) {
	remoteURL := "https://example.com/file.zip"
	challengeConf := ChallengeYaml{
		Name:     "Test Challenge",
		Category: "Web",
		Provide:  &remoteURL,
	}

	attachmentCreated := false
	challengeData := &gzapi.Challenge{
		Id:     1,
		GameId: 123,
		Title:  "Test Challenge",
	}

	api, cleanup := mockGZAPI(t, map[string]http.HandlerFunc{
		"/api/edit/games/123/challenges/1/attachment": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}

			var form gzapi.CreateAttachmentForm
			if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if form.AttachmentType != "Remote" {
				t.Errorf("Expected AttachmentType 'Remote', got %s", form.AttachmentType)
			}

			if form.RemoteUrl != remoteURL {
				t.Errorf("Expected RemoteUrl %s, got %s", remoteURL, form.RemoteUrl)
			}

			attachmentCreated = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   1,
				"type": "Remote",
				"url":  remoteURL,
			})
		},
	})
	defer cleanup()

	challengeData.CS = api

	err := HandleChallengeAttachments(challengeConf, challengeData, api)
	if err != nil {
		t.Errorf("HandleChallengeAttachments() with remote URL error = %v, want nil", err)
	}

	if !attachmentCreated {
		t.Error("Expected attachment to be created, but it wasn't")
	}
}

func TestHandleChallengeAttachments_RemoveExisting(t *testing.T) {
	challengeConf := ChallengeYaml{
		Name:     "Test Challenge",
		Category: "Web",
		Provide:  nil, // No new attachment
	}

	attachmentRemoved := false
	challengeData := &gzapi.Challenge{
		Id:     1,
		GameId: 123,
		Title:  "Test Challenge",
		Attachment: &gzapi.Attachment{
			Id:   1,
			Type: "Local",
			Url:  "/assets/test.zip",
		},
	}

	api, cleanup := mockGZAPI(t, map[string]http.HandlerFunc{
		"/api/edit/games/123/challenges/1/attachment": func(w http.ResponseWriter, r *http.Request) {
			var form gzapi.CreateAttachmentForm
			if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if form.AttachmentType != "None" {
				t.Errorf("Expected AttachmentType 'None', got %s", form.AttachmentType)
			}

			attachmentRemoved = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		},
	})
	defer cleanup()

	challengeData.CS = api

	err := HandleChallengeAttachments(challengeConf, challengeData, api)
	if err != nil {
		t.Errorf("HandleChallengeAttachments() removing attachment error = %v, want nil", err)
	}

	if !attachmentRemoved {
		t.Error("Expected existing attachment to be removed, but it wasn't")
	}
}

func TestCreateUniqueAttachmentFile(t *testing.T) {
	// Create source file
	srcFile, err := os.CreateTemp("", "src-*.txt")
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	defer os.Remove(srcFile.Name())

	originalContent := "original file content"
	if _, err := srcFile.WriteString(originalContent); err != nil {
		t.Fatalf("Failed to write to source file: %v", err)
	}
	srcFile.Close()

	// Create destination path
	dstPath := filepath.Join(os.TempDir(), "dst-test.txt")
	defer os.Remove(dstPath)

	challengeName := "Test Challenge"

	err = CreateUniqueAttachmentFile(srcFile.Name(), dstPath, challengeName)
	if err != nil {
		t.Fatalf("CreateUniqueAttachmentFile() error = %v, want nil", err)
	}

	// Verify destination file exists
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Fatal("Destination file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	contentStr := string(content)

	// Should contain original content
	if !strings.Contains(contentStr, originalContent) {
		t.Errorf("Destination file missing original content")
	}

	// Should contain challenge name
	if !strings.Contains(contentStr, challengeName) {
		t.Errorf("Destination file missing challenge name metadata")
	}
}

func TestCreateUniqueAttachmentFile_SourceNotFound(t *testing.T) {
	dstPath := filepath.Join(os.TempDir(), "dst-test.txt")
	defer os.Remove(dstPath)

	err := CreateUniqueAttachmentFile("/nonexistent/file.txt", dstPath, "Test")
	if err == nil {
		t.Error("CreateUniqueAttachmentFile() with non-existent source expected error, got nil")
	}
}

func TestCreateAssetsIfNotExistOrDifferent_NewAsset(t *testing.T) {
	// Create a test file
	tmpFile, err := os.CreateTemp("", "asset-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := []byte("test asset content")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	assetCreated := false

	api, cleanup := mockGZAPI(t, map[string]http.HandlerFunc{
		"/api/admin/files": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			// Return empty list - no existing assets
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []gzapi.FileInfo{},
			})
		},
		"/api/assets": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			// Create new asset
			assetCreated = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]gzapi.FileInfo{
				{
					Hash: "testhash123",
					Name: filepath.Base(tmpFile.Name()),
				},
			})
		},
	})
	defer cleanup()

	fileInfo, err := CreateAssetsIfNotExistOrDifferent(tmpFile.Name(), api)
	if err != nil {
		t.Fatalf("CreateAssetsIfNotExistOrDifferent() error = %v, want nil", err)
	}

	if fileInfo == nil {
		t.Fatal("Expected file info, got nil")
	}

	if !assetCreated {
		t.Error("Expected asset to be created, but it wasn't")
	}

	if fileInfo.Hash != "testhash123" {
		t.Errorf("Expected hash 'testhash123', got %s", fileInfo.Hash)
	}
}

func TestCreateAssetsIfNotExistOrDifferent_ExistingAsset(t *testing.T) {
	// Create a test file
	tmpFile, err := os.CreateTemp("", "asset-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := []byte("test asset content")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	existingHash := "existinghash456"

	api, cleanup := mockGZAPI(t, map[string]http.HandlerFunc{
		"/api/admin/files": func(w http.ResponseWriter, r *http.Request) {
			// Return existing asset with matching hash
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []gzapi.FileInfo{
					{
						Hash: existingHash,
						Name: "existing-file.txt",
					},
				},
			})
		},
		"/api/assets": func(w http.ResponseWriter, r *http.Request) {
			// Should not be called if asset exists, but will be since hash won't match
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]gzapi.FileInfo{
				{
					Hash: "newhash789",
					Name: filepath.Base(tmpFile.Name()),
				},
			})
		},
	})
	defer cleanup()

	// This test will create a new asset since the computed hash won't match existingHash
	_, err = CreateAssetsIfNotExistOrDifferent(tmpFile.Name(), api)
	if err != nil {
		t.Fatalf("CreateAssetsIfNotExistOrDifferent() error = %v", err)
	}
}

func TestCreateAssetsIfNotExistOrDifferent_FileNotFound(t *testing.T) {
	api, cleanup := mockGZAPI(t, nil)
	defer cleanup()

	_, err := CreateAssetsIfNotExistOrDifferent("/nonexistent/file.txt", api)
	if err == nil {
		t.Error("CreateAssetsIfNotExistOrDifferent() with non-existent file expected error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get file hash") {
		t.Errorf("Expected error about file hash, got: %v", err)
	}
}

func TestCreateAssetsIfNotExistOrDifferent_APIError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "asset-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := []byte("test content")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	api, cleanup := mockGZAPI(t, map[string]http.HandlerFunc{
		"/api/admin/files": func(w http.ResponseWriter, r *http.Request) {
			// Return error
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
		},
	})
	defer cleanup()

	_, err = CreateAssetsIfNotExistOrDifferent(tmpFile.Name(), api)
	if err == nil {
		t.Error("CreateAssetsIfNotExistOrDifferent() with API error expected error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get assets") {
		t.Errorf("Expected error about failed to get assets, got: %v", err)
	}
}

func TestHandleLocalAttachment_DirectoryZip(t *testing.T) {
	// Create a temporary directory with some files
	tmpDir, err := os.MkdirTemp("", "test-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some files in the directory
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	providePath := tmpDir
	challengeConf := ChallengeYaml{
		Name:     "Test Challenge",
		Category: "Web",
		Provide:  &providePath,
		Cwd:      os.TempDir(),
	}

	challengeData := &gzapi.Challenge{
		Id:         1,
		GameId:     123,
		Title:      "Test Challenge",
		Attachment: nil,
	}

	api, cleanup := mockGZAPI(t, map[string]http.HandlerFunc{
		"/api/admin/files": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []gzapi.FileInfo{},
			})
		},
		"/api/assets": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]gzapi.FileInfo{
				{Hash: "newhash123", Name: "test.zip"},
			})
		},
		"/api/edit/games/123/challenges/1/attachment": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		},
	})
	defer cleanup()

	challengeData.CS = api

	err = HandleLocalAttachment(challengeConf, challengeData, api)
	if err != nil {
		t.Errorf("HandleLocalAttachment() with directory error = %v, want nil", err)
	}
}

func TestHandleLocalAttachment_ExistingFile(t *testing.T) {
	// Create a test zip file
	tmpFile, err := os.CreateTemp("", "test-*.zip")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := []byte("PK\x03\x04") // ZIP file signature
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	providePath := tmpFile.Name()
	challengeConf := ChallengeYaml{
		Name:     "Test Challenge",
		Category: "Web",
		Provide:  &providePath,
		Cwd:      filepath.Dir(tmpFile.Name()),
	}

	challengeData := &gzapi.Challenge{
		Id:         1,
		GameId:     123,
		Title:      "Test Challenge",
		Attachment: nil,
	}

	api, cleanup := mockGZAPI(t, map[string]http.HandlerFunc{
		"/api/admin/files": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []gzapi.FileInfo{},
			})
		},
		"/api/assets": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]gzapi.FileInfo{
				{Hash: "filehash456", Name: filepath.Base(tmpFile.Name())},
			})
		},
		"/api/edit/games/123/challenges/1/attachment": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		},
	})
	defer cleanup()

	challengeData.CS = api

	err = HandleLocalAttachment(challengeConf, challengeData, api)
	if err != nil {
		t.Errorf("HandleLocalAttachment() with existing file error = %v, want nil", err)
	}
}
