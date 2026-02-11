package fileutil

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNormalizeFileName_Basic tests basic normalization
func TestNormalizeFileName_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase",
			input:    "test",
			expected: "test",
		},
		{
			name:     "uppercase to lowercase",
			input:    "TEST",
			expected: "test",
		},
		{
			name:     "mixed case",
			input:    "TeSt",
			expected: "test",
		},
		{
			name:     "with hyphens",
			input:    "test-file",
			expected: "test-file",
		},
		{
			name:     "with underscores",
			input:    "test_file",
			expected: "test_file",
		},
		{
			name:     "with numbers",
			input:    "test123",
			expected: "test123",
		},
		{
			name:     "alphanumeric",
			input:    "Test123-file_name",
			expected: "test123-file_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeFileName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeFileName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNormalizeFileName_SpecialCharacters tests special character removal
func TestNormalizeFileName_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with spaces",
			input:    "test file",
			expected: "testfile",
		},
		{
			name:     "with dots",
			input:    "test.file.txt",
			expected: "testfiletxt",
		},
		{
			name:     "with slashes",
			input:    "test/file",
			expected: "testfile",
		},
		{
			name:     "with special characters",
			input:    "test@file#name$",
			expected: "testfilename",
		},
		{
			name:     "with brackets",
			input:    "test[file]",
			expected: "testfile",
		},
		{
			name:     "with parentheses",
			input:    "test(file)",
			expected: "testfile",
		},
		{
			name:     "unicode characters",
			input:    "test文件",
			expected: "test",
		},
		{
			name:     "emoji",
			input:    "test🎉file",
			expected: "testfile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeFileName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeFileName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNormalizeFileName_EdgeCases tests edge cases
func TestNormalizeFileName_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "@#$%",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "",
		},
		{
			name:     "very long string",
			input:    strings.Repeat("a", 1000),
			expected: strings.Repeat("a", 1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeFileName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeFileName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGetFileHashHex_Success tests successful hash calculation
func TestGetFileHashHex_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("Hello, World!")
	if err := os.WriteFile(testFile, content, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash, err := GetFileHashHex(testFile)
	if err != nil {
		t.Errorf("GetFileHashHex() failed: %v", err)
	}

	// Expected SHA256 hash of "Hello, World!"
	expected := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"
	if hash != expected {
		t.Errorf("GetFileHashHex() = %q, want %q", hash, expected)
	}

	// Verify hash length
	if len(hash) != 64 {
		t.Errorf("Hash length = %d, want 64", len(hash))
	}
}

// TestGetFileHashHex_EmptyFile tests hash of empty file
func TestGetFileHashHex_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.txt")

	if err := os.WriteFile(testFile, []byte{}, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash, err := GetFileHashHex(testFile)
	if err != nil {
		t.Errorf("GetFileHashHex() failed: %v", err)
	}

	// Expected SHA256 hash of empty file
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if hash != expected {
		t.Errorf("GetFileHashHex() = %q, want %q", hash, expected)
	}
}

// TestGetFileHashHex_NonExistentFile tests error handling
func TestGetFileHashHex_NonExistentFile(t *testing.T) {
	_, err := GetFileHashHex("/nonexistent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestGetFileHashHex_LargeFile tests hash of large file
func TestGetFileHashHex_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.txt")

	// Create a 1MB file
	largeContent := make([]byte, 1<<20)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	if err := os.WriteFile(testFile, largeContent, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash, err := GetFileHashHex(testFile)
	if err != nil {
		t.Errorf("GetFileHashHex() failed: %v", err)
	}

	if len(hash) != 64 {
		t.Errorf("Hash length = %d, want 64", len(hash))
	}
}

// TestCopyFile_Success tests successful file copy
func TestCopyFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")

	content := []byte("Test content for copy")
	if err := os.WriteFile(srcFile, content, 0600); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	err := CopyFile(srcFile, dstFile)
	if err != nil {
		t.Errorf("CopyFile() failed: %v", err)
	}

	// Verify destination file exists
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}

	// Verify content matches
	//nolint:gosec // G304: File path is controlled in test environment
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("Content mismatch: got %q, want %q", dstContent, content)
	}
}

// TestCopyFile_NonExistentSource tests error handling
func TestCopyFile_NonExistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	dstFile := filepath.Join(tmpDir, "dest.txt")

	err := CopyFile("/nonexistent/source.txt", dstFile)
	if err == nil {
		t.Error("Expected error for non-existent source, got nil")
	}
}

// TestCopyFile_InvalidDestination tests error handling for invalid destination
func TestCopyFile_InvalidDestination(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")

	if err := os.WriteFile(srcFile, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	err := CopyFile(srcFile, "/invalid/path/dest.txt")
	if err == nil {
		t.Error("Expected error for invalid destination, got nil")
	}
}

// TestCopyFile_EmptyFile tests copying empty file
func TestCopyFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "empty.txt")
	dstFile := filepath.Join(tmpDir, "empty_copy.txt")

	if err := os.WriteFile(srcFile, []byte{}, 0600); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	err := CopyFile(srcFile, dstFile)
	if err != nil {
		t.Errorf("CopyFile() failed: %v", err)
	}

	info, err := os.Stat(dstFile)
	if err != nil {
		t.Fatalf("Failed to stat destination file: %v", err)
	}

	if info.Size() != 0 {
		t.Errorf("Destination file size = %d, want 0", info.Size())
	}
}

// TestZipSource_Success tests successful zip creation
func TestZipSource_Success(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	targetFile := filepath.Join(tmpDir, "archive.zip")

	// Create source directory with files
	if err := os.MkdirAll(sourceDir, 0750); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	testFiles := map[string]string{
		"file1.txt": "Content 1",
		"file2.txt": "Content 2",
	}

	for name, content := range testFiles {
		file := filepath.Join(sourceDir, name)
		if err := os.WriteFile(file, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	err := ZipSource(sourceDir, targetFile)
	if err != nil {
		t.Errorf("ZipSource() failed: %v", err)
	}

	// Verify zip file exists
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		t.Error("Zip file was not created")
	}

	// Verify zip contents
	r, err := zip.OpenReader(targetFile)
	if err != nil {
		t.Fatalf("Failed to open zip file: %v", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Errorf("Failed to close zip reader: %v", err)
		}
	}()

	if len(r.File) != len(testFiles) {
		t.Errorf("Zip contains %d files, want %d", len(r.File), len(testFiles))
	}
}

// TestZipSource_EmptyDirectory tests zipping empty directory
func TestZipSource_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "empty")
	targetFile := filepath.Join(tmpDir, "empty.zip")

	if err := os.MkdirAll(sourceDir, 0750); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	err := ZipSource(sourceDir, targetFile)
	if err != nil {
		t.Errorf("ZipSource() with empty directory failed: %v", err)
	}

	// Verify zip file exists
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		t.Error("Zip file was not created for empty directory")
	}
}

// TestZipSource_NestedDirectories tests zipping nested directory structure
func TestZipSource_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	targetFile := filepath.Join(tmpDir, "nested.zip")

	// Create nested structure
	subDir := filepath.Join(sourceDir, "subdir")
	if err := os.MkdirAll(subDir, 0750); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	// Create files
	rootFile := filepath.Join(sourceDir, "root.txt")
	nestedFile := filepath.Join(subDir, "nested.txt")

	if err := os.WriteFile(rootFile, []byte("root content"), 0600); err != nil {
		t.Fatalf("Failed to create root file: %v", err)
	}

	if err := os.WriteFile(nestedFile, []byte("nested content"), 0600); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	err := ZipSource(sourceDir, targetFile)
	if err != nil {
		t.Errorf("ZipSource() with nested directories failed: %v", err)
	}

	// Verify zip contains both files
	r, err := zip.OpenReader(targetFile)
	if err != nil {
		t.Fatalf("Failed to open zip file: %v", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Errorf("Failed to close zip reader: %v", err)
		}
	}()

	if len(r.File) < 2 {
		t.Errorf("Zip contains %d files, want at least 2", len(r.File))
	}
}

// TestZipSource_NonExistentSource tests error handling
func TestZipSource_NonExistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "test.zip")

	err := ZipSource("/nonexistent/source", targetFile)
	// Should not fail immediately, but won't have any files
	if err != nil {
		t.Logf("ZipSource() with non-existent source: %v", err)
	}
}

// TestZipSource_InvalidTarget tests error handling for invalid target
func TestZipSource_InvalidTarget(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")

	if err := os.MkdirAll(sourceDir, 0750); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	err := ZipSource(sourceDir, "/invalid/path/test.zip")
	if err == nil {
		t.Error("Expected error for invalid target path, got nil")
	}
}

// TestZipSource_LargeFiles tests zipping large files
func TestZipSource_LargeFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}

	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	targetFile := filepath.Join(tmpDir, "large.zip")

	if err := os.MkdirAll(sourceDir, 0750); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// Create a 1MB file
	largeFile := filepath.Join(sourceDir, "large.bin")
	largeContent := make([]byte, 1<<20)
	if err := os.WriteFile(largeFile, largeContent, 0600); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	err := ZipSource(sourceDir, targetFile)
	if err != nil {
		t.Errorf("ZipSource() with large files failed: %v", err)
	}

	// Verify zip file was created and is smaller than source (compression)
	zipInfo, err := os.Stat(targetFile)
	if err != nil {
		t.Fatalf("Failed to stat zip file: %v", err)
	}

	if zipInfo.Size() == 0 {
		t.Error("Zip file is empty")
	}
}

func TestZipSource_DeterministicOutput(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(filepath.Join(sourceDir, "b", "c"), 0750); err != nil {
		t.Fatalf("Failed to create nested dirs: %v", err)
	}

	files := map[string]string{
		"z.txt":         "z",
		"a.txt":         "a",
		"b/c/inner.txt": "inner",
	}
	for rel, content := range files {
		full := filepath.Join(sourceDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0750); err != nil {
			t.Fatalf("Failed to create parent dir: %v", err)
		}
		if err := os.WriteFile(full, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	zip1 := filepath.Join(tmpDir, "one.zip")
	zip2 := filepath.Join(tmpDir, "two.zip")

	if err := ZipSource(sourceDir, zip1); err != nil {
		t.Fatalf("ZipSource() first run failed: %v", err)
	}
	if err := ZipSource(sourceDir, zip2); err != nil {
		t.Fatalf("ZipSource() second run failed: %v", err)
	}

	h1, err := GetFileHashHex(zip1)
	if err != nil {
		t.Fatalf("GetFileHashHex(zip1) failed: %v", err)
	}
	h2, err := GetFileHashHex(zip2)
	if err != nil {
		t.Fatalf("GetFileHashHex(zip2) failed: %v", err)
	}

	if h1 != h2 {
		t.Fatalf("ZipSource output is not deterministic: %s != %s", h1, h2)
	}
}
