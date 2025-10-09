// Package event provides event and scoreboard management functionality.
//
// This package handles CTF event operations including:
//   - Removing all events/games from the platform
//   - Converting scoreboards to CTFTime-compatible feed format
//
// Example usage:
//
//	api := gzapi.New("https://ctf.example.com")
//
//	// Remove all events
//	if err := event.RemoveAllEvent(api); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Convert scoreboard to CTFTime feed
//	feed, err := event.Scoreboard2CTFTimeFeed(game)
//	if err != nil {
//	    log.Fatal(err)
//	}
package event

import (
	"context"
	"fmt"
	"sync"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/service"
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

// RemoveAllEvent removes all events/games from the platform concurrently.
//
// This function fetches all games and deletes them in parallel using a worker pool
// of 5 concurrent goroutines to avoid overwhelming the API.
//
// Returns an error if fetching games fails or if any deletion fails.
// If multiple deletions fail, only the first error is returned.
//
// Example:
//
//	api := gzapi.New("https://ctf.example.com")
//	if err := event.RemoveAllEvent(api); err != nil {
//	    log.Fatalf("Failed to remove events: %v", err)
//	}
func RemoveAllEvent(api *gzapi.GZAPI) error {
	// Use service layer for game operations
	gameSvc := service.NewGameService(nil, nil, api)
	
	ctx := context.Background()
	games, err := gameSvc.GetGames(ctx)
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

	// Collect all errors instead of just the first one
	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("failed to delete game: %w", err)
		}
		return nil
	default:
		return nil
	}
}

// Scoreboard2CTFTimeFeed converts scoreboard to CTFTime feed format
func Scoreboard2CTFTimeFeed(event *gzapi.Game) (*CTFTimeFeed, error) {
	scoreboard, err := event.GetScoreboard()
	if err != nil {
		return nil, fmt.Errorf("scoreboard error: %w", err)
	}

	// Calculate exact capacity for tasks
	taskCount := 0
	for _, items := range scoreboard.Challenges {
		taskCount += len(items)
	}

	feed := &CTFTimeFeed{
		Standings: make([]Standing, 0, len(scoreboard.Items)),
		Tasks:     make([]string, 0, taskCount),
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
