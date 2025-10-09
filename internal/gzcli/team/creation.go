// Package team provides team creation and management functionality
package team

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sethvargo/go-password/password"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/service"
	"github.com/dimasma0305/gzcli/internal/log"
)

// initializeCredentials initializes credentials for a new user
func initializeCredentials(teamCreds *TeamCreds, existingTeamNames, existingUserNames map[string]struct{}, credsCache []*TeamCreds, generateUsername func(string, int, map[string]struct{}) (string, error)) (*TeamCreds, error) {
	pass, err := password.Generate(24, 10, 0, false, false)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %v", err)
	}

	// Generate a unique username
	username, err := generateUsername(teamCreds.Username, 15, existingUserNames)
	if err != nil {
		return nil, fmt.Errorf("failed to generate username: %v", err)
	}

	// Normalize the team name
	const maxTeamNameLength = 20
	teamName := NormalizeTeamName(teamCreds.TeamName, maxTeamNameLength, existingTeamNames)

	// If registration fails, attempt to initialize API with cached credentials
	for _, creds := range credsCache {
		if creds.Email == teamCreds.Email {
			return creds, nil
		}
	}

	// Return new credentials
	currentCreds := teamCreds
	currentCreds.Username = username
	currentCreds.Password = pass
	currentCreds.TeamName = teamName

	return currentCreds, nil
}

// authenticateUser authenticates a user by logging in or registering
func authenticateUser(currentCreds *TeamCreds, config ConfigInterface, isExistingCreds bool) (*gzapi.GZAPI, error) {
	if isExistingCreds {
		api, err := gzapi.Init(config.GetUrl(), &gzapi.Creds{
			Username: currentCreds.Username,
			Password: currentCreds.Password,
		})
		if err != nil {
			log.Error("error login using: %v", currentCreds)
			return nil, err
		}
		return api, nil
	}

	api, err := gzapi.Register(config.GetUrl(), &gzapi.RegisterForm{
		Email:    currentCreds.Email,
		Username: currentCreds.Username,
		Password: currentCreds.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register: %v", err)
	}

	return api, nil
}

// ensureTeamCreated ensures a team is created for the user
func ensureTeamCreated(api *gzapi.GZAPI, currentCreds *TeamCreds, username, teamName string) {
	log.Info("Creating user %s with team %s", username, teamName)

	if currentCreds.IsTeamCreated {
		log.InfoH2("Team %s already created", teamName)
		return
	}

	// Use service layer for team creation
	teamSvc := service.NewTeamService(service.TeamServiceConfig{
		API: api,
	})
	
	serviceTeamCreds := &service.TeamCreds{
		TeamName: teamName,
		Username: teamCreds.Username,
		Password: teamCreds.Password,
		Email:    teamCreds.Email,
	}
	
	ctx := context.Background()
	err := teamSvc.CreateTeam(ctx, serviceTeamCreds)
	if err != nil {
		log.ErrorH2("Team %s already exist", teamName)
	}
	currentCreds.IsTeamCreated = true
}

// sendCredentialsEmail sends credentials via email if needed
func sendCredentialsEmail(teamCreds *TeamCreds, currentCreds *TeamCreds, config ConfigInterface, isSendEmail bool) {
	if !isSendEmail || currentCreds.IsEmailAlreadySent {
		log.ErrorH2("Email to %s already sended before", currentCreds.Email)
		return
	}

	environtURL := os.Getenv("URL")
	if environtURL == "" {
		environtURL = config.GetUrl()
	}

	if err := SendEmail(teamCreds.Username, environtURL, currentCreds, config.GetAppSettings()); err != nil {
		log.ErrorH2("Failed to send email to %s: %v", currentCreds.Email, err)
		return
	}

	log.InfoH2("Successfully sending email to %s", currentCreds.Email)
	currentCreds.IsEmailAlreadySent = true
}

// joinTeamToGame joins a team to the game
func joinTeamToGame(api *gzapi.GZAPI, config ConfigInterface) error {
	// Use service layer for team operations
	teamSvc := service.NewTeamService(service.TeamServiceConfig{
		API: api,
	})
	
	ctx := context.Background()
	// Note: GetTeams would need to be implemented in the service
	// For now, we'll use the direct API call
	team, err := api.GetTeams()
	if err != nil {
		log.Error("%s", err.Error())
		return err
	}

	if err := api.JoinGame(config.GetEventId(), &gzapi.GameJoinModel{
		TeamId:     team[0].Id,
		InviteCode: config.GetInviteCode(),
	}); err != nil {
		log.Error("%s", err.Error())
		return err
	}

	log.InfoH2("Successfully joining team %s to game %s", team[0].Name, config.GetEventTitle())
	return nil
}

// CreateTeamAndUser creates a team and user, ensuring the team name is unique and within the specified length.
func CreateTeamAndUser(teamCreds *TeamCreds, config ConfigInterface, existingTeamNames, existingUserNames map[string]struct{}, credsCache []*TeamCreds, isSendEmail bool, generateUsername func(string, int, map[string]struct{}) (string, error)) (*TeamCreds, error) {
	log.Info("Creating user %s with team %s", teamCreds.Username, teamCreds.TeamName)

	currentCreds, err := initializeCredentials(teamCreds, existingTeamNames, existingUserNames, credsCache, generateUsername)
	if err != nil {
		return nil, err
	}

	// Check if we found existing credentials in cache
	isExistingCreds := false
	for _, creds := range credsCache {
		if creds.Email == teamCreds.Email && creds == currentCreds {
			isExistingCreds = true
			break
		}
	}

	api, err := authenticateUser(currentCreds, config, isExistingCreds)
	if err != nil {
		return nil, err
	}

	ensureTeamCreated(api, currentCreds, currentCreds.Username, currentCreds.TeamName)
	sendCredentialsEmail(teamCreds, currentCreds, config, isSendEmail)

	if err := joinTeamToGame(api, config); err != nil {
		// Log error but don't fail the entire operation
		log.Error("Failed to join game: %v", err)
	}

	return currentCreds, nil
}

// NormalizeTeamName ensures team name is unique and within length limit
func NormalizeTeamName(teamName string, maxLen int, existingTeamNames map[string]struct{}) string {
	// Sanitize: remove null bytes and other problematic characters
	teamName = strings.ReplaceAll(teamName, "\x00", "")
	teamName = strings.ReplaceAll(teamName, "\n", " ")
	teamName = strings.ReplaceAll(teamName, "\r", " ")
	teamName = strings.ReplaceAll(teamName, "\t", " ")
	teamName = strings.TrimSpace(teamName)

	// Truncate if too long
	if len(teamName) > maxLen {
		teamName = teamName[:maxLen]
	}

	// Ensure uniqueness
	originalName := teamName
	counter := 1
	for {
		if _, exists := existingTeamNames[teamName]; !exists {
			existingTeamNames[teamName] = struct{}{}
			break
		}
		teamName = fmt.Sprintf("%s%d", originalName, counter)
		if len(teamName) > maxLen {
			// Adjust to fit counter within max length
			trimLen := maxLen - len(fmt.Sprintf("%d", counter))
			teamName = fmt.Sprintf("%s%d", originalName[:trimLen], counter)
		}
		counter++
	}

	return teamName
}

// ConfigInterface provides access to configuration values needed by team creation
type ConfigInterface interface {
	GetUrl() string
	GetEventId() int
	GetEventTitle() string
	GetInviteCode() string
	GetAppSettings() AppSettingsInterface
}

// AppSettingsInterface provides access to app settings
type AppSettingsInterface interface {
	GetEmailConfig() EmailConfig
}

// EmailConfig contains email configuration settings
type EmailConfig struct {
	UserName string
	Password string
	SMTP     struct {
		Host string
		Port int
	}
}
