//nolint:revive // Package name is intentional to test unexported template package internals.
package template

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEmbeddedFS_WindowsPathNormalization tests that Windows-style paths
// work correctly with the embedded filesystem
func TestEmbeddedFS_WindowsPathNormalization(t *testing.T) {
	// Create an embedded filesystem wrapper
	fsys := embeddedFS{File}

	// Test cases with Windows-style paths
	testCases := []struct {
		name        string
		windowsPath string
		unixPath    string
	}{
		{
			name:        "ctf-template gitignore",
			windowsPath: `templates\others\ctf-template\.gitignore`,
			unixPath:    `templates/others/ctf-template/.gitignore`,
		},
		{
			name:        "ctf-template Makefile",
			windowsPath: `templates\others\ctf-template\Makefile`,
			unixPath:    `templates/others/ctf-template/Makefile`,
		},
		{
			name:        "event-template gzevent",
			windowsPath: `templates\others\event-template\.gzevent`,
			unixPath:    `templates/others/event-template/.gzevent`,
		},
		{
			name:        "event-template directory",
			windowsPath: `templates\others\event-template`,
			unixPath:    `templates/others/event-template`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test Stat with Windows-style path
			winInfo, winErr := fsys.Stat(tc.windowsPath)
			// Test Stat with Unix-style path
			unixInfo, unixErr := fsys.Stat(tc.unixPath)

			// Both should succeed or both should fail
			if (winErr == nil) != (unixErr == nil) {
				t.Errorf("Windows path error status (%v) doesn't match Unix path error status (%v)",
					winErr, unixErr)
				return
			}

			// If both succeeded, file info should match
			if winErr == nil && unixErr == nil {
				if winInfo.Name() != unixInfo.Name() {
					t.Errorf("File names don't match: Windows=%s, Unix=%s",
						winInfo.Name(), unixInfo.Name())
				}
				if winInfo.IsDir() != unixInfo.IsDir() {
					t.Errorf("IsDir doesn't match: Windows=%v, Unix=%v",
						winInfo.IsDir(), unixInfo.IsDir())
				}
			}
		})
	}
}

// TestEmbeddedFS_ReadFileWithWindowsPath tests reading files with Windows paths
func TestEmbeddedFS_ReadFileWithWindowsPath(t *testing.T) {
	fsys := embeddedFS{File}

	// Test reading a file with Windows-style path
	windowsPath := `templates\others\ctf-template\Makefile`
	unixPath := `templates/others/ctf-template/Makefile`

	winData, winErr := fsys.ReadFile(windowsPath)
	unixData, unixErr := fsys.ReadFile(unixPath)

	if winErr != nil {
		t.Errorf("Failed to read file with Windows path: %v", winErr)
	}
	if unixErr != nil {
		t.Errorf("Failed to read file with Unix path: %v", unixErr)
	}

	if winErr == nil && unixErr == nil {
		if string(winData) != string(unixData) {
			t.Errorf("File contents don't match between Windows and Unix paths")
		}
	}
}

// TestEmbeddedFS_ReadDirWithWindowsPath tests reading directories with Windows paths
func TestEmbeddedFS_ReadDirWithWindowsPath(t *testing.T) {
	fsys := embeddedFS{File}

	windowsPath := `templates\others\ctf-template`
	unixPath := `templates/others/ctf-template`

	winEntries, winErr := fsys.ReadDir(windowsPath)
	unixEntries, unixErr := fsys.ReadDir(unixPath)

	if winErr != nil {
		t.Errorf("Failed to read dir with Windows path: %v", winErr)
	}
	if unixErr != nil {
		t.Errorf("Failed to read dir with Unix path: %v", unixErr)
	}

	if winErr == nil && unixErr == nil {
		if len(winEntries) != len(unixEntries) {
			t.Errorf("Entry count doesn't match: Windows=%d, Unix=%d",
				len(winEntries), len(unixEntries))
		}
	}
}

// TestTemplateFSToDestination_WithDotFiles tests that dot files are correctly copied
func TestTemplateFSToDestination_WithDotFiles(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "template-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Copy ctf-template which contains dot files
	errs := TemplateFSToDestination("templates/others/ctf-template", nil, tmpDir)
	if len(errs) > 0 {
		// Some template processing errors are expected for files with {{}} placeholders
		t.Logf("Errors during template processing: %d", len(errs))
		for _, err := range errs {
			t.Logf("  - %v", err)
		}
	}

	// Check that dot files were created
	dotFiles := []string{".gitignore", ".gzctf"}
	for _, dotFile := range dotFiles {
		path := filepath.Join(tmpDir, dotFile)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Dot file not found: %s (error: %v)", dotFile, err)
		}
	}

	// Check that regular files were also created
	regularFiles := []string{"Makefile"}
	for _, file := range regularFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Regular file not found: %s (error: %v)", file, err)
		}
	}
}
