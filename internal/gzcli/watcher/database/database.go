// Package database provides SQLite database operations for watcher logging and state management
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dimasma0305/gzcli/internal/log"

	// Import SQLite driver for database/sql
	_ "github.com/mattn/go-sqlite3"
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

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

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

	log.Info("Database tables created successfully")
	return nil
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
