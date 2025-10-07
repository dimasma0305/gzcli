package database

import (
	"encoding/json"
	"fmt"
)

// LogToDatabase logs a message to the database
func (d *DB) LogToDatabase(level, component, challenge, script, message, errorMsg string, duration int64) {
	if !d.enabled {
		return
	}

	db := d.GetDB()
	if db == nil {
		return
	}

	query := `
		INSERT INTO watcher_logs (level, component, challenge, script, message, error, duration)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query, level, component, challenge, script, message, errorMsg, duration)
	if err != nil {
		// Don't use log.Error here to avoid potential recursion
		fmt.Printf("Failed to log to database: %v\n", err)
	}
}

// UpdateChallengeState updates or inserts the state of a challenge in the database
func (d *DB) UpdateChallengeState(challengeName, status, errorMessage string, activeScripts map[string][]string) {
	if !d.enabled {
		return
	}

	db := d.GetDB()
	if db == nil {
		return
	}

	// Get current script states
	scriptStatesJSON, _ := json.Marshal(activeScripts[challengeName])

	query := `
		INSERT OR REPLACE INTO challenge_states (challenge_name, status, last_update, error_message, script_states)
		VALUES (?, ?, CURRENT_TIMESTAMP, ?, ?)
	`

	_, err := db.Exec(query, challengeName, status, errorMessage, string(scriptStatesJSON))
	if err != nil {
		fmt.Printf("Failed to update challenge state: %v\n", err)
	}
}

// LogScriptExecution logs a script execution to the database
func (d *DB) LogScriptExecution(challengeName, scriptName, scriptType, command, status string, duration int64, output, errorOutput string, exitCode int) {
	if !d.enabled {
		return
	}

	db := d.GetDB()
	if db == nil {
		return
	}

	query := `
		INSERT INTO script_executions (challenge_name, script_name, script_type, command, status, duration, output, error_output, exit_code)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query, challengeName, scriptName, scriptType, command, status, duration, output, errorOutput, exitCode)
	if err != nil {
		fmt.Printf("Failed to log script execution: %v\n", err)
	}
}
