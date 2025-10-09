// Package database provides SQLite database operations for watcher logging and state management
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dimasma0305/gzcli/internal/log"

	// Import pure-Go SQLite driver for database/sql (no CGO required)
	_ "modernc.org/sqlite"
)

// DB wraps database operations for the watcher
type DB struct {
	db      *sql.DB
	mu      sync.RWMutex
	enabled bool
	path    string
}

// New creates a new database instance
func New(dbPath string, enabled bool) *DB {
	return &DB{
		path:    dbPath,
		enabled: enabled,
	}
}

// Init initializes the database connection and creates tables
func (d *DB) Init() error {
	if !d.enabled {
		log.Info("Database logging disabled")
		return nil
	}

	dbPath := d.path
	log.Info("Initializing SQLite database: %s", dbPath)

	// Create database directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0750); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database with pragmas for better concurrency and performance
	// Use WAL mode for concurrent reads/writes and set busy timeout
	dbPath += "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings for better concurrency
	db.SetMaxOpenConns(1) // SQLite works best with a single writer
	db.SetMaxIdleConns(1)

	// Test connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.mu.Lock()
	d.db = db
	d.mu.Unlock()

	// Create tables
	if err := d.createTables(); err != nil {
		return fmt.Errorf("failed to create database tables: %w", err)
	}

	log.Info("Database initialized successfully")
	return nil
}

// createTables creates the necessary database tables
func (d *DB) createTables() error {
	d.mu.RLock()
	db := d.db
	d.mu.RUnlock()

	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Create watcher_logs table
	createLogsTable := `
		CREATE TABLE IF NOT EXISTS watcher_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			level TEXT NOT NULL,
			component TEXT NOT NULL,
			challenge TEXT,
			script TEXT,
			message TEXT NOT NULL,
			error TEXT,
			duration INTEGER
		);
		CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON watcher_logs(timestamp);
		CREATE INDEX IF NOT EXISTS idx_logs_level ON watcher_logs(level);
		CREATE INDEX IF NOT EXISTS idx_logs_challenge ON watcher_logs(challenge);
	`

	// Create challenge_states table
	createStatesTable := `
		CREATE TABLE IF NOT EXISTS challenge_states (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			challenge_name TEXT UNIQUE NOT NULL,
			status TEXT NOT NULL,
			last_update DATETIME DEFAULT CURRENT_TIMESTAMP,
			error_message TEXT,
			script_states TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_states_name ON challenge_states(challenge_name);
		CREATE INDEX IF NOT EXISTS idx_states_status ON challenge_states(status);
	`

	// Create script_executions table
	createExecutionsTable := `
		CREATE TABLE IF NOT EXISTS script_executions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			challenge_name TEXT NOT NULL,
			script_name TEXT NOT NULL,
			script_type TEXT NOT NULL,
			command TEXT NOT NULL,
			status TEXT NOT NULL,
			duration INTEGER,
			output TEXT,
			error_output TEXT,
			exit_code INTEGER
		);
		CREATE INDEX IF NOT EXISTS idx_executions_timestamp ON script_executions(timestamp);
		CREATE INDEX IF NOT EXISTS idx_executions_challenge ON script_executions(challenge_name);
		CREATE INDEX IF NOT EXISTS idx_executions_script ON script_executions(script_name);
		CREATE INDEX IF NOT EXISTS idx_executions_status ON script_executions(status);
	`

	// Create challenge_mappings table for tracking folder → GZCTF challenge ID
	createMappingsTable := `
		CREATE TABLE IF NOT EXISTS challenge_mappings (
			event TEXT NOT NULL,
			folder_path TEXT NOT NULL,
			challenge_id INTEGER NOT NULL,
			challenge_title TEXT NOT NULL,
			last_synced DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (event, folder_path)
		);
		CREATE INDEX IF NOT EXISTS idx_mappings_challenge_id ON challenge_mappings(challenge_id);
		CREATE INDEX IF NOT EXISTS idx_mappings_event ON challenge_mappings(event);
	`

	// Execute table creation statements
	if _, err := db.Exec(createLogsTable); err != nil {
		return fmt.Errorf("failed to create watcher_logs table: %w", err)
	}

	if _, err := db.Exec(createStatesTable); err != nil {
		return fmt.Errorf("failed to create challenge_states table: %w", err)
	}

	if _, err := db.Exec(createExecutionsTable); err != nil {
		return fmt.Errorf("failed to create script_executions table: %w", err)
	}

	if _, err := db.Exec(createMappingsTable); err != nil {
		return fmt.Errorf("failed to create challenge_mappings table: %w", err)
	}

	log.Info("Database tables created successfully")
	return nil
}

// ChallengeMapping represents a mapping between folder path and GZCTF challenge ID
type ChallengeMapping struct {
	Event          string
	FolderPath     string
	ChallengeID    int
	ChallengeTitle string
	LastSynced     string
}

// GetChallengeMapping retrieves a challenge mapping by event and folder path
func (d *DB) GetChallengeMapping(event, folderPath string) (*ChallengeMapping, error) {
	if !d.enabled || d.db == nil {
		return nil, fmt.Errorf("database not enabled or not initialized")
	}

	d.mu.RLock()
	db := d.db
	d.mu.RUnlock()

	query := `SELECT event, folder_path, challenge_id, challenge_title, last_synced
	          FROM challenge_mappings
	          WHERE event = ? AND folder_path = ?`

	var mapping ChallengeMapping
	err := db.QueryRow(query, event, folderPath).Scan(
		&mapping.Event,
		&mapping.FolderPath,
		&mapping.ChallengeID,
		&mapping.ChallengeTitle,
		&mapping.LastSynced,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found, not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query challenge mapping: %w", err)
	}

	return &mapping, nil
}

// SetChallengeMapping stores or updates a challenge mapping
func (d *DB) SetChallengeMapping(event, folderPath string, challengeID int, challengeTitle string) error {
	if !d.enabled || d.db == nil {
		return nil // Silently skip if database not enabled
	}

	d.mu.RLock()
	db := d.db
	d.mu.RUnlock()

	query := `INSERT INTO challenge_mappings (event, folder_path, challenge_id, challenge_title, last_synced)
	          VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	          ON CONFLICT(event, folder_path)
	          DO UPDATE SET challenge_id = ?, challenge_title = ?, last_synced = CURRENT_TIMESTAMP`

	_, err := db.Exec(query, event, folderPath, challengeID, challengeTitle, challengeID, challengeTitle)
	if err != nil {
		return fmt.Errorf("failed to set challenge mapping: %w", err)
	}

	log.DebugH3("Stored challenge mapping: %s/%s → ID %d (%s)", event, folderPath, challengeID, challengeTitle)
	return nil
}

// DeleteChallengeMapping removes a challenge mapping
func (d *DB) DeleteChallengeMapping(event, folderPath string) error {
	if !d.enabled || d.db == nil {
		return nil // Silently skip if database not enabled
	}

	d.mu.RLock()
	db := d.db
	d.mu.RUnlock()

	query := `DELETE FROM challenge_mappings WHERE event = ? AND folder_path = ?`
	_, err := db.Exec(query, event, folderPath)
	if err != nil {
		return fmt.Errorf("failed to delete challenge mapping: %w", err)
	}

	log.DebugH3("Deleted challenge mapping: %s/%s", event, folderPath)
	return nil
}

// ListChallengeMappings returns all mappings for a specific event
func (d *DB) ListChallengeMappings(event string) ([]ChallengeMapping, error) {
	if !d.enabled || d.db == nil {
		return []ChallengeMapping{}, nil
	}

	d.mu.RLock()
	db := d.db
	d.mu.RUnlock()

	query := `SELECT event, folder_path, challenge_id, challenge_title, last_synced
	          FROM challenge_mappings
	          WHERE event = ?
	          ORDER BY folder_path`

	rows, err := db.Query(query, event)
	if err != nil {
		return nil, fmt.Errorf("failed to list challenge mappings: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var mappings []ChallengeMapping
	for rows.Next() {
		var mapping ChallengeMapping
		if err := rows.Scan(
			&mapping.Event,
			&mapping.FolderPath,
			&mapping.ChallengeID,
			&mapping.ChallengeTitle,
			&mapping.LastSynced,
		); err != nil {
			return nil, fmt.Errorf("failed to scan challenge mapping: %w", err)
		}
		mappings = append(mappings, mapping)
	}

	return mappings, rows.Err()
}

// Close closes the database connection
func (d *DB) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		log.Info("Closing database connection")
		err := d.db.Close()
		d.db = nil
		return err
	}
	return nil
}

// GetDB returns the underlying database connection (for queries)
func (d *DB) GetDB() *sql.DB {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.db
}

// IsEnabled returns whether the database is enabled
func (d *DB) IsEnabled() bool {
	return d.enabled
}
