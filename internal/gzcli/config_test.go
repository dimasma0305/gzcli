package gzcli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/utils"
)

func TestNormalizeFileName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple name",
			input: "test",
			want:  "test",
		},
		{
			name:  "name with spaces",
			input: "test file",
			want:  "testfile",
		},
		{
			name:  "name with special chars",
			input: "test@#$%file",
			want:  "testfile",
		},
		{
			name:  "name with uppercase",
			input: "TestFile",
			want:  "testfile",
		},
		{
			name:  "name with dashes and underscores",
			input: "test-file_name",
			want:  "test-file_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.NormalizeFileName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeFileName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseYamlFromBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "valid yaml",
			input:   []byte("key: value\n"),
			wantErr: false,
		},
		{
			name:    "invalid yaml",
			input:   []byte("key: [invalid\n"),
			wantErr: true,
		},
		{
			name:    "empty yaml",
			input:   []byte(""),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]interface{}
			err := utils.ParseYamlFromBytes(tt.input, &result)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseYamlFromBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetFileHashHex(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	tests := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{
			name:    "valid file",
			file:    tmpFile,
			wantErr: false,
		},
		{
			name:    "non-existent file",
			file:    "/nonexistent/file.txt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.GetFileHashHex(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileHashHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Errorf("GetFileHashHex() returned empty hash")
			}
		})
	}
}
