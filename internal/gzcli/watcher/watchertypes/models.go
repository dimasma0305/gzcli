// Package watchertypes provides type definitions for the watcher
package watchertypes

import (
	"time"
)

// Database models for persistent storage

// WatcherLog represents a log entry in the database
type WatcherLog struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Component string    `json:"component"`
	Challenge string    `json:"challenge,omitempty"`
	Script    string    `json:"script,omitempty"`
	Message   string    `json:"message"`
	Error     string    `json:"error,omitempty"`
	Duration  int64     `json:"duration,omitempty"` // milliseconds
}

// ChallengeState represents the state of a challenge in the database
type ChallengeState struct {
	ID            int64     `json:"id"`
	ChallengeName string    `json:"challenge_name"`
	Status        string    `json:"status"` // watching, updating, deploying, error
	LastUpdate    time.Time `json:"last_update"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	ScriptStates  string    `json:"script_states"` // JSON of active interval scripts
}

// ScriptExecution represents a script execution record in the database
type ScriptExecution struct {
	ID            int64     `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	ChallengeName string    `json:"challenge_name"`
	ScriptName    string    `json:"script_name"`
	ScriptType    string    `json:"script_type"` // one-time, interval
	Command       string    `json:"command"`
	Status        string    `json:"status"`             // started, completed, failed, cancelled
	Duration      int64     `json:"duration,omitempty"` // nanoseconds
	Output        string    `json:"output,omitempty"`
	ErrorOutput   string    `json:"error_output,omitempty"`
	ExitCode      int       `json:"exit_code,omitempty"`
	Success       bool      `json:"success"` // computed field based on status and exit code
}
