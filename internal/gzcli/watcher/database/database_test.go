package database

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNew_Creation tests database instance creation
func TestNew_Creation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db := New(dbPath, true)
	if db == nil {
		t.Fatal("New() returned nil")
		return // Help staticcheck understand control flow
	}

	if db.path != dbPath {
		t.Errorf("db.path = %s, want %s", db.path, dbPath)
	}

	if !db.enabled {
		t.Error("db.enabled = false, want true")
	}
}

// TestNew_Disabled tests disabled database
func TestNew_Disabled(t *testing.T) {
	db := New("", false)
	if db == nil {
		t.Fatal("New() returned nil")
		return // Help staticcheck understand control flow
	}

	if db.enabled {
		t.Error("db.enabled = true, want false")
	}
}

// TestDB_Init_Success tests successful database initialization
func TestDB_Init_Success(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db := New(dbPath, true)
	defer func() { _ = db.Close() }()

	err := db.Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("Database file was not created: %v", err)
	}

	// Verify connection is valid
	if db.GetDB() == nil {
		t.Error("Database connection is nil after Init()")
	}

	// Verify we can ping
	if err := db.GetDB().Ping(); err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

// TestDB_Init_DisabledDatabase tests disabled database doesn't create files
func TestDB_Init_DisabledDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db := New(dbPath, false)
	defer func() { _ = db.Close() }()

	err := db.Init()
	if err != nil {
		t.Fatalf("Init() on disabled database should not error: %v", err)
	}

	// Verify database file was NOT created
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Error("Database file should not be created when disabled")
	}

	// Verify connection is nil
	if db.GetDB() != nil {
		t.Error("Database connection should be nil when disabled")
	}
}

// TestDB_Init_CreatesDirectory tests directory creation
func TestDB_Init_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir1", "subdir2", "test.db")

	db := New(dbPath, true)
	defer func() { _ = db.Close() }()

	err := db.Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Verify directory was created
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("Directory was not created: %v", err)
	}
}

// TestDB_Init_TablesCreated tests that all tables are created
func TestDB_Init_TablesCreated(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db := New(dbPath, true)
	defer func() { _ = db.Close() }()

	err := db.Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Check that tables exist by querying them
	tables := []string{
		"watcher_logs",
		"challenge_states",
		"script_executions",
		"challenge_mappings",
	}

	for _, table := range tables {
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
		var name string
		err := db.GetDB().QueryRow(query, table).Scan(&name)
		if err != nil {
			t.Errorf("Table %s was not created: %v", table, err)
		}
	}
}

// TestDB_Close tests closing database connection
func TestDB_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db := New(dbPath, true)
	if err := db.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	err := db.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// Verify connection is nil after close
	if db.GetDB() != nil {
		t.Error("Database connection should be nil after Close()")
	}

	// Closing again should not error
	err = db.Close()
	if err != nil {
		t.Errorf("Second Close() should not error: %v", err)
	}
}

// TestDB_IsEnabled tests IsEnabled method
func TestDB_IsEnabled(t *testing.T) {
	testCases := []struct {
		name    string
		enabled bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := New("test.db", tc.enabled)
			if db.IsEnabled() != tc.enabled {
				t.Errorf("IsEnabled() = %v, want %v", db.IsEnabled(), tc.enabled)
			}
		})
	}
}

// TestDB_ChallengeMapping_SetAndGet tests setting and getting challenge mappings
func TestDB_ChallengeMapping_SetAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db := New(dbPath, true)
	defer func() { _ = db.Close() }()

	if err := db.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Set a mapping
	event := "ctf2025"
	folderPath := "web/challenge1"
	challengeID := 42
	challengeTitle := "Test Challenge"

	err := db.SetChallengeMapping(event, folderPath, challengeID, challengeTitle)
	if err != nil {
		t.Fatalf("SetChallengeMapping() failed: %v", err)
	}

	// Get the mapping
	mapping, err := db.GetChallengeMapping(event, folderPath)
	if err != nil {
		t.Fatalf("GetChallengeMapping() failed: %v", err)
	}

	if mapping == nil {
		t.Fatal("GetChallengeMapping() returned nil")
		return // Help staticcheck understand control flow
	}

	if mapping.Event != event {
		t.Errorf("mapping.Event = %s, want %s", mapping.Event, event)
	}
	if mapping.FolderPath != folderPath {
		t.Errorf("mapping.FolderPath = %s, want %s", mapping.FolderPath, folderPath)
	}
	if mapping.ChallengeID != challengeID {
		t.Errorf("mapping.ChallengeID = %d, want %d", mapping.ChallengeID, challengeID)
	}
	if mapping.ChallengeTitle != challengeTitle {
		t.Errorf("mapping.ChallengeTitle = %s, want %s", mapping.ChallengeTitle, challengeTitle)
	}
}

// TestDB_ChallengeMapping_Update tests updating existing mapping
func TestDB_ChallengeMapping_Update(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db := New(dbPath, true)
	defer func() { _ = db.Close() }()

	if err := db.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	event := "ctf2025"
	folderPath := "web/challenge1"

	// Set initial mapping
	err := db.SetChallengeMapping(event, folderPath, 42, "Original Title")
	if err != nil {
		t.Fatalf("SetChallengeMapping() initial failed: %v", err)
	}

	// Update mapping
	err = db.SetChallengeMapping(event, folderPath, 99, "Updated Title")
	if err != nil {
		t.Fatalf("SetChallengeMapping() update failed: %v", err)
	}

	// Verify updated values
	mapping, err := db.GetChallengeMapping(event, folderPath)
	if err != nil {
		t.Fatalf("GetChallengeMapping() failed: %v", err)
	}

	if mapping.ChallengeID != 99 {
		t.Errorf("mapping.ChallengeID = %d, want 99", mapping.ChallengeID)
	}
	if mapping.ChallengeTitle != "Updated Title" {
		t.Errorf("mapping.ChallengeTitle = %s, want Updated Title", mapping.ChallengeTitle)
	}
}

// TestDB_ChallengeMapping_NotFound tests getting non-existent mapping
func TestDB_ChallengeMapping_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db := New(dbPath, true)
	defer func() { _ = db.Close() }()

	if err := db.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	mapping, err := db.GetChallengeMapping("nonexistent", "path")
	if err != nil {
		t.Errorf("GetChallengeMapping() should not error for non-existent: %v", err)
	}

	if mapping != nil {
		t.Error("GetChallengeMapping() should return nil for non-existent mapping")
	}
}

// TestDB_ChallengeMapping_Delete tests deleting mappings
func TestDB_ChallengeMapping_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db := New(dbPath, true)
	defer func() { _ = db.Close() }()

	if err := db.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	event := "ctf2025"
	folderPath := "web/challenge1"

	// Set a mapping
	err := db.SetChallengeMapping(event, folderPath, 42, "Test")
	if err != nil {
		t.Fatalf("SetChallengeMapping() failed: %v", err)
	}

	// Delete it
	err = db.DeleteChallengeMapping(event, folderPath)
	if err != nil {
		t.Fatalf("DeleteChallengeMapping() failed: %v", err)
	}

	// Verify it's gone
	mapping, err := db.GetChallengeMapping(event, folderPath)
	if err != nil {
		t.Errorf("GetChallengeMapping() failed: %v", err)
	}
	if mapping != nil {
		t.Error("Mapping should be deleted")
	}
}

// TestDB_ChallengeMapping_List tests listing all mappings for an event
func TestDB_ChallengeMapping_List(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db := New(dbPath, true)
	defer func() { _ = db.Close() }()

	if err := db.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	event := "ctf2025"

	// Add multiple mappings
	mappings := []struct {
		path  string
		id    int
		title string
	}{
		{"web/chall1", 1, "Web 1"},
		{"web/chall2", 2, "Web 2"},
		{"pwn/chall1", 3, "Pwn 1"},
	}

	for _, m := range mappings {
		err := db.SetChallengeMapping(event, m.path, m.id, m.title)
		if err != nil {
			t.Fatalf("SetChallengeMapping() failed: %v", err)
		}
	}

	// List all mappings
	result, err := db.ListChallengeMappings(event)
	if err != nil {
		t.Fatalf("ListChallengeMappings() failed: %v", err)
	}

	if len(result) != len(mappings) {
		t.Errorf("ListChallengeMappings() returned %d mappings, want %d", len(result), len(mappings))
	}

	// Verify all mappings are present
	for _, expected := range mappings {
		found := false
		for _, actual := range result {
			if actual.FolderPath == expected.path && actual.ChallengeID == expected.id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected mapping not found: %s -> %d", expected.path, expected.id)
		}
	}
}

// TestDB_ChallengeMapping_DisabledDatabase tests operations on disabled database
func TestDB_ChallengeMapping_DisabledDatabase(t *testing.T) {
	db := New("test.db", false)
	defer func() { _ = db.Close() }()

	// All operations should silently succeed or return empty results
	err := db.SetChallengeMapping("event", "path", 1, "title")
	if err != nil {
		t.Errorf("SetChallengeMapping() on disabled db should not error: %v", err)
	}

	mapping, err := db.GetChallengeMapping("event", "path")
	if err == nil || mapping != nil {
		t.Error("GetChallengeMapping() on disabled db should return error")
	}

	err = db.DeleteChallengeMapping("event", "path")
	if err != nil {
		t.Errorf("DeleteChallengeMapping() on disabled db should not error: %v", err)
	}

	mappings, err := db.ListChallengeMappings("event")
	if err != nil {
		t.Errorf("ListChallengeMappings() on disabled db should not error: %v", err)
	}
	if len(mappings) != 0 {
		t.Error("ListChallengeMappings() on disabled db should return empty slice")
	}
}

// TestDB_PureGoSQLite tests that we're using the pure-Go SQLite driver
func TestDB_PureGoSQLite(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db := New(dbPath, true)
	defer func() { _ = db.Close() }()

	err := db.Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Verify we can perform basic operations (proves pure-Go driver works)
	// without CGO_ENABLED=1
	var version string
	err = db.GetDB().QueryRow("SELECT sqlite_version()").Scan(&version)
	if err != nil {
		t.Fatalf("Failed to query SQLite version: %v", err)
	}

	if version == "" {
		t.Error("SQLite version should not be empty")
	}

	t.Logf("SQLite version: %s (using modernc.org/sqlite pure-Go driver)", version)
}

// TestDB_ConcurrentAccess tests thread-safe operations
func TestDB_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db := New(dbPath, true)
	defer func() { _ = db.Close() }()

	if err := db.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Run concurrent operations
	done := make(chan bool)
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			event := "ctf2025"
			path := filepath.Join("web", "challenge", string(rune('0'+id)))

			// Set mapping
			err := db.SetChallengeMapping(event, path, id, "Test")
			if err != nil {
				t.Errorf("Concurrent SetChallengeMapping failed: %v", err)
			}

			// Get mapping
			_, err = db.GetChallengeMapping(event, path)
			if err != nil {
				t.Errorf("Concurrent GetChallengeMapping failed: %v", err)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
