package gzcli

import (
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

func TestScriptValue_UnmarshalYAML_SimpleString(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    string
		wantErr bool
	}{
		{
			name:    "simple command",
			yaml:    "deploy: docker compose up -d",
			want:    "docker compose up -d",
			wantErr: false,
		},
		{
			name:    "script with special chars",
			yaml:    "test: ./test.sh --verbose",
			want:    "./test.sh --verbose",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data map[string]ScriptValue
			err := yaml.Unmarshal([]byte(tt.yaml), &data)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				for _, sv := range data {
					if sv.GetCommand() != tt.want {
						t.Errorf("GetCommand() = %v, want %v", sv.GetCommand(), tt.want)
					}
				}
			}
		})
	}
}

func TestScriptValue_UnmarshalYAML_ComplexObject(t *testing.T) {
	yamlData := `
deploy:
  execute: docker compose up -d
  interval: 5m
`
	var data map[string]ScriptValue
	err := yaml.Unmarshal([]byte(yamlData), &data)
	if err != nil {
		t.Fatalf("UnmarshalYAML() error = %v", err)
	}

	sv, ok := data["deploy"]
	if !ok {
		t.Fatal("deploy script not found")
	}

	if sv.GetCommand() != "docker compose up -d" {
		t.Errorf("GetCommand() = %v, want %v", sv.GetCommand(), "docker compose up -d")
	}

	expectedInterval := 5 * time.Minute
	if sv.GetInterval() != expectedInterval {
		t.Errorf("GetInterval() = %v, want %v", sv.GetInterval(), expectedInterval)
	}

	if !sv.HasInterval() {
		t.Error("HasInterval() = false, want true")
	}

	if sv.IsSimple() {
		t.Error("IsSimple() = true, want false for complex script")
	}
}

func TestScriptValue_UnmarshalYAML_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name:    "invalid yaml",
			yaml:    "deploy: [invalid",
			wantErr: true,
		},
		{
			name:    "invalid type",
			yaml:    "deploy: 123",
			wantErr: false, // Numbers are valid, will be converted to string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data map[string]ScriptValue
			err := yaml.Unmarshal([]byte(tt.yaml), &data)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScriptValue_IsSimple(t *testing.T) {
	tests := []struct {
		name   string
		script ScriptValue
		want   bool
	}{
		{
			name: "simple string",
			script: ScriptValue{
				Simple: "docker compose up",
			},
			want: true,
		},
		{
			name: "complex object",
			script: ScriptValue{
				Complex: &ScriptConfig{
					Execute:  "docker compose up",
					Interval: time.Minute,
				},
			},
			want: false,
		},
		{
			name:   "empty",
			script: ScriptValue{},
			want:   false,
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
		script ScriptValue
		want   string
	}{
		{
			name: "simple command",
			script: ScriptValue{
				Simple: "echo hello",
			},
			want: "echo hello",
		},
		{
			name: "complex command",
			script: ScriptValue{
				Complex: &ScriptConfig{
					Execute: "echo world",
				},
			},
			want: "echo world",
		},
		{
			name:   "empty",
			script: ScriptValue{},
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
		script ScriptValue
		want   time.Duration
	}{
		{
			name: "with interval",
			script: ScriptValue{
				Complex: &ScriptConfig{
					Execute:  "test",
					Interval: 10 * time.Second,
				},
			},
			want: 10 * time.Second,
		},
		{
			name: "without interval",
			script: ScriptValue{
				Simple: "test",
			},
			want: 0,
		},
		{
			name: "nil complex",
			script: ScriptValue{
				Complex: nil,
			},
			want: 0,
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
		script ScriptValue
		want   bool
	}{
		{
			name: "has interval",
			script: ScriptValue{
				Complex: &ScriptConfig{
					Execute:  "test",
					Interval: time.Minute,
				},
			},
			want: true,
		},
		{
			name: "zero interval",
			script: ScriptValue{
				Complex: &ScriptConfig{
					Execute:  "test",
					Interval: 0,
				},
			},
			want: false,
		},
		{
			name: "simple script",
			script: ScriptValue{
				Simple: "test",
			},
			want: false,
		},
		{
			name: "nil complex",
			script: ScriptValue{
				Complex: nil,
			},
			want: false,
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
