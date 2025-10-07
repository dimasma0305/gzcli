package gzapi

import (
	"fmt"
)

// ScoreboardChallenge represents a challenge on the scoreboard
type ScoreboardChallenge struct {
	Score    int    `json:"score"`
	Category string `json:"category"`
	Title    string `json:"title"`
}

// ScoreboardItem represents a team's score and ranking
type ScoreboardItem struct {
	Name  string `json:"name"`
	Rank  int    `json:"rank"`
	Score int    `json:"score"`
}

// Scoreboard represents the game scoreboard with challenges and team rankings
type Scoreboard struct {
	Challenges map[string][]ScoreboardChallenge `json:"challenges"`
	Items      []ScoreboardItem                 `json:"items"`
}

// GetScoreboard retrieves the current scoreboard for the game
func (g *Game) GetScoreboard() (*Scoreboard, error) {
	var scoreboard Scoreboard
	err := g.CS.get(fmt.Sprintf("/api/game/%d/scoreboard", g.Id), &scoreboard)
	if err != nil {
		return nil, err
	}
	return &scoreboard, nil
}
