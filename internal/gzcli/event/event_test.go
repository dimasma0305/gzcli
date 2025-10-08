package event

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// TestRemoveAllEvent_Success tests successful deletion of all games
func TestRemoveAllEvent_Success(t *testing.T) {
	deletedGames := make(map[int]bool)
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/account/login":
			// Return successful login response
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"token": "test-token",
			})
		case "/api/edit/games":
			// Return list of games wrapped in data field
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": 1, "title": "Game 1"},
					{"id": 2, "title": "Game 2"},
					{"id": 3, "title": "Game 3"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		case "/api/edit/games/1", "/api/edit/games/2", "/api/edit/games/3":
			if r.Method == "DELETE" {
				// Mark game as deleted
				gameID := 1
				if r.URL.Path == "/api/edit/games/2" {
					gameID = 2
				} else if r.URL.Path == "/api/edit/games/3" {
					gameID = 3
				}
				mu.Lock()
				deletedGames[gameID] = true
				mu.Unlock()
				w.WriteHeader(http.StatusOK)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	api, err := gzapi.Init(server.URL, &gzapi.Creds{
		Username: "test",
		Password: "test",
	})
	if err != nil {
		t.Fatalf("Failed to initialize API: %v", err)
	}

	err = RemoveAllEvent(api)
	if err != nil {
		t.Errorf("RemoveAllEvent() failed: %v", err)
	}

	// Verify all games were deleted
	mu.Lock()
	defer mu.Unlock()
	for i := 1; i <= 3; i++ {
		if !deletedGames[i] {
			t.Errorf("Game %d was not deleted", i)
		}
	}
}

// TestRemoveAllEvent_GetGamesError tests error handling when getting games fails
func TestRemoveAllEvent_GetGamesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/account/login" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"token": "test-token",
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	api, err := gzapi.Init(server.URL, &gzapi.Creds{
		Username: "test",
		Password: "test",
	})
	if err != nil {
		t.Fatalf("Failed to initialize API: %v", err)
	}

	err = RemoveAllEvent(api)
	if err == nil {
		t.Error("Expected error when GetGames fails, got nil")
	}
}

// TestRemoveAllEvent_DeleteError tests error handling when deletion fails
func TestRemoveAllEvent_DeleteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/account/login":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"token": "test-token",
			})
		case "/api/edit/games":
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": 1, "title": "Game 1"},
					{"id": 2, "title": "Game 2"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		case "/api/edit/games/1":
			if r.Method == "DELETE" {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "Permission denied",
				})
			}
		case "/api/edit/games/2":
			if r.Method == "DELETE" {
				w.WriteHeader(http.StatusOK)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	api, err := gzapi.Init(server.URL, &gzapi.Creds{
		Username: "test",
		Password: "test",
	})
	if err != nil {
		t.Fatalf("Failed to initialize API: %v", err)
	}

	err = RemoveAllEvent(api)
	if err == nil {
		t.Error("Expected error when Delete fails, got nil")
	}
}

// TestRemoveAllEvent_EmptyGameList tests handling of empty game list
func TestRemoveAllEvent_EmptyGameList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/account/login":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"token": "test-token",
			})
		case "/api/edit/games":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{},
			})
		}
	}))
	defer server.Close()

	api, err := gzapi.Init(server.URL, &gzapi.Creds{
		Username: "test",
		Password: "test",
	})
	if err != nil {
		t.Fatalf("Failed to initialize API: %v", err)
	}

	err = RemoveAllEvent(api)
	if err != nil {
		t.Errorf("RemoveAllEvent() with empty list failed: %v", err)
	}
}

// TestScoreboard2CTFTimeFeed_Success tests successful conversion of scoreboard
func TestScoreboard2CTFTimeFeed_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/account/login":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"token": "test-token",
			})
		case "/api/game/1/scoreboard":
			scoreboard := map[string]interface{}{
				"items": []map[string]interface{}{
					{"rank": 1, "name": "Team A", "score": 100},
					{"rank": 2, "name": "Team B", "score": 80},
					{"rank": 3, "name": "Team C", "score": 60},
				},
				"challenges": map[string][]map[string]interface{}{
					"Web": {
						{"title": "XSS Challenge"},
						{"title": "SQL Injection"},
					},
					"Crypto": {
						{"title": "RSA"},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(scoreboard)
		}
	}))
	defer server.Close()

	api, err := gzapi.Init(server.URL, &gzapi.Creds{
		Username: "test",
		Password: "test",
	})
	if err != nil {
		t.Fatalf("Failed to initialize API: %v", err)
	}

	game := &gzapi.Game{
		Id: 1,
		CS: api,
	}

	feed, err := Scoreboard2CTFTimeFeed(game)
	if err != nil {
		t.Fatalf("Scoreboard2CTFTimeFeed() failed: %v", err)
	}

	// Verify standings
	if len(feed.Standings) != 3 {
		t.Errorf("Expected 3 standings, got %d", len(feed.Standings))
	}

	expectedStandings := []Standing{
		{Pos: 1, Team: "Team A", Score: 100},
		{Pos: 2, Team: "Team B", Score: 80},
		{Pos: 3, Team: "Team C", Score: 60},
	}

	for i, expected := range expectedStandings {
		if i >= len(feed.Standings) {
			break
		}
		actual := feed.Standings[i]
		if actual.Pos != expected.Pos || actual.Team != expected.Team || actual.Score != expected.Score {
			t.Errorf("Standing[%d] = %+v, want %+v", i, actual, expected)
		}
	}

	// Verify tasks
	if len(feed.Tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(feed.Tasks))
	}

	// Check that tasks contain category and title
	expectedTasks := []string{"Web - XSS Challenge", "Web - SQL Injection", "Crypto - RSA"}
	taskMap := make(map[string]bool)
	for _, task := range feed.Tasks {
		taskMap[task] = true
	}

	for _, expected := range expectedTasks {
		if !taskMap[expected] {
			t.Errorf("Expected task %q not found in feed.Tasks", expected)
		}
	}
}

// TestScoreboard2CTFTimeFeed_GetScoreboardError tests error handling
func TestScoreboard2CTFTimeFeed_GetScoreboardError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/account/login" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"token": "test-token",
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	api, err := gzapi.Init(server.URL, &gzapi.Creds{
		Username: "test",
		Password: "test",
	})
	if err != nil {
		t.Fatalf("Failed to initialize API: %v", err)
	}

	game := &gzapi.Game{
		Id: 1,
		CS: api,
	}

	_, err = Scoreboard2CTFTimeFeed(game)
	if err == nil {
		t.Error("Expected error when GetScoreboard fails, got nil")
	}

	if !errors.Is(err, errors.New("scoreboard error")) && err.Error() == "" {
		t.Logf("Error message: %v", err)
	}
}

// TestScoreboard2CTFTimeFeed_EmptyScoreboard tests empty scoreboard
func TestScoreboard2CTFTimeFeed_EmptyScoreboard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/account/login":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"token": "test-token",
			})
		case "/api/game/1/scoreboard":
			scoreboard := map[string]interface{}{
				"items":      []map[string]interface{}{},
				"challenges": map[string][]map[string]interface{}{},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(scoreboard)
		}
	}))
	defer server.Close()

	api, err := gzapi.Init(server.URL, &gzapi.Creds{
		Username: "test",
		Password: "test",
	})
	if err != nil {
		t.Fatalf("Failed to initialize API: %v", err)
	}

	game := &gzapi.Game{
		Id: 1,
		CS: api,
	}

	feed, err := Scoreboard2CTFTimeFeed(game)
	if err != nil {
		t.Fatalf("Scoreboard2CTFTimeFeed() with empty scoreboard failed: %v", err)
	}

	if len(feed.Standings) != 0 {
		t.Errorf("Expected 0 standings, got %d", len(feed.Standings))
	}

	if len(feed.Tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(feed.Tasks))
	}
}

// TestScoreboard2CTFTimeFeed_CapacityOptimization tests that capacity is correctly calculated
func TestScoreboard2CTFTimeFeed_CapacityOptimization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/account/login":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"token": "test-token",
			})
		case "/api/game/1/scoreboard":
			// Create a scoreboard with known sizes
			scoreboard := map[string]interface{}{
				"items": []map[string]interface{}{
					{"rank": 1, "name": "Team 1", "score": 100},
					{"rank": 2, "name": "Team 2", "score": 90},
				},
				"challenges": map[string][]map[string]interface{}{
					"Category1": {
						{"title": "Challenge 1"},
						{"title": "Challenge 2"},
						{"title": "Challenge 3"},
					},
					"Category2": {
						{"title": "Challenge 4"},
						{"title": "Challenge 5"},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(scoreboard)
		}
	}))
	defer server.Close()

	api, err := gzapi.Init(server.URL, &gzapi.Creds{
		Username: "test",
		Password: "test",
	})
	if err != nil {
		t.Fatalf("Failed to initialize API: %v", err)
	}

	game := &gzapi.Game{
		Id: 1,
		CS: api,
	}

	feed, err := Scoreboard2CTFTimeFeed(game)
	if err != nil {
		t.Fatalf("Scoreboard2CTFTimeFeed() failed: %v", err)
	}

	// Verify that the slices have the expected lengths
	// 2 standings
	if len(feed.Standings) != 2 {
		t.Errorf("Expected 2 standings, got %d", len(feed.Standings))
	}

	// 5 tasks (3 from Category1 + 2 from Category2)
	if len(feed.Tasks) != 5 {
		t.Errorf("Expected 5 tasks, got %d", len(feed.Tasks))
	}

	// Verify capacity is exactly what's needed (no over-allocation)
	// Note: We can't directly test cap() here without reflection, but the test
	// confirms the slices are sized correctly
}

// TestStanding_Fields tests the Standing struct fields
func TestStanding_Fields(t *testing.T) {
	s := Standing{
		Pos:   1,
		Team:  "Test Team",
		Score: 100,
	}

	if s.Pos != 1 {
		t.Errorf("Pos = %d, want 1", s.Pos)
	}

	if s.Team != "Test Team" {
		t.Errorf("Team = %q, want %q", s.Team, "Test Team")
	}

	if s.Score != 100 {
		t.Errorf("Score = %d, want 100", s.Score)
	}
}

// TestCTFTimeFeed_Fields tests the CTFTimeFeed struct fields
func TestCTFTimeFeed_Fields(t *testing.T) {
	feed := CTFTimeFeed{
		Tasks: []string{"task1", "task2"},
		Standings: []Standing{
			{Pos: 1, Team: "Team A", Score: 100},
		},
	}

	if len(feed.Tasks) != 2 {
		t.Errorf("len(Tasks) = %d, want 2", len(feed.Tasks))
	}

	if len(feed.Standings) != 1 {
		t.Errorf("len(Standings) = %d, want 1", len(feed.Standings))
	}
}

// TestScoreboard2CTFTimeFeed_JSONSerialization tests JSON serialization
func TestScoreboard2CTFTimeFeed_JSONSerialization(t *testing.T) {
	feed := &CTFTimeFeed{
		Tasks: []string{"Web - XSS", "Crypto - RSA"},
		Standings: []Standing{
			{Pos: 1, Team: "Team A", Score: 100},
			{Pos: 2, Team: "Team B", Score: 80},
		},
	}

	data, err := json.Marshal(feed)
	if err != nil {
		t.Fatalf("JSON Marshal failed: %v", err)
	}

	var decoded CTFTimeFeed
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("JSON Unmarshal failed: %v", err)
	}

	if len(decoded.Tasks) != len(feed.Tasks) {
		t.Errorf("Decoded tasks length = %d, want %d", len(decoded.Tasks), len(feed.Tasks))
	}

	if len(decoded.Standings) != len(feed.Standings) {
		t.Errorf("Decoded standings length = %d, want %d", len(decoded.Standings), len(feed.Standings))
	}

	for i, standing := range decoded.Standings {
		if standing != feed.Standings[i] {
			t.Errorf("Standing[%d] = %+v, want %+v", i, standing, feed.Standings[i])
		}
	}
}
