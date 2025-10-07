//nolint:errcheck,gosec,revive // Test file with acceptable error handling patterns
package team

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

func TestGetData_FromHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test,data,from,http"))
	}))
	defer server.Close()

	data, err := GetData(server.URL)
	if err != nil {
		t.Errorf("GetData() from HTTP failed: %v", err)
	}

	if string(data) != "test,data,from,http" {
		t.Errorf("Expected 'test,data,from,http', got %s", string(data))
	}
}

func TestGetData_FromHTTP_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := GetData(server.URL)
	if err == nil {
		t.Error("Expected error for 404 response")
	}
}

func TestGetData_FromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	content := []byte("name,email,team\nJohn,john@test.com,Team1")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	data, err := GetData(tmpFile.Name())
	if err != nil {
		t.Errorf("GetData() from file failed: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("Expected %s, got %s", string(content), string(data))
	}
}

func TestGetData_FromFileProtocol(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	content := []byte("test data")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	data, err := GetData("file://" + tmpFile.Name())
	if err != nil {
		t.Errorf("GetData() with file:// protocol failed: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("Expected %s, got %s", string(content), string(data))
	}
}

func TestGetData_UnsupportedProtocol(t *testing.T) {
	_, err := GetData("ftp://example.com/file.csv")
	if err == nil {
		t.Error("Expected error for unsupported protocol")
	}

	if !strings.Contains(err.Error(), "unsupported source prefix") {
		t.Errorf("Expected 'unsupported source prefix' error, got: %v", err)
	}
}

func TestNormalizeTeamName(t *testing.T) {
	tests := []struct {
		name              string
		teamName          string
		maxLen            int
		existingTeamNames map[string]struct{}
		want              string
	}{
		{
			name:              "simple name within limit",
			teamName:          "Team1",
			maxLen:            20,
			existingTeamNames: make(map[string]struct{}),
			want:              "Team1",
		},
		{
			name:              "name too long",
			teamName:          "ThisIsAVeryLongTeamName",
			maxLen:            10,
			existingTeamNames: make(map[string]struct{}),
			want:              "ThisIsAVer",
		},
		{
			name:     "duplicate name - add counter",
			teamName: "Team1",
			maxLen:   20,
			existingTeamNames: map[string]struct{}{
				"Team1": {},
			},
			want: "Team11",
		},
		{
			name:     "multiple duplicates",
			teamName: "Team1",
			maxLen:   20,
			existingTeamNames: map[string]struct{}{
				"Team1":  {},
				"Team11": {},
			},
			want: "Team12",
		},
		{
			name:     "long name with duplicate",
			teamName: "LongTeamName",
			maxLen:   12,
			existingTeamNames: map[string]struct{}{
				"LongTeamName": {},
			},
			want: "LongTeamNam1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeTeamName(tt.teamName, tt.maxLen, tt.existingTeamNames)
			if got != tt.want {
				t.Errorf("NormalizeTeamName() = %s, want %s", got, tt.want)
			}

			if len(got) > tt.maxLen {
				t.Errorf("Result %s exceeds max length %d", got, tt.maxLen)
			}
		})
	}
}

func TestParseCSV_Valid(t *testing.T) {
	csvData := []byte(`RealName,Email,TeamName
John Doe,john@example.com,Team1
Jane Smith,jane@example.com,Team2`)

	config := &mockConfig{
		url:        "http://test.com",
		eventId:    1,
		eventTitle: "Test CTF",
	}

	createdTeams := []string{}

	createTeamFunc := func(creds *TeamCreds, cfg ConfigInterface, existingTeams, existingUsers map[string]struct{}, cache []*TeamCreds, sendEmail bool, genUser func(string, int, map[string]struct{}) (string, error)) (*TeamCreds, error) {
		createdTeams = append(createdTeams, creds.TeamName)
		return &TeamCreds{
			Username: creds.Username,
			Email:    creds.Email,
			TeamName: creds.TeamName,
			Password: "testpass",
		}, nil
	}

	generateUsername := func(name string, maxLen int, existing map[string]struct{}) (string, error) {
		return strings.ToLower(strings.ReplaceAll(name, " ", "")), nil
	}

	setCache := func(key string, value interface{}) error {
		return nil
	}

	err := ParseCSV(csvData, config, []*TeamCreds{}, false, createTeamFunc, generateUsername, setCache)
	if err != nil {
		t.Errorf("ParseCSV() failed: %v", err)
	}

	if len(createdTeams) != 2 {
		t.Errorf("Expected 2 teams to be created, got %d", len(createdTeams))
	}
}

func TestParseCSV_Empty(t *testing.T) {
	csvData := []byte(``)

	config := &mockConfig{}

	err := ParseCSV(csvData, config, nil, false, nil, nil, nil)
	if err == nil {
		t.Error("Expected error for empty CSV")
	}

	if !strings.Contains(err.Error(), "CSV is empty") {
		t.Errorf("Expected 'CSV is empty' error, got: %v", err)
	}
}

func TestParseCSV_MissingHeaders(t *testing.T) {
	csvData := []byte(`Name,Email
John,john@test.com`)

	config := &mockConfig{}

	err := ParseCSV(csvData, config, nil, false, nil, nil, nil)
	if err == nil {
		t.Error("Expected error for missing required headers")
	}

	if !strings.Contains(err.Error(), "missing required header") {
		t.Errorf("Expected 'missing required header' error, got: %v", err)
	}
}

func TestParseCSV_InvalidCSV(t *testing.T) {
	csvData := []byte(`RealName,Email,TeamName
John,"unclosed quote`)

	config := &mockConfig{}

	err := ParseCSV(csvData, config, nil, false, nil, nil, nil)
	if err == nil {
		t.Error("Expected error for invalid CSV")
	}

	if !strings.Contains(err.Error(), "failed to read CSV data") {
		t.Errorf("Expected 'failed to read CSV data' error, got: %v", err)
	}
}

// Mock implementations for testing
type mockConfig struct {
	url         string
	eventId     int
	eventTitle  string
	inviteCode  string
	appSettings *mockAppSettings
}

func (m *mockConfig) GetUrl() string        { return m.url }
func (m *mockConfig) GetEventId() int       { return m.eventId }
func (m *mockConfig) GetEventTitle() string { return m.eventTitle }
func (m *mockConfig) GetInviteCode() string { return m.inviteCode }
func (m *mockConfig) GetAppSettings() AppSettingsInterface {
	if m.appSettings == nil {
		m.appSettings = &mockAppSettings{}
	}
	return m.appSettings
}

type mockAppSettings struct{}

func (m *mockAppSettings) GetEmailConfig() EmailConfig {
	return EmailConfig{
		UserName: "test@example.com",
		Password: "testpass",
		SMTP: struct {
			Host string
			Port int
		}{
			Host: "smtp.example.com",
			Port: 587,
		},
	}
}

func TestCreateTeamAndUser_NewUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/account/register":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		case "/api/account/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		case "/api/team":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		case "/api/team/":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]gzapi.Team{
				{Id: 1, Name: "Test Team"},
			})
		case "/api/game/1":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		default:
			t.Logf("Unhandled path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	config := &mockConfig{
		url:        server.URL,
		eventId:    1,
		eventTitle: "Test CTF",
	}

	teamCreds := &TeamCreds{
		Username: "John Doe",
		Email:    "john@example.com",
		TeamName: "Test Team",
	}

	existingTeamNames := make(map[string]struct{})
	existingUserNames := make(map[string]struct{})

	generateUsername := func(name string, maxLen int, existing map[string]struct{}) (string, error) {
		username := strings.ToLower(strings.ReplaceAll(name, " ", ""))
		return username, nil
	}

	result, err := CreateTeamAndUser(teamCreds, config, existingTeamNames, existingUserNames, []*TeamCreds{}, false, generateUsername)
	if err != nil {
		t.Errorf("CreateTeamAndUser() failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got %s", result.Email)
	}

	if !result.IsTeamCreated {
		t.Error("Expected IsTeamCreated to be true")
	}
}

func TestNormalizeTeamName_CounterWithinLimit(t *testing.T) {
	existingTeamNames := map[string]struct{}{
		"Team": {},
	}

	result := NormalizeTeamName("Team", 6, existingTeamNames)

	// Should be "Team1" which fits within 6 chars
	if result != "Team1" {
		t.Errorf("Expected 'Team1', got %s", result)
	}

	if len(result) > 6 {
		t.Errorf("Result %s exceeds max length 6", result)
	}
}
