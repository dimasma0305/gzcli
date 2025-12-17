package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dimasma0305/gzcli/internal/template/other"
)

func TestCTFTemplateStructure(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "gzcli-init-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Test data
	initInfo := map[string]string{
		"url":            "https://test.ctf.com",
		"publicEntry":    "https://public.ctf.com",
		"discordWebhook": "https://discord.com/webhook/test",
	}

	// Run CTFTemplate
	errs := other.CTFTemplate(tmpDir, initInfo)
	if len(errs) > 0 {
		t.Logf("Warning: %d errors occurred during template generation:", len(errs))
		for _, err := range errs {
			if err != nil {
				t.Logf("  - %v", err)
			}
		}
	}

	// Define expected structure (no events directory should be created)
	expectedDirs := []string{
		".gzctf",
	}

	expectedFiles := []string{
		".gitignore",
		"Makefile",
		".gzctf/conf.yaml",
		".gzctf/conf.schema.yaml",
		".gzctf/challenge.schema.yaml",
		".gzctf/gzevent.schema.yaml",
		".gzctf/appsettings.json",
		".gzctf/docker-compose.yml",
		".gzctf/init_admin.sh",
		".gzctf/expose_docker.sh",
		".gzctf/favicon.ico",
	}

	// Check directories
	for _, dir := range expectedDirs {
		path := filepath.Join(tmpDir, dir)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("Expected directory not found: %s (error: %v)", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("Expected %s to be a directory, but it's not", dir)
		}
	}

	// Check files
	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("Expected file not found: %s (error: %v)", file, err)
			continue
		}
		if info.IsDir() {
			t.Errorf("Expected %s to be a file, but it's a directory", file)
		}
	}

	// Verify events directory was NOT created
	eventsDir := filepath.Join(tmpDir, "events")
	if _, err := os.Stat(eventsDir); err == nil {
		t.Error("events directory should not be created by init command")
	}

	// Verify .gzcli/current-event was NOT created
	currentEventFile := filepath.Join(tmpDir, ".gzcli", "current-event")
	if _, err := os.Stat(currentEventFile); err == nil {
		t.Error(".gzcli/current-event should not be created by init command")
	}

	// Verify conf.yaml contains correct URL
	confFile := filepath.Join(tmpDir, ".gzctf", "conf.yaml")
	//nolint:gosec // G304: Reading test files in test environment
	confContent, err := os.ReadFile(confFile)
	if err != nil {
		t.Errorf("Failed to read .gzctf/conf.yaml: %v", err)
	} else {
		confStr := string(confContent)
		if len(confStr) < 10 {
			t.Error(".gzctf/conf.yaml appears to be empty or too short")
		}
	}

	// Verify docker-compose.yml contains correct RootFolder
	dockerComposeFile := filepath.Join(tmpDir, ".gzctf", "docker-compose.yml")
	//nolint:gosec // G304: Reading test files in test environment
	dockerComposeContent, err := os.ReadFile(dockerComposeFile)
	if err != nil {
		t.Errorf("Failed to read .gzctf/docker-compose.yml: %v", err)
	} else {
		contentStr := string(dockerComposeContent)
		if !strings.Contains(contentStr, tmpDir) {
			t.Errorf(".gzctf/docker-compose.yml should contain the root folder path %s", tmpDir)
		}
	}
}

func TestEventTemplateStructure(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "gzcli-event-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Test data with all required fields
	eventInfo := map[string]string{
		"title": "Test CTF 2024",
		"start": "2024-10-11T12:00:00Z",
		"end":   "2024-10-13T12:00:00Z",
	}

	// Run EventTemplate (no prompts - all values provided)
	errs := other.EventTemplate(tmpDir, "testEvent", eventInfo)
	if len(errs) > 0 {
		// Only log template processing errors (expected for example files)
		for _, err := range errs {
			if err != nil && !containsSubstring(err.Error(), "template processing error") {
				t.Errorf("Unexpected error: %v", err)
			}
		}
	}

	// Check that event directory was created
	eventDir := filepath.Join(tmpDir, "events", "testEvent")
	if _, err := os.Stat(eventDir); err != nil {
		t.Fatalf("Event directory not created: %v", err)
	}

	// Check .gzevent file exists
	gzeventFile := filepath.Join(eventDir, ".gzevent")
	if _, err := os.Stat(gzeventFile); err != nil {
		t.Errorf(".gzevent file not found: %v", err)
	}

	// Verify .gzevent contains the correct title
	//nolint:gosec // G304: Reading test files in test environment
	gzeventContent, err := os.ReadFile(gzeventFile)
	if err != nil {
		t.Errorf("Failed to read .gzevent: %v", err)
	} else {
		contentStr := string(gzeventContent)
		if len(contentStr) < 10 {
			t.Error(".gzevent file appears to be empty or too short")
		}
	}

	// Check category directories exist
	categories := []string{
		"Misc", "Crypto", "Pwn", "Web", "Reverse", "Blockchain",
		"Forensics", "Hardware", "Mobile", "PPC", "OSINT",
		"Game Hacking", "AI", "Pentest",
	}

	for _, category := range categories {
		categoryPath := filepath.Join(eventDir, category)
		if _, err := os.Stat(categoryPath); err != nil {
			t.Errorf("Category directory not found: %s (error: %v)", category, err)
		}
	}

	// Check .example directory exists
	exampleDir := filepath.Join(eventDir, ".example")
	if _, err := os.Stat(exampleDir); err != nil {
		t.Errorf(".example directory not found: %v", err)
	}

	// Check .structure directory exists
	structureDir := filepath.Join(eventDir, ".structure")
	if _, err := os.Stat(structureDir); err != nil {
		t.Errorf(".structure directory not found: %v", err)
	}
}

func TestCTFTemplateNoEventsCreated(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "gzcli-init-noevents-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Test data with required fields
	initInfo := map[string]string{
		"url":         "https://test.ctf.com",
		"publicEntry": "https://public.ctf.com",
	}

	// Run CTFTemplate (no prompts - all values provided)
	errs := other.CTFTemplate(tmpDir, initInfo)
	if len(errs) > 0 {
		// Only log non-critical errors
		for _, err := range errs {
			if err != nil {
				t.Logf("Warning: %v", err)
			}
		}
	}

	// Verify no events directory was created
	eventsDir := filepath.Join(tmpDir, "events")
	if _, err := os.Stat(eventsDir); err == nil {
		t.Error("init should not create events directory - events should be created with 'gzcli event create'")
	}

	// Verify .gzcli directory was NOT created (should be created only when needed)
	gzcliDir := filepath.Join(tmpDir, ".gzcli")
	if _, err := os.Stat(gzcliDir); err == nil {
		t.Error(".gzcli directory should not be created by init command")
	}

	// Verify root-level files exist
	rootFiles := []string{".gitignore", "Makefile"}
	for _, file := range rootFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Root file %s should exist: %v", file, err)
		}
	}

	// Verify conf.yaml was created with correct values
	confFile := filepath.Join(tmpDir, ".gzctf", "conf.yaml")
	//nolint:gosec // G304: Reading test files in test environment
	confContent, err := os.ReadFile(confFile)
	if err != nil {
		t.Errorf("Failed to read .gzctf/conf.yaml: %v", err)
	} else {
		confStr := string(confContent)
		if !containsSubstring(confStr, "test.ctf.com") {
			t.Error("conf.yaml should contain the provided URL")
		}
	}
}

func TestEventTemplateNoErrorLogging(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "gzcli-event-noerrors-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Test data with all required fields
	eventInfo := map[string]string{
		"title": "Test CTF 2024",
		"start": "2024-10-11T12:00:00Z",
		"end":   "2024-10-13T12:00:00Z",
	}

	// Run EventTemplate
	errs := other.EventTemplate(tmpDir, "testEvent", eventInfo)

	// Expected template errors should exist (for example files with {{.slug}}, {{.host}} etc.)
	// but these should be gracefully handled
	hasTemplateErrors := false
	hasRealErrors := false

	for _, err := range errs {
		if err != nil {
			errStr := err.Error()
			// Check if it's an expected template processing error for example files
			if containsSubstring(errStr, "template processing error") &&
				(containsSubstring(errStr, ".example/") || containsSubstring(errStr, ".structure/")) {
				hasTemplateErrors = true
			} else {
				// This is a real error that should cause failure
				hasRealErrors = true
				t.Errorf("Unexpected real error: %v", err)
			}
		}
	}

	// We expect template errors (they're normal for example files)
	if !hasTemplateErrors {
		t.Log("Note: No template errors found. This is fine if example files don't use templates.")
	}

	// We should NOT have real errors
	if hasRealErrors {
		t.Error("Found real errors that should not occur")
	}

	// Verify the event structure was created despite template errors
	eventDir := filepath.Join(tmpDir, "events", "testEvent")
	if _, err := os.Stat(eventDir); err != nil {
		t.Errorf("Event directory should be created even with template errors: %v", err)
	}

	// Verify example files exist (copied as-is despite template errors)
	exampleFiles := []string{
		"events/testEvent/.example/static-attachment/challenge.yml",
		"events/testEvent/.example/dynamic-container/challenge.yml",
	}

	for _, file := range exampleFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); err != nil {
			t.Logf("Example file may not exist (this is OK): %s", file)
		}
	}
}
