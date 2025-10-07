package gzapi

import (
	"encoding/json"
	"fmt"
	"time"
)

// Game represents a CTF game/event in the GZCTF platform
//
//nolint:revive // Field names match API responses
type Game struct {
	Id                   int        `json:"id" yaml:"id"`
	Title                string     `json:"title" yaml:"title"`
	Hidden               bool       `json:"hidden" yaml:"hidden"`
	Summary              string     `json:"summary" yaml:"summary"`
	Content              string     `json:"content" yaml:"content"`
	AcceptWithoutReview  bool       `json:"acceptWithoutReview" yaml:"acceptWithoutReview"`
	WriteupRequired      bool       `json:"writeupRequired" yaml:"writeupRequired"`
	InviteCode           string     `json:"inviteCode,omitempty" yaml:"inviteCode,omitempty"`
	Organizations        []string   `json:"organizations,omitempty" yaml:"organizations,omitempty"`
	TeamMemberCountLimit int        `json:"teamMemberCountLimit" yaml:"teamMemberCountLimit"`
	ContainerCountLimit  int        `json:"containerCountLimit" yaml:"containerCountLimit"`
	Poster               string     `json:"poster,omitempty" yaml:"poster,omitempty"`
	PublicKey            string     `json:"publicKey" yaml:"publicKey"`
	PracticeMode         bool       `json:"practiceMode" yaml:"practiceMode"`
	Start                CustomTime `json:"start" yaml:"start"`
	End                  CustomTime `json:"end" yaml:"end"`
	WriteupDeadline      CustomTime `json:"writeupDeadline,omitempty" yaml:"writeupDeadline,omitempty"`
	WriteupNote          string     `json:"writeupNote" yaml:"writeupNote"`
	BloodBonus           int        `json:"bloodBonus" yaml:"bloodBonus"`
	CS                   *GZAPI     `json:"-" yaml:"-"`
}

// CustomTime wraps time.Time for custom JSON marshaling/unmarshaling
type CustomTime struct {
	time.Time
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	// The input comes as a number (milliseconds since epoch).
	var ms int64
	if err := json.Unmarshal(b, &ms); err != nil {
		// Try to parse as a string.
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return fmt.Errorf("invalid time format: %s", string(b))
		}
		// Parse the string as a time.
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return fmt.Errorf("invalid time format: %s", s)
		}
		ct.Time = t
		return nil
	}

	// Convert milliseconds to seconds and set the time.
	ct.Time = time.Unix(0, ms*int64(time.Millisecond))
	return nil
}

// GetGames retrieves all games from the GZCTF platform
func (cs *GZAPI) GetGames() ([]*Game, error) {
	var data struct {
		Data []*Game `json:"data"`
	}
	if err := cs.get("/api/edit/games?count=100&skip=0", &data); err != nil {
		return nil, err
	}
	for _, game := range data.Data {
		game.CS = cs
	}
	return data.Data, nil
}

// GetGameById retrieves a specific game by its ID
//
//nolint:revive // Method name matches existing API
func (cs *GZAPI) GetGameById(id int) (*Game, error) {
	var data *Game
	if err := cs.get(fmt.Sprintf("/api/edit/games/%d", id), &data); err != nil {
		return nil, err
	}
	data.CS = cs
	return data, nil
}

// GetGameByTitle retrieves a specific game by its title
func (cs *GZAPI) GetGameByTitle(title string) (*Game, error) {
	var games []*Game
	games, err := cs.GetGames()
	if err != nil {
		return nil, err
	}
	for _, game := range games {
		if game.Title == title {
			return game, nil
		}
	}
	return nil, fmt.Errorf("game not found")
}

// Delete removes the game from the platform
func (g *Game) Delete() error {
	return g.CS.delete(fmt.Sprintf("/api/edit/games/%d", g.Id), nil)
}

// Update updates the game configuration on the platform
func (g *Game) Update(game *Game) error {
	// Create a copy to avoid modifying the original
	gameCopy := *game

	// Convert all time fields to UTC to avoid PostgreSQL timezone issues
	gameCopy.Start.Time = gameCopy.Start.UTC()
	gameCopy.End.Time = gameCopy.End.UTC()
	if !gameCopy.WriteupDeadline.IsZero() {
		gameCopy.WriteupDeadline.Time = gameCopy.WriteupDeadline.UTC()
	}

	return g.CS.put(fmt.Sprintf("/api/edit/games/%d", g.Id), &gameCopy, nil)
}

// UploadPoster uploads a poster image for the game
func (g *Game) UploadPoster(poster string) (string, error) {
	var path string
	if err := g.CS.putMultiPart(fmt.Sprintf("/api/edit/games/%d/poster", g.Id), poster, &path); err != nil {
		return "", err
	}
	return path, nil
}

// CreateGameForm contains the data required to create a new game
type CreateGameForm struct {
	Title string    `json:"title"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// CreateGame creates a new game on the GZCTF platform
func (cs *GZAPI) CreateGame(game CreateGameForm) (*Game, error) {
	var data *Game
	game.Start = game.Start.UTC()
	game.End = game.End.UTC()
	if err := cs.post("/api/edit/games", game, &data); err != nil {
		return nil, err
	}
	data.CS = cs
	return data, nil
}

// GameJoinModel contains the data required for a team to join a game
//
//nolint:revive // Field and parameter names match API specification
type GameJoinModel struct {
	TeamId     int    `json:"teamId"`
	Division   string `json:"division,omitempty"`
	InviteCode string `json:"inviteCode,omitempty"`
}

// JoinGame joins a team to the game with optional division and invite code
//
//nolint:revive // Parameter name matches API specification
func (g *Game) JoinGame(teamId int, division string, inviteCode string) error {
	joinModel := &GameJoinModel{
		TeamId:     teamId,
		Division:   division,
		InviteCode: inviteCode,
	}
	return g.CS.post(fmt.Sprintf("/api/game/%d", g.Id), joinModel, nil)
}
