//nolint:gosec // Test files use 0644 permissions in temporary directories
package structure

import (
	"os"
	"path/filepath"
	"testing"
)

// mockChallengeData implements ChallengeData for testing
type mockChallengeData struct {
	cwd string
}

func (m *mockChallengeData) GetCwd() string {
	return m.cwd
}

// TestGenerateStructure_Success tests successful structure generation
func TestGenerateStructure_Success(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	structureDir := filepath.Join(tmpDir, ".structure")
	targetDir := filepath.Join(tmpDir, "challenge1")

	// Create .structure directory with test files
	if err := os.MkdirAll(structureDir, 0755); err != nil {
		t.Fatalf("Failed to create structure dir: %v", err)
	}

	testFile := filepath.Join(structureDir, "README.md")
	if err := os.WriteFile(testFile, []byte("Test README"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target dir: %v", err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create challenges
	challenges := []ChallengeData{
		&mockChallengeData{cwd: targetDir},
	}

	err = GenerateStructure(challenges)
	if err != nil {
		t.Errorf("GenerateStructure() failed: %v", err)
	}

	// Verify file was copied
	copiedFile := filepath.Join(targetDir, "README.md")
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Error("Expected file to be copied to target directory")
	}
}

// TestGenerateStructure_EmptyChallengeList tests with no challenges
func TestGenerateStructure_EmptyChallengeList(t *testing.T) {
	var challenges []ChallengeData

	err := GenerateStructure(challenges)
	if err == nil {
		t.Error("Expected error for empty challenge list, got nil")
	}

	expectedError := "no challenges provided"
	if err.Error() != expectedError {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedError)
	}
}

// TestGenerateStructure_MissingStructureDir tests when .structure doesn't exist
func TestGenerateStructure_MissingStructureDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory (without .structure)
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	challenges := []ChallengeData{
		&mockChallengeData{cwd: tmpDir},
	}

	err = GenerateStructure(challenges)
	if err == nil {
		t.Error("Expected error for missing .structure dir, got nil")
	}
}

// TestGenerateStructure_NilChallenge tests handling of nil challenge
func TestGenerateStructure_NilChallenge(t *testing.T) {
	tmpDir := t.TempDir()
	structureDir := filepath.Join(tmpDir, ".structure")

	// Create .structure directory
	if err := os.MkdirAll(structureDir, 0755); err != nil {
		t.Fatalf("Failed to create structure dir: %v", err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	challenges := []ChallengeData{
		nil, // Nil challenge
		&mockChallengeData{cwd: tmpDir},
	}

	err = GenerateStructure(challenges)
	// Should not return error, just skip nil challenge
	if err != nil {
		t.Errorf("GenerateStructure() with nil challenge failed: %v", err)
	}
}

// TestGenerateStructure_EmptyCwd tests handling of empty working directory
func TestGenerateStructure_EmptyCwd(t *testing.T) {
	tmpDir := t.TempDir()
	structureDir := filepath.Join(tmpDir, ".structure")

	// Create .structure directory
	if err := os.MkdirAll(structureDir, 0755); err != nil {
		t.Fatalf("Failed to create structure dir: %v", err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	challenges := []ChallengeData{
		&mockChallengeData{cwd: ""}, // Empty cwd
	}

	err = GenerateStructure(challenges)
	// Should not return error, just skip challenge with empty cwd
	if err != nil {
		t.Errorf("GenerateStructure() with empty cwd failed: %v", err)
	}
}

// TestGenerateStructure_MultipleChallenges tests with multiple challenges
func TestGenerateStructure_MultipleChallenges(t *testing.T) {
	tmpDir := t.TempDir()
	structureDir := filepath.Join(tmpDir, ".structure")

	// Create .structure directory with test files
	if err := os.MkdirAll(structureDir, 0755); err != nil {
		t.Fatalf("Failed to create structure dir: %v", err)
	}

	testFile := filepath.Join(structureDir, "template.txt")
	if err := os.WriteFile(testFile, []byte("Template content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create target directories
	targetDir1 := filepath.Join(tmpDir, "challenge1")
	targetDir2 := filepath.Join(tmpDir, "challenge2")
	targetDir3 := filepath.Join(tmpDir, "challenge3")

	for _, dir := range []string{targetDir1, targetDir2, targetDir3} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create target dir: %v", err)
		}
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	challenges := []ChallengeData{
		&mockChallengeData{cwd: targetDir1},
		&mockChallengeData{cwd: targetDir2},
		&mockChallengeData{cwd: targetDir3},
	}

	err = GenerateStructure(challenges)
	if err != nil {
		t.Errorf("GenerateStructure() with multiple challenges failed: %v", err)
	}

	// Verify files were copied to all directories
	for i, dir := range []string{targetDir1, targetDir2, targetDir3} {
		copiedFile := filepath.Join(dir, "template.txt")
		if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
			t.Errorf("Expected file to be copied to challenge%d directory", i+1)
		}
	}
}

// TestGenerateStructure_MixedValidInvalid tests with mix of valid and invalid challenges
func TestGenerateStructure_MixedValidInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	structureDir := filepath.Join(tmpDir, ".structure")
	validDir := filepath.Join(tmpDir, "valid")

	// Create .structure directory
	if err := os.MkdirAll(structureDir, 0755); err != nil {
		t.Fatalf("Failed to create structure dir: %v", err)
	}

	testFile := filepath.Join(structureDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create valid target directory
	if err := os.MkdirAll(validDir, 0755); err != nil {
		t.Fatalf("Failed to create valid dir: %v", err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	challenges := []ChallengeData{
		nil,                               // Invalid: nil
		&mockChallengeData{cwd: ""},      // Invalid: empty cwd
		&mockChallengeData{cwd: validDir}, // Valid
	}

	err = GenerateStructure(challenges)
	if err != nil {
		t.Errorf("GenerateStructure() with mixed challenges failed: %v", err)
	}

	// Verify file was copied only to valid directory
	copiedFile := filepath.Join(validDir, "test.txt")
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Error("Expected file to be copied to valid directory")
	}
}

// TestGenerateStructure_NonExistentTargetDir tests when target dir doesn't exist
func TestGenerateStructure_NonExistentTargetDir(t *testing.T) {
	tmpDir := t.TempDir()
	structureDir := filepath.Join(tmpDir, ".structure")
	nonExistentDir := filepath.Join(tmpDir, "nonexistent")

	// Create .structure directory
	if err := os.MkdirAll(structureDir, 0755); err != nil {
		t.Fatalf("Failed to create structure dir: %v", err)
	}

	testFile := filepath.Join(structureDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	challenges := []ChallengeData{
		&mockChallengeData{cwd: nonExistentDir},
	}

	err = GenerateStructure(challenges)
	// Should not fail, template system should create the directory or log error
	if err != nil {
		t.Errorf("GenerateStructure() with nonexistent target failed: %v", err)
	}
}

// TestGenerateStructure_NestedStructure tests with nested directory structure
func TestGenerateStructure_NestedStructure(t *testing.T) {
	tmpDir := t.TempDir()
	structureDir := filepath.Join(tmpDir, ".structure")
	subDir := filepath.Join(structureDir, "subdir")
	targetDir := filepath.Join(tmpDir, "challenge")

	// Create nested .structure directory
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create nested structure dir: %v", err)
	}

	// Create files in nested structure
	rootFile := filepath.Join(structureDir, "root.txt")
	nestedFile := filepath.Join(subDir, "nested.txt")

	if err := os.WriteFile(rootFile, []byte("root"), 0644); err != nil {
		t.Fatalf("Failed to create root file: %v", err)
	}

	if err := os.WriteFile(nestedFile, []byte("nested"), 0644); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target dir: %v", err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	challenges := []ChallengeData{
		&mockChallengeData{cwd: targetDir},
	}

	err = GenerateStructure(challenges)
	if err != nil {
		t.Errorf("GenerateStructure() with nested structure failed: %v", err)
	}

	// Verify nested structure was copied
	copiedRoot := filepath.Join(targetDir, "root.txt")
	copiedNested := filepath.Join(targetDir, "subdir", "nested.txt")

	if _, err := os.Stat(copiedRoot); os.IsNotExist(err) {
		t.Error("Expected root file to be copied")
	}

	if _, err := os.Stat(copiedNested); os.IsNotExist(err) {
		t.Error("Expected nested file to be copied")
	}
}

// TestChallengeData_Interface tests the ChallengeData interface
func TestChallengeData_Interface(t *testing.T) {
	challenge := &mockChallengeData{cwd: "/test/path"}

	if challenge.GetCwd() != "/test/path" {
		t.Errorf("GetCwd() = %q, want %q", challenge.GetCwd(), "/test/path")
	}
}

// TestGenerateStructure_PermissionHandling tests handling of permission errors
func TestGenerateStructure_PermissionHandling(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	structureDir := filepath.Join(tmpDir, ".structure")
	readOnlyDir := filepath.Join(tmpDir, "readonly")

	// Create .structure directory
	if err := os.MkdirAll(structureDir, 0755); err != nil {
		t.Fatalf("Failed to create structure dir: %v", err)
	}

	testFile := filepath.Join(structureDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create read-only directory
	if err := os.MkdirAll(readOnlyDir, 0555); err != nil {
		t.Fatalf("Failed to create readonly dir: %v", err)
	}
	defer func() { _ = os.Chmod(readOnlyDir, 0755) }() // Cleanup

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	challenges := []ChallengeData{
		&mockChallengeData{cwd: readOnlyDir},
	}

	err = GenerateStructure(challenges)
	// Should not return error, but should log the error and continue
	if err != nil {
		t.Errorf("GenerateStructure() with permission error failed: %v", err)
	}
}
