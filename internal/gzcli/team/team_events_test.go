//nolint:errcheck,gosec,revive // Test file with acceptable error handling patterns
package team

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

func TestJoinTeamToGame_MultiEvents(t *testing.T) {
	// 1. Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/account/register":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		case "/api/account/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"succeeded": true}`))
		case "/api/team":
			// Simulate successful team creation request
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		case "/api/team/":
			// Simulate getting user's team
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]gzapi.Team{
				{Id: 101, Name: "MultiEventTeam"},
			})
		case "/api/game", "/api/edit/games":
			// Return list of available games
			w.WriteHeader(http.StatusOK)
			// API expects { "data": [...] } structure
			response := struct {
				Data []gzapi.Game `json:"data"`
			}{
				Data: []gzapi.Game{
					{Id: 1, Title: "Event One"},
					{Id: 2, Title: "Event Two"},
					{Id: 3, Title: "Event Three"},
				},
			}
			json.NewEncoder(w).Encode(response)
		case "/api/game/1", "/api/game/3":
			// Successfully join game 1 and 3
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		default:
			if strings.HasPrefix(r.URL.Path, "/api/game/") {
				// Fail others if any
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"success": false, "error": "Game not joinable"}`))
				return
			}
			t.Logf("Unhandled path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// 2. Setup config
	dummyApi, _ := gzapi.Init(server.URL, &gzapi.Creds{Username: "admin", Password: "password"})
	config := &mockConfig{
		url:        server.URL,
		eventId:    99, // Should be ignored if events are present
		eventTitle: "Default Event",
		inviteCode: "global-secret",
		adminApi:   dummyApi,
	}

	// 3. Setup input data with events
	teamCreds := &TeamCreds{
		Username: "MultiJoiner",
		Email:    "multi@example.com",
		TeamName: "MultiEventTeam",
		Events:   []string{"Event One", "Event Three"},
	}

	// 4. Mock dependencies
	existingTeamNames := make(map[string]struct{})
	existingUserNames := make(map[string]struct{})
	generateUsername := func(name string, _ int, existing map[string]struct{}) (string, error) {
		return strings.ToLower(strings.ReplaceAll(name, " ", "")), nil
	}

	// 5. Execute
	_, err := CreateTeamAndUser(teamCreds, config, existingTeamNames, existingUserNames, []*TeamCreds{}, false, generateUsername)

	// 6. Verify
	if err != nil {
		t.Errorf("CreateTeamAndUser() failed: %v", err)
	}
	// Note: We can't easily verify the exact API calls without a more complex mock or spy,
	// but success implies no errors logged for joining.
}
