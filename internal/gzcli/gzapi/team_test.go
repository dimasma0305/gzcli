package gzapi

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGZAPI_CreateTeam(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/team": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}

			var form TeamForm
			json.NewDecoder(r.Body).Decode(&form)

			if form.Name != "New Team" {
				t.Errorf("Expected team name 'New Team', got %s", form.Name)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	teamForm := &TeamForm{
		Name: "New Team",
		Bio:  "Team description",
	}

	err = api.CreateTeam(teamForm)
	if err != nil {
		t.Errorf("CreateTeam() failed: %v", err)
	}
}

func TestGZAPI_GetTeams(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/team/": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]Team{
				{Id: 1, Name: "Team 1"},
				{Id: 2, Name: "Team 2"},
			})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	teams, err := api.GetTeams()
	if err != nil {
		t.Errorf("GetTeams() failed: %v", err)
	}

	if len(teams) != 2 {
		t.Errorf("Expected 2 teams, got %d", len(teams))
	}
}

func TestGZAPI_Teams(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/admin/teams": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []Team{
					{Id: 1, Name: "Team 1"},
					{Id: 2, Name: "Team 2"},
				},
			})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	teams, err := api.Teams()
	if err != nil {
		t.Errorf("Teams() failed: %v", err)
	}

	if len(teams) != 2 {
		t.Errorf("Expected 2 teams, got %d", len(teams))
	}

	// Verify CS is set
	for _, team := range teams {
		if team.CS == nil {
			t.Error("Expected CS to be set for team")
		}
	}
}

func TestTeam_Delete(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/admin/teams/5": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"deleted": true}`))
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	team := &Team{
		Id: 5,
		CS: api,
	}

	err = team.Delete()
	if err != nil {
		t.Errorf("Team.Delete() failed: %v", err)
	}
}

func TestGZAPI_JoinGame(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/game/1": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}

			var joinModel GameJoinModel
			json.NewDecoder(r.Body).Decode(&joinModel)

			if joinModel.TeamId != 10 {
				t.Errorf("Expected TeamId 10, got %d", joinModel.TeamId)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	joinModel := &GameJoinModel{
		TeamId:   10,
		Division: "Open",
	}

	err = api.JoinGame(1, joinModel)
	if err != nil {
		t.Errorf("JoinGame() failed: %v", err)
	}
}

// Helper functions are in common_test.go
