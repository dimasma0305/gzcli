package challenge

import (
	"testing"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
)

func TestScriptValue_IsSimple(t *testing.T) {
	tests := []struct {
		name   string
		script config.ScriptValue
		want   bool
	}{
		{
			name:   "simple string command",
			script: config.ScriptValue{Simple: "echo hello"},
			want:   true,
		},
		{
			name:   "empty simple command",
			script: config.ScriptValue{Simple: ""},
			want:   false,
		},
		{
			name: "complex command",
			script: config.ScriptValue{Complex: &config.ScriptConfig{
				Execute: "docker build",
			}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.script.IsSimple(); got != tt.want {
				t.Errorf("IsSimple() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScriptValue_GetCommand(t *testing.T) {
	tests := []struct {
		name   string
		script config.ScriptValue
		want   string
	}{
		{
			name:   "simple command",
			script: config.ScriptValue{Simple: "echo test"},
			want:   "echo test",
		},
		{
			name: "complex command",
			script: config.ScriptValue{Complex: &config.ScriptConfig{
				Execute: "docker build -t test .",
			}},
			want: "docker build -t test .",
		},
		{
			name:   "empty script",
			script: config.ScriptValue{},
			want:   "",
		},
		{
			name:   "complex but nil",
			script: config.ScriptValue{Complex: nil},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.script.GetCommand(); got != tt.want {
				t.Errorf("GetCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScriptValue_GetInterval(t *testing.T) {
	tests := []struct {
		name   string
		script config.ScriptValue
		want   time.Duration
	}{
		{
			name:   "simple command has no interval",
			script: config.ScriptValue{Simple: "echo test"},
			want:   0,
		},
		{
			name: "complex with interval",
			script: config.ScriptValue{Complex: &config.ScriptConfig{
				Execute:  "test.sh",
				Interval: 5 * time.Minute,
			}},
			want: 5 * time.Minute,
		},
		{
			name: "complex without interval",
			script: config.ScriptValue{Complex: &config.ScriptConfig{
				Execute: "test.sh",
			}},
			want: 0,
		},
		{
			name:   "nil complex",
			script: config.ScriptValue{Complex: nil},
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.script.GetInterval(); got != tt.want {
				t.Errorf("GetInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScriptValue_HasInterval(t *testing.T) {
	tests := []struct {
		name   string
		script config.ScriptValue
		want   bool
	}{
		{
			name:   "simple command has no interval",
			script: config.ScriptValue{Simple: "echo test"},
			want:   false,
		},
		{
			name: "complex with interval",
			script: config.ScriptValue{Complex: &config.ScriptConfig{
				Execute:  "test.sh",
				Interval: 5 * time.Minute,
			}},
			want: true,
		},
		{
			name: "complex with zero interval",
			script: config.ScriptValue{Complex: &config.ScriptConfig{
				Execute:  "test.sh",
				Interval: 0,
			}},
			want: false,
		},
		{
			name:   "nil complex",
			script: config.ScriptValue{Complex: nil},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.script.HasInterval(); got != tt.want {
				t.Errorf("HasInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}
