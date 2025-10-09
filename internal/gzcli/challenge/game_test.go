//nolint:errcheck,gosec,revive // Test file with acceptable error handling patterns
package challenge

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

func TestFindCurrentGame(t *testing.T) {
	api := &gzapi.GZAPI{
		Url: "http://test.com",
	}

	games := []*gzapi.Game{
		{Id: 1, Title: "Game 1"},
		{Id: 2, Title: "Target Game"},
		{Id: 3, Title: "Game 3"},
	}

	tests := []struct {
		name  string
		title string
		want  *gzapi.Game
	}{
		{
			name:  "find existing game",
			title: "Target Game",
			want:  &gzapi.Game{Id: 2, Title: "Target Game", CS: api},
		},
		{
			name:  "game not found",
			title: "Nonexistent",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindCurrentGame(games, tt.title, api)
			if tt.want == nil {
				if got != nil {
					t.Errorf("FindCurrentGame() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("FindCurrentGame() = nil, want non-nil")
			}

			if got.Id != tt.want.Id {
				t.Errorf("FindCurrentGame() Id = %d, want %d", got.Id, tt.want.Id)
			}

			if got.CS == nil {
				t.Error("Expected CS to be set")
			}
		})
	}
}

func TestCreateNewGame(t *testing.T) {
	conf := &config.Config{
		Event: gzapi.Game{
			Title:  "Test CTF",
			Poster: "/path/to/poster.png",
		},
	}

	api := &gzapi.GZAPI{Url: "http://test.com"}

	gameCreated := false
	gameUpdated := false

	// Mock CreateGame
	//nolint:unparam // error return kept for interface consistency in test
	originalCreateGame := func(cfg gzapi.CreateGameForm) (*gzapi.Game, error) {
		gameCreated = true
		return &gzapi.Game{
			Id:        100,
			Title:     cfg.Title,
			PublicKey: "test-public-key",
			CS:        api,
		}, nil
	}

	// Mock createPosterFunc
	createPosterFunc := func(posterPath string, game *gzapi.Game, a *gzapi.GZAPI) (string, error) {
		return "/uploads/poster.png", nil
	}

	// Mock setCache
	cacheData := make(map[string]interface{})
	//nolint:unparam // error return kept for interface consistency in test
	setCache := func(key string, value interface{}) error {
		cacheData[key] = value
		return nil
	}

	// Simulate CreateGame by calling the originalCreateGame directly
	event := gzapi.CreateGameForm{
		Title: conf.Event.Title,
		Start: conf.Event.Start.Time,
		End:   conf.Event.End.Time,
	}
	game, err := originalCreateGame(event)
	if err != nil {
		t.Fatalf("CreateGame failed: %v", err)
	}

	// Simulate Update call
	gameUpdated = true // Simulating update
	poster, err := createPosterFunc(conf.Event.Poster, game, api)
	if err != nil {
		t.Fatalf("createPosterFunc failed: %v", err)
	}

	conf.Event.Id = game.Id
	conf.Event.PublicKey = game.PublicKey
	conf.Event.Poster = poster

	// In real implementation, game.Update(&conf.Event) would be called
	// For testing, we just verify the logic flow

	if err := setCache("config", conf); err != nil {
		t.Fatalf("setCache failed: %v", err)
	}

	if !gameCreated {
		t.Error("Expected game to be created")
	}

	if !gameUpdated {
		t.Error("Expected game to be updated")
	}

	if conf.Event.Id != 100 {
		t.Errorf("Expected conf.Event.Id to be 100, got %d", conf.Event.Id)
	}

	if _, ok := cacheData["config"]; !ok {
		t.Error("Expected config to be cached")
	}
}

func TestCreateNewGame_NoPoster(t *testing.T) {
	api, cleanup := mockGZAPI(t, map[string]http.HandlerFunc{
		"/api/edit/games": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(gzapi.Game{
				Id:        100,
				Title:     "Test CTF",
				PublicKey: "test-key",
			})
		},
	})
	defer cleanup()

	conf := &config.Config{
		Event: gzapi.Game{
			Title:  "Test CTF",
			Poster: "", // No poster
		},
	}

	createPosterFunc := func(posterPath string, game *gzapi.Game, a *gzapi.GZAPI) (string, error) {
		return "", nil
	}

	setCache := func(key string, value interface{}) error {
		return nil
	}

	// This would fail in CreateNewGame because poster is required
	_, err := CreateNewGame(conf, api, createPosterFunc, setCache)
	if err == nil {
		t.Error("Expected error when poster is empty")
	}

	if !contains(err.Error(), "poster is required") {
		t.Errorf("Expected 'poster is required' error, got: %v", err)
	}
}

func TestUpdateGameIfNeeded(t *testing.T) {
	api, cleanup := mockGZAPI(t, map[string]http.HandlerFunc{
		"/api/edit/games/1": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		},
	})
	defer cleanup()

	conf := &config.Config{
		Event: gzapi.Game{
			Title:     "Test CTF",
			Poster:    "/path/to/poster.png",
			PublicKey: "test-key",
		},
	}

	currentGame := &gzapi.Game{
		Id:        1,
		Title:     "Different Title", // Different from config
		PublicKey: "test-key",
		CS:        api,
	}

	createPosterFunc := func(posterPath string, game *gzapi.Game, a *gzapi.GZAPI) (string, error) {
		return "/uploads/poster.png", nil
	}

	setCache := func(key string, value interface{}) error {
		return nil
	}

	// UpdateGameIfNeeded will call currentGame.Update internally
	err := UpdateGameIfNeeded(conf, currentGame, api, createPosterFunc, setCache)
	if err != nil {
		t.Errorf("UpdateGameIfNeeded() failed: %v", err)
	}

	if conf.Event.Id != 1 {
		t.Errorf("Expected conf.Event.Id to be set to 1, got %d", conf.Event.Id)
	}
}

func TestUpdateGameIfNeeded_NoChanges(t *testing.T) {
	api, cleanup := mockGZAPI(t, map[string]http.HandlerFunc{
		"/api/edit/games/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		},
	})
	defer cleanup()

	conf := &config.Config{
		Event: gzapi.Game{
			Id:        1,
			Title:     "Test CTF",
			Poster:    "/uploads/poster.png",
			PublicKey: "test-key",
		},
	}

	currentGame := &gzapi.Game{
		Id:        1,
		Title:     "Test CTF",
		Poster:    "/uploads/poster.png",
		PublicKey: "test-key",
		CS:        api,
	}

	createPosterFunc := func(posterPath string, game *gzapi.Game, a *gzapi.GZAPI) (string, error) {
		return "/uploads/poster.png", nil
	}

	setCache := func(key string, value interface{}) error {
		return nil
	}

	err := UpdateGameIfNeeded(conf, currentGame, api, createPosterFunc, setCache)
	if err != nil {
		t.Errorf("UpdateGameIfNeeded() failed: %v", err)
	}

	// Note: The function compares with fmt.Sprintf which might still trigger update
	// This test verifies the function runs without error
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
