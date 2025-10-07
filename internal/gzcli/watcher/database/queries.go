package database

import (
	"database/sql"
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/types"
)

// GetRecentLogs retrieves recent log entries from the database
func (d *DB) GetRecentLogs(limit int) ([]types.WatcherLog, error) {
	db := d.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, timestamp, level, component, challenge, script, message, error, duration
		FROM watcher_logs
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var logs []types.WatcherLog
	for rows.Next() {
		var log types.WatcherLog
		var challenge, script, errorMsg sql.NullString
		var duration sql.NullInt64

		err := rows.Scan(
			&log.ID, &log.Timestamp, &log.Level, &log.Component,
			&challenge, &script, &log.Message, &errorMsg, &duration,
		)
		if err != nil {
			return nil, err
		}

		log.Challenge = challenge.String
		log.Script = script.String
		log.Error = errorMsg.String
		log.Duration = duration.Int64

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetScriptExecutions retrieves script execution records from the database
func (d *DB) GetScriptExecutions(challengeName string, limit int) ([]types.ScriptExecution, error) {
	db := d.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var query string
	var args []interface{}

	if challengeName != "" {
		query = `
			SELECT id, timestamp, challenge_name, script_name, script_type, command, status, duration, output, error_output, exit_code
			FROM script_executions
			WHERE challenge_name = ?
			ORDER BY timestamp DESC
			LIMIT ?
		`
		args = []interface{}{challengeName, limit}
	} else {
		query = `
			SELECT id, timestamp, challenge_name, script_name, script_type, command, status, duration, output, error_output, exit_code
			FROM script_executions
			ORDER BY timestamp DESC
			LIMIT ?
		`
		args = []interface{}{limit}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var executions []types.ScriptExecution
	for rows.Next() {
		var exec types.ScriptExecution
		var duration sql.NullInt64
		var output, errorOutput sql.NullString
		var exitCode sql.NullInt64

		err := rows.Scan(
			&exec.ID, &exec.Timestamp, &exec.ChallengeName, &exec.ScriptName,
			&exec.ScriptType, &exec.Command, &exec.Status, &duration,
			&output, &errorOutput, &exitCode,
		)
		if err != nil {
			return nil, err
		}

		exec.Duration = duration.Int64
		exec.Output = output.String
		exec.ErrorOutput = errorOutput.String
		exec.ExitCode = int(exitCode.Int64)

		// Compute success based on status and exit code
		exec.Success = (exec.Status == "completed" && exitCode.Valid && exitCode.Int64 == 0) ||
			(exec.Status == "completed" && !exitCode.Valid)

		executions = append(executions, exec)
	}

	return executions, rows.Err()
}
