// Package event provides event and scoreboard management functionality
package event

import (
	"fmt"
	"sync"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// CTFTimeFeed represents a CTFTime-compatible feed format
type CTFTimeFeed struct {
	Tasks     []string   `json:"tasks"`
	Standings []Standing `json:"standings"`
}

// Standing represents a team's position in the scoreboard
type Standing struct {
	Pos   int    `json:"pos"`
	Team  string `json:"team"`
	Score int    `json:"score"`
}

// RemoveAllEvent removes all events/games
func RemoveAllEvent(api *gzapi.GZAPI) error {
	games, err := api.GetGames()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(games))
	sem := make(chan struct{}, 5) // Limit concurrent deletions

	for _, game := range games {
		wg.Add(1)
		sem <- struct{}{}
		go func(g gzapi.Game) {
			defer func() {
				<-sem
				wg.Done()
			}()
			if err := g.Delete(); err != nil {
				errChan <- err
			}
		}(*game)
	}

	wg.Wait()
	close(errChan)

	return <-errChan // Return first error if any
}

// Scoreboard2CTFTimeFeed converts scoreboard to CTFTime feed format
func Scoreboard2CTFTimeFeed(event *gzapi.Game) (*CTFTimeFeed, error) {
	scoreboard, err := event.GetScoreboard()
	if err != nil {
		return nil, fmt.Errorf("scoreboard error: %w", err)
	}

	feed := &CTFTimeFeed{
		Standings: make([]Standing, 0, len(scoreboard.Items)),
		Tasks:     make([]string, 0, len(scoreboard.Challenges)*5),
	}

	for _, item := range scoreboard.Items {
		feed.Standings = append(feed.Standings, Standing{
			Pos:   item.Rank,
			Team:  item.Name,
			Score: item.Score,
		})
	}

	for category, items := range scoreboard.Challenges {
		for _, item := range items {
			feed.Tasks = append(feed.Tasks, fmt.Sprintf("%s - %s", category, item.Title))
		}
	}
	return feed, nil
}
