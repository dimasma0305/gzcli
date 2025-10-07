package gzapi

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestGZAPI_GetGames(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []Game{
					{Id: 1, Title: "Game 1"},
					{Id: 2, Title: "Game 2"},
				},
			}); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	games, err := api.GetGames()
	if err != nil {
		t.Errorf("GetGames() failed: %v", err)
	}

	if len(games) != 2 {
		t.Errorf("Expected 2 games, got %d", len(games))
	}

	// Verify CS is set for each game
	for _, game := range games {
		if game.CS == nil {
			t.Error("Expected CS to be set for game")
		}
	}
}

func TestGZAPI_GetGameById(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games/5": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Game{
				Id:    5,
				Title: "Test Game",
			})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	game, err := api.GetGameById(5)
	if err != nil {
		t.Errorf("GetGameById() failed: %v", err)
	}

	if game.Id != 5 {
		t.Errorf("Expected game ID 5, got %d", game.Id)
	}

	if game.CS == nil {
		t.Error("Expected CS to be set")
	}
}

func TestGZAPI_GetGameByTitle(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []Game{
					{Id: 1, Title: "Game 1"},
					{Id: 2, Title: "Target Game"},
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

	game, err := api.GetGameByTitle("Target Game")
	if err != nil {
		t.Errorf("GetGameByTitle() failed: %v", err)
	}

	if game.Id != 2 {
		t.Errorf("Expected game ID 2, got %d", game.Id)
	}

	if game.Title != "Target Game" {
		t.Errorf("Expected title 'Target Game', got %s", game.Title)
	}
}

func TestGZAPI_GetGameByTitle_NotFound(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []Game{
					{Id: 1, Title: "Game 1"},
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

	_, err = api.GetGameByTitle("Nonexistent Game")
	if err == nil {
		t.Error("Expected error for nonexistent game")
	}

	if !contains(err.Error(), "game not found") {
		t.Errorf("Expected 'game not found' error, got: %v", err)
	}
}

func TestGame_Delete(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games/5": func(w http.ResponseWriter, r *http.Request) {
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

	game := &Game{
		Id: 5,
		CS: api,
	}

	err = game.Delete()
	if err != nil {
		t.Errorf("Game.Delete() failed: %v", err)
	}
}

func TestGame_Update(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games/5": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}

			var updatedGame Game
			if err := json.NewDecoder(r.Body).Decode(&updatedGame); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// Verify time fields are in UTC
			if updatedGame.Start.Location() != time.UTC {
				t.Error("Expected Start time to be in UTC")
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

	game := &Game{
		Id: 5,
		CS: api,
	}

	updateData := &Game{
		Title:   "Updated Game",
		Start:   CustomTime{time.Now()},
		End:     CustomTime{time.Now().Add(24 * time.Hour)},
		Summary: "Updated summary",
	}

	err = game.Update(updateData)
	if err != nil {
		t.Errorf("Game.Update() failed: %v", err)
	}
}

func TestGame_UploadPoster(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "poster-*.png")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	content := []byte("fake image data")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games/5/poster": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}

			if err := r.ParseMultipartForm(32 << 20); err != nil {
				t.Errorf("Failed to parse multipart form: %v", err)
			}

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode("/uploads/poster.png"); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	game := &Game{
		Id: 5,
		CS: api,
	}

	path, err := game.UploadPoster(tmpFile.Name())
	if err != nil {
		t.Errorf("Game.UploadPoster() failed: %v", err)
	}

	if path != "/uploads/poster.png" {
		t.Errorf("Expected path '/uploads/poster.png', got %s", path)
	}
}

func TestGZAPI_CreateGame(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/edit/games": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}

			var form CreateGameForm
			if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Game{
				Id:    100,
				Title: form.Title,
			})
		},
	})
	defer server.Close()

	creds := &Creds{Username: "test", Password: "test"}
	api, err := Init(server.URL, creds)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	form := CreateGameForm{
		Title: "New Game",
		Start: time.Now(),
		End:   time.Now().Add(48 * time.Hour),
	}

	game, err := api.CreateGame(form)
	if err != nil {
		t.Errorf("CreateGame() failed: %v", err)
	}

	if game.Id != 100 {
		t.Errorf("Expected game ID 100, got %d", game.Id)
	}

	if game.CS == nil {
		t.Error("Expected CS to be set")
	}
}

func TestGame_JoinGame(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/game/1": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}

			var joinModel GameJoinModel
			if err := json.NewDecoder(r.Body).Decode(&joinModel); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if joinModel.TeamId != 5 {
				t.Errorf("Expected TeamId 5, got %d", joinModel.TeamId)
			}

			if joinModel.Division != "Open" {
				t.Errorf("Expected Division 'Open', got %s", joinModel.Division)
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

	game := &Game{Id: 1, CS: api}

	err = game.JoinGame(5, "Open", "")
	if err != nil {
		t.Errorf("Game.JoinGame() failed: %v", err)
	}
}

func TestCustomTime_UnmarshalJSON_Milliseconds(t *testing.T) {
	jsonData := []byte(`1609459200000`) // 2021-01-01 00:00:00 UTC in ms

	var ct CustomTime
	err := json.Unmarshal(jsonData, &ct)
	if err != nil {
		t.Errorf("UnmarshalJSON() failed: %v", err)
	}

	expected := time.Unix(1609459200, 0)
	if !ct.Time.Equal(expected) {
		t.Errorf("Expected time %v, got %v", expected, ct.Time)
	}
}

func TestCustomTime_UnmarshalJSON_RFC3339(t *testing.T) {
	jsonData := []byte(`"2024-01-01T12:00:00Z"`)

	var ct CustomTime
	err := json.Unmarshal(jsonData, &ct)
	if err != nil {
		t.Errorf("UnmarshalJSON() failed: %v", err)
	}

	expected, _ := time.Parse(time.RFC3339, "2024-01-01T12:00:00Z")
	if !ct.Time.Equal(expected) {
		t.Errorf("Expected time %v, got %v", expected, ct.Time)
	}
}

func TestCustomTime_UnmarshalJSON_Invalid(t *testing.T) {
	jsonData := []byte(`"invalid-time-format"`)

	var ct CustomTime
	err := json.Unmarshal(jsonData, &ct)
	if err == nil {
		t.Error("Expected error for invalid time format")
	}

	if !contains(err.Error(), "invalid time format") {
		t.Errorf("Expected 'invalid time format' error, got: %v", err)
	}
}

func TestGame_GetScoreboard(t *testing.T) {
	server := mockServer(t, map[string]http.HandlerFunc{
		"/api/game/1/scoreboard": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Scoreboard{
				Challenges: map[string][]ScoreboardChallenge{
					"Web": {
						{Score: 100, Category: "Web", Title: "Challenge 1"},
					},
				},
				Items: []ScoreboardItem{
					{Name: "Team 1", Rank: 1, Score: 100},
					{Name: "Team 2", Rank: 2, Score: 50},
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

	game := &Game{Id: 1, CS: api}

	scoreboard, err := game.GetScoreboard()
	if err != nil {
		t.Errorf("GetScoreboard() failed: %v", err)
	}

	if len(scoreboard.Items) != 2 {
		t.Errorf("Expected 2 scoreboard items, got %d", len(scoreboard.Items))
	}

	if scoreboard.Items[0].Rank != 1 {
		t.Errorf("Expected rank 1, got %d", scoreboard.Items[0].Rank)
	}
}

// Helper functions are in common_test.go
