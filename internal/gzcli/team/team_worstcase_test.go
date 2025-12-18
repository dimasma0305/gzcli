//nolint:errcheck,gosec,revive // Test file with acceptable error handling patterns
package team

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/testutil"
)

// TestParseCSV_MalformedCSV tests various malformed CSV scenarios
func TestParseCSV_MalformedCSV(t *testing.T) {
	testCases := []struct {
		name    string
		csvData string
	}{
		{"unclosed quotes", testutil.CreateMalformedCSV("unclosed_quotes")},
		{"bom characters", testutil.CreateMalformedCSV("bom")},
		{"null bytes", testutil.CreateMalformedCSV("null_bytes")},
		{"sql injection", testutil.CreateMalformedCSV("sql_injection")},
		{"xss attempt", testutil.CreateMalformedCSV("xss_attempt")},
	}

	config := &mockConfig{
		url:        "http://test.com",
		eventId:    1,
		eventTitle: "Test CTF",
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ParseCSV(
				[]byte(tc.csvData),
				config,
				&TeamConfig{ColumnMapping: ColumnMapping{RealName: "RealName", Email: "Email", TeamName: "TeamName"}},
				nil, // Changed from []*TeamCreds{} to nil
				false,
				mockCreateTeam,
				mockGenerateUsername,
				mockSetCache,
			) // Some malformed CSV should fail, others might be handled
			if err != nil {
				t.Logf("Malformed CSV %s handled: %v", tc.name, err)
			}
		})
	}
}

// TestNormalizeTeamName_ExtremeCollisions tests heavy name collision scenarios
func TestNormalizeTeamName_ExtremeCollisions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping extreme collision test in short mode")
	}

	// Create 1000 existing teams: Team, Team1, Team2, ..., Team999
	existingTeams := make(map[string]struct{})
	existingTeams["Team"] = struct{}{}
	for i := 1; i < 1000; i++ {
		existingTeams[fmt.Sprintf("Team%d", i)] = struct{}{}
	}

	// Try to create another "Team" - should get Team1000
	result := NormalizeTeamName("Team", 20, existingTeams)

	// NormalizeTeamName adds the result to existingTeams, so we need to check before it was added
	// The result should be "Team1000" since Team, Team1...Team999 already exist
	expectedResult := "Team1000"
	if result != expectedResult {
		t.Errorf("Expected %s, got %s", expectedResult, result)
	}

	if len(result) > 20 {
		t.Errorf("Result exceeds max length: %s (len=%d)", result, len(result))
	}

	t.Logf("With 1000 collisions, generated: %s", result)
}

// TestNormalizeTeamName_UnicodeNames tests Unicode in team names
func TestNormalizeTeamName_UnicodeNames(t *testing.T) {
	testCases := []struct {
		name     string
		teamName string
		maxLen   int
	}{
		{"Japanese", "日本チーム", 20},
		{"Russian", "Команда", 15},
		{"Arabic", "فريق", 10},
		{"Emoji", "Team🚀🔥💻", 15},
		{"Mixed", "Team-日本-🚀", 20},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NormalizeTeamName(tc.teamName, tc.maxLen, make(map[string]struct{}))

			// Result might be truncated or transformed
			t.Logf("Unicode team name %s -> %s", tc.teamName, result)

			if len(result) > tc.maxLen {
				t.Errorf("Result exceeds max length: %d > %d", len(result), tc.maxLen)
			}
		})
	}
}

// TestNormalizeTeamName_SpecialCharacters tests special characters
func TestNormalizeTeamName_SpecialCharacters(t *testing.T) {
	testCases := []string{
		"Team<script>alert('xss')</script>",
		"Team'; DROP TABLE teams; --",
		"Team\x00WithNull",
		"Team\nWith\nNewlines",
		"Team\t\t\tTabs",
		strings.Repeat("A", 1000), // Very long
	}

	for _, teamName := range testCases {
		t.Run(fmt.Sprintf("name-%q", teamName[:min(20, len(teamName))]), func(t *testing.T) {
			result := NormalizeTeamName(teamName, 50, make(map[string]struct{}))

			// Should produce safe output
			if strings.Contains(result, "\x00") {
				t.Error("Result contains null bytes")
			}

			if len(result) > 50 {
				t.Errorf("Result exceeds max length: %d", len(result))
			}

			t.Logf("Normalized: %q -> %q", teamName[:min(50, len(teamName))], result)
		})
	}
}

// TestParseCSV_LargeDataset tests processing large CSV files
func TestParseCSV_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	// Create CSV with 10,000 teams
	var csvBuilder strings.Builder
	csvBuilder.WriteString("RealName,Email,TeamName\n")
	for i := 0; i < 10000; i++ {
		csvBuilder.WriteString(fmt.Sprintf("User%d,user%d@test.com,Team%d\n", i, i, i))
	}

	config := &mockConfig{
		url:        "http://test.com",
		eventId:    1,
		eventTitle: "Test CTF",
	}

	createdCount := 0
	createTeam := func(creds *TeamCreds, cfg ConfigInterface, existingTeams, existingUsers map[string]struct{}, cache []*TeamCreds, sendEmail bool, genUser func(string, int, map[string]struct{}) (string, error)) (*TeamCreds, error) {
		createdCount++
		return creds, nil
	}

	err := ParseCSV([]byte(csvBuilder.String()), config, &TeamConfig{ColumnMapping: ColumnMapping{RealName: "RealName", Email: "Email", TeamName: "TeamName"}}, []*TeamCreds{}, false, createTeam, mockGenerateUsername, mockSetCache)
	if err != nil {
		t.Errorf("Large dataset failed: %v", err)
	}

	t.Logf("Processed %d teams from large dataset", createdCount)
}

// TestParseCSV_ConcurrentParsing tests concurrent CSV parsing
func TestParseCSV_ConcurrentParsing(t *testing.T) {
	csvData := []byte(`RealName,Email,TeamName
User1,user1@test.com,Team1
User2,user2@test.com,Team2
User3,user3@test.com,Team3`)

	config := &mockConfig{
		url:        "http://test.com",
		eventId:    1,
		eventTitle: "Test CTF",
	}

	// Parse same CSV concurrently
	testutil.ConcurrentTest(t, 5, 3, func(id, iter int) error {
		return ParseCSV(csvData, config, &TeamConfig{ColumnMapping: ColumnMapping{RealName: "RealName", Email: "Email", TeamName: "TeamName"}}, []*TeamCreds{}, false, mockCreateTeam, mockGenerateUsername, mockSetCache)
	})
}

// TestGetData_VeryLargeHTTPResponse tests large HTTP responses
func TestGetData_VeryLargeHTTPResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large HTTP test in short mode")
	}

	// Create large CSV content (10MB)
	largeContent := strings.Repeat("test,data,here\n", 1000000)

	server := testutil.MockServerWithDelay(t, 0, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeContent))
	})
	defer server.Close()

	data, err := GetData(server.URL)
	if err != nil {
		t.Errorf("Failed to get large data: %v", err)
	}

	if len(data) < 1000000 {
		t.Errorf("Expected large data, got %d bytes", len(data))
	}
}

// TestGetData_SlowResponse tests slow HTTP responses
func TestGetData_SlowResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping slow response test in short mode")
	}

	server := testutil.MockServerWithDelay(t, 2*time.Second, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("delayed,data"))
	})
	defer server.Close()

	data, err := GetData(server.URL)
	if err != nil {
		t.Logf("Slow response handled: %v", err)
	} else {
		t.Logf("Got data after delay: %d bytes", len(data))
	}
}

// Helper functions and mocks
func mockCreateTeam(creds *TeamCreds, cfg ConfigInterface, existingTeams, existingUsers map[string]struct{}, cache []*TeamCreds, sendEmail bool, genUser func(string, int, map[string]struct{}) (string, error)) (*TeamCreds, error) {
	return creds, nil
}

func mockGenerateUsername(name string, maxLen int, existing map[string]struct{}) (string, error) {
	return strings.ToLower(name), nil
}

func mockSetCache(key string, value interface{}) error {
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
