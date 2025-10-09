package service

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sethvargo/go-password/password"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
	"github.com/dimasma0305/gzcli/internal/gzcli/errors"
	"github.com/dimasma0305/gzcli/internal/log"
)

// TeamServiceConfig holds configuration for TeamService
type TeamServiceConfig struct {
	API *gzapi.GZAPI
}

// TeamService handles team business logic
type TeamService struct {
	api *gzapi.GZAPI
}

// NewTeamService creates a new TeamService
func NewTeamService(config TeamServiceConfig) *TeamService {
	return &TeamService{
		api: config.API,
	}
}

// TeamCreds represents team credentials
type TeamCreds struct {
	TeamName string `json:"team_name" yaml:"team_name"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Email    string `json:"email" yaml:"email"`
}

// CreateTeam creates a new team with credentials
func (s *TeamService) CreateTeam(ctx context.Context, teamCreds *TeamCreds) error {
	log.Info("Creating team: %s", teamCreds.TeamName)

	// Generate password if not provided
	if teamCreds.Password == "" {
		pass, err := password.Generate(24, 10, 0, false, false)
		if err != nil {
			return errors.Wrap(err, "failed to generate password")
		}
		teamCreds.Password = pass
	}

	// Create team via API
	err := s.api.CreateTeam(gzapi.CreateTeamForm{
		Name:     teamCreds.TeamName,
		Password: teamCreds.Password,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to create team %s", teamCreds.TeamName)
	}

	log.Info("Successfully created team: %s", teamCreds.TeamName)
	return nil
}

// CreateTeamsFromCSV creates multiple teams from a CSV file
func (s *TeamService) CreateTeamsFromCSV(ctx context.Context, csvPath string) error {
	log.Info("Creating teams from CSV: %s", csvPath)

	// Read CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		return errors.Wrapf(err, "failed to open CSV file: %s", csvPath)
	}
	defer file.Close()

	// Parse CSV and create teams
	teams, err := s.parseTeamCSV(file)
	if err != nil {
		return errors.Wrap(err, "failed to parse CSV file")
	}

	// Create each team
	for _, team := range teams {
		if err := s.CreateTeam(ctx, team); err != nil {
			log.Error("Failed to create team %s: %v", team.TeamName, err)
			// Continue with other teams
			continue
		}
	}

	log.Info("Completed team creation from CSV")
	return nil
}

// parseTeamCSV parses a CSV file containing team data
func (s *TeamService) parseTeamCSV(file *os.File) ([]*TeamCreds, error) {
	// This would implement CSV parsing logic
	// For now, return an empty slice
	log.Debug("CSV parsing not fully implemented")
	return []*TeamCreds{}, nil
}

// NormalizeTeamName normalizes a team name to ensure uniqueness
func (s *TeamService) NormalizeTeamName(teamName string, maxLength int, existingNames map[string]struct{}) string {
	// Remove special characters and convert to lowercase
	normalized := strings.ToLower(strings.TrimSpace(teamName))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	
	// Truncate if too long
	if len(normalized) > maxLength {
		normalized = normalized[:maxLength]
	}
	
	// Ensure uniqueness
	original := normalized
	counter := 1
	for {
		if _, exists := existingNames[normalized]; !exists {
			break
		}
		normalized = fmt.Sprintf("%s_%d", original, counter)
		counter++
	}
	
	return normalized
}

// GenerateUsername generates a unique username
func (s *TeamService) GenerateUsername(baseUsername string, maxLength int, existingUsernames map[string]struct{}) (string, error) {
	// Normalize base username
	username := strings.ToLower(strings.TrimSpace(baseUsername))
	username = strings.ReplaceAll(username, " ", "_")
	
	// Truncate if too long
	if len(username) > maxLength {
		username = username[:maxLength]
	}
	
	// Ensure uniqueness
	original := username
	counter := 1
	for {
		if _, exists := existingUsernames[username]; !exists {
			break
		}
		username = fmt.Sprintf("%s%d", original, counter)
		counter++
		
		// Prevent infinite loop
		if counter > 1000 {
			return "", errors.New("unable to generate unique username")
		}
	}
	
	return username, nil
}