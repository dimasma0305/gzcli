//nolint:errcheck,gosec // Test file with acceptable error handling patterns
package challenge

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
)

func TestGenStructure_DirNotExist(t *testing.T) {
	// Save current dir and change to temp dir
	originalDir, _ := os.Getwd()
	tmpDir, err := os.MkdirTemp("", "structure-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)

	challenges := []config.ChallengeYaml{
		{Name: "Test Challenge", Cwd: filepath.Join(tmpDir, "challenge1")},
	}

	err = GenStructure(challenges)
	if err == nil {
		t.Error("Expected error when .structure directory doesn't exist")
	}

	if !contains(err.Error(), ".structure dir doesn't exist") {
		t.Errorf("Expected '.structure dir doesn't exist' error, got: %v", err)
	}
}

func TestGenStructure_Success(t *testing.T) {
	// Create temp directory structure
	originalDir, _ := os.Getwd()
	tmpDir, err := os.MkdirTemp("", "structure-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)

	// Create .structure directory
	structureDir := filepath.Join(tmpDir, ".structure")
	if err := os.Mkdir(structureDir, 0755); err != nil {
		t.Fatalf("Failed to create .structure dir: %v", err)
	}

	// Create a test file in .structure
	testFile := filepath.Join(structureDir, "template.txt")
	if err := os.WriteFile(testFile, []byte("template content"), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Create challenge destination directory
	challengeDir := filepath.Join(tmpDir, "challenge1")
	if err := os.Mkdir(challengeDir, 0755); err != nil {
		t.Fatalf("Failed to create challenge dir: %v", err)
	}

	challenges := []config.ChallengeYaml{
		{Name: "Test Challenge", Cwd: challengeDir},
	}

	err = GenStructure(challenges)
	if err != nil {
		t.Errorf("GenStructure() failed: %v", err)
	}

	// Verify template was copied (actual implementation depends on template.TemplateToDestination)
	// This is a basic test to ensure the function runs without error
}

// Helper function - contains is also in game_test.go
