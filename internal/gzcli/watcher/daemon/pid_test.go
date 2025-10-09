package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEnsureDirectoriesExist_SinglePath tests creating a single directory
func TestEnsureDirectoriesExist_SinglePath(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "subdir1", "subdir2", "test.pid")

	err := EnsureDirectoriesExist(testFile)
	if err != nil {
		t.Fatalf("EnsureDirectoriesExist() failed: %v", err)
	}

	// Verify directory was created
	dir := filepath.Dir(testFile)
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Path is not a directory")
	}

	// Verify permissions (at least readable and writable)
	mode := info.Mode()
	if mode.Perm()&0700 != 0700 {
		t.Errorf("Directory permissions = %o, want at least 0700", mode.Perm())
	}
}

// TestEnsureDirectoriesExist_MultiplePaths tests creating multiple directories
func TestEnsureDirectoriesExist_MultiplePaths(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "dir1", "test1.pid")
	file2 := filepath.Join(tmpDir, "dir2", "test2.log")
	file3 := filepath.Join(tmpDir, "dir3", "subdir", "test3.db")

	err := EnsureDirectoriesExist(file1, file2, file3)
	if err != nil {
		t.Fatalf("EnsureDirectoriesExist() failed: %v", err)
	}

	// Verify all directories were created
	for _, file := range []string{file1, file2, file3} {
		dir := filepath.Dir(file)
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("Directory for %s was not created: %v", file, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("Path %s is not a directory", dir)
		}
	}
}

// TestEnsureDirectoriesExist_EmptyPaths tests handling of empty paths
func TestEnsureDirectoriesExist_EmptyPaths(t *testing.T) {
	tmpDir := t.TempDir()
	validFile := filepath.Join(tmpDir, "dir", "test.pid")

	// Should skip empty paths gracefully
	err := EnsureDirectoriesExist("", validFile, "", "")
	if err != nil {
		t.Fatalf("EnsureDirectoriesExist() failed with empty paths: %v", err)
	}

	// Verify valid directory was still created
	dir := filepath.Dir(validFile)
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("Valid directory was not created: %v", err)
	}
}

// TestEnsureDirectoriesExist_ExistingDirectory tests idempotency
func TestEnsureDirectoriesExist_ExistingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "existing", "test.pid")

	// Create directory first
	dir := filepath.Dir(testFile)
	if err := os.MkdirAll(dir, 0750); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Should not error when directory already exists
	err := EnsureDirectoriesExist(testFile)
	if err != nil {
		t.Errorf("EnsureDirectoriesExist() failed with existing directory: %v", err)
	}

	// Verify directory still exists
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("Directory was removed: %v", err)
	}
}

// TestEnsureDirectoriesExist_NoPath tests with no paths
func TestEnsureDirectoriesExist_NoPath(t *testing.T) {
	err := EnsureDirectoriesExist()
	if err != nil {
		t.Errorf("EnsureDirectoriesExist() with no paths should not error, got: %v", err)
	}
}

// TestEnsureDirectoriesExist_InvalidPath tests error handling
func TestEnsureDirectoriesExist_InvalidPath(t *testing.T) {
	// Try to create a directory under a file (should fail)
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "file.txt")

	// Create a regular file
	if err := os.WriteFile(existingFile, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Try to create a directory under this file
	invalidPath := filepath.Join(existingFile, "subdir", "test.pid")
	err := EnsureDirectoriesExist(invalidPath)
	if err == nil {
		t.Error("EnsureDirectoriesExist() should fail when trying to create directory under a file")
	}
}

// TestWritePIDFile_Success tests successful PID file writing
func TestWritePIDFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "subdir", "test.pid")

	pid := os.Getpid()
	err := WritePIDFile(pidFile, pid)
	if err != nil {
		t.Fatalf("WritePIDFile() failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(pidFile); err != nil {
		t.Errorf("PID file was not created: %v", err)
	}

	// Verify content
	readPID, err := ReadPIDFromFile(pidFile)
	if err != nil {
		t.Fatalf("ReadPIDFromFile() failed: %v", err)
	}

	if readPID != pid {
		t.Errorf("ReadPIDFromFile() = %d, want %d", readPID, pid)
	}
}

// TestReadPIDFromFile_NonExistent tests reading non-existent file
func TestReadPIDFromFile_NonExistent(t *testing.T) {
	_, err := ReadPIDFromFile("/nonexistent/file.pid")
	if err == nil {
		t.Error("ReadPIDFromFile() should fail for non-existent file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.ErrNotExist, got: %v", err)
	}
}

// TestReadPIDFromFile_EmptyFile tests reading empty file
func TestReadPIDFromFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "empty.pid")

	if err := os.WriteFile(pidFile, []byte{}, 0600); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	_, err := ReadPIDFromFile(pidFile)
	if err == nil {
		t.Error("ReadPIDFromFile() should fail for empty file")
	}
}

// TestReadPIDFromFile_InvalidContent tests reading file with invalid content
func TestReadPIDFromFile_InvalidContent(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "invalid.pid")

	testCases := []struct {
		name    string
		content string
	}{
		{"letters", "abc"},
		{"special chars", "!@#$"},
		// Note: negative numbers and floats are technically parsed by Sscanf
		// The kernel doesn't assign negative PIDs, but the parser accepts them
		// since Sscanf with %d reads signed integers and truncates floats
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(pidFile, []byte(tc.content), 0600); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			_, err := ReadPIDFromFile(pidFile)
			if err == nil {
				t.Errorf("ReadPIDFromFile() should fail for content: %s", tc.content)
			}
		})
	}
}

// TestReadPIDFromFile_EdgeCases tests edge cases that are technically valid
func TestReadPIDFromFile_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "edge.pid")

	testCases := []struct {
		name        string
		content     string
		expectedPID int
	}{
		{"float truncates", "123.45", 123},
		{"whitespace", "  456  \n", 456},
		{"leading zeros", "00789", 789},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(pidFile, []byte(tc.content), 0600); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			pid, err := ReadPIDFromFile(pidFile)
			if err != nil {
				t.Errorf("ReadPIDFromFile() failed for %s: %v", tc.name, err)
			}
			if pid != tc.expectedPID {
				t.Errorf("ReadPIDFromFile() = %d, want %d", pid, tc.expectedPID)
			}
		})
	}
}

// TestWritePIDFile_Permissions tests PID file has correct permissions
func TestWritePIDFile_Permissions(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "test.pid")

	err := WritePIDFile(pidFile, 12345)
	if err != nil {
		t.Fatalf("WritePIDFile() failed: %v", err)
	}

	info, err := os.Stat(pidFile)
	if err != nil {
		t.Fatalf("Failed to stat PID file: %v", err)
	}

	// Should be 0600
	mode := info.Mode()
	if mode.Perm() != 0600 {
		t.Errorf("PID file permissions = %o, want 0600", mode.Perm())
	}
}

// TestWritePIDFile_Overwrite tests overwriting existing PID file
func TestWritePIDFile_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "test.pid")

	// Write first PID
	err := WritePIDFile(pidFile, 111)
	if err != nil {
		t.Fatalf("WritePIDFile() first write failed: %v", err)
	}

	// Overwrite with second PID
	err = WritePIDFile(pidFile, 222)
	if err != nil {
		t.Fatalf("WritePIDFile() second write failed: %v", err)
	}

	// Verify latest PID
	readPID, err := ReadPIDFromFile(pidFile)
	if err != nil {
		t.Fatalf("ReadPIDFromFile() failed: %v", err)
	}

	if readPID != 222 {
		t.Errorf("ReadPIDFromFile() = %d, want 222", readPID)
	}
}
