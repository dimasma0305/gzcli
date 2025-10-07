// Package team provides team creation and management functionality
package team

import (
	"fmt"
	"os"
	"strings"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/log"
	"github.com/sethvargo/go-password/password"
)

// CreateTeamAndUser creates a team and user, ensuring the team name is unique and within the specified length.
func CreateTeamAndUser(teamCreds *TeamCreds, config ConfigInterface, existingTeamNames, existingUserNames map[string]struct{}, credsCache []*TeamCreds, isSendEmail bool, generateUsername func(string, int, map[string]struct{}) (string, error)) (*TeamCreds, error) {
	log.Info("Creating user %s with team %s", teamCreds.Username, teamCreds.TeamName)
	var api *gzapi.GZAPI
	var currentCreds *TeamCreds
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

	alreadyLogin := false

	// If registration fails, attempt to initialize API with cached credentials
	for _, creds := range credsCache {
		if creds.Email == teamCreds.Email {
			currentCreds = creds
		}
	}

	if currentCreds != nil {
		api, err = gzapi.Init(config.GetUrl(), &gzapi.Creds{
			Username: currentCreds.Username,
			Password: currentCreds.Password,
		})
		if err == nil {
			alreadyLogin = true
		} else {
			log.Error("error login using: %v", currentCreds)
			return nil, err
		}

	} else {
		currentCreds = teamCreds
		currentCreds.Username = username
		currentCreds.Password = pass
		currentCreds.TeamName = teamName
	}

	if !alreadyLogin {
		api, err = gzapi.Register(config.GetUrl(), &gzapi.RegisterForm{
			Email:    currentCreds.Email,
			Username: currentCreds.Username,
			Password: currentCreds.Password,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to register: %v", err)
		}
	}

	// Create the team
	log.Info("Creating user %s with team %s", username, teamName)
	if !currentCreds.IsTeamCreated {
		err = api.CreateTeam(&gzapi.TeamForm{
			Bio:  "",
			Name: teamName,
		})
		if err != nil {
			log.ErrorH2("Team %s already exist", teamName)
		}
	} else {
		log.InfoH2("Team %s already created", teamName)
	}
	currentCreds.IsTeamCreated = true

	// get environt URL
	environtURL := os.Getenv("URL")
	if environtURL == "" {
		environtURL = config.GetUrl()
	}

	// Send credentials via email if enabled in the config
	if isSendEmail && !currentCreds.IsEmailAlreadySent {
		if err := SendEmail(teamCreds.Username, environtURL, currentCreds, config.GetAppSettings()); err != nil {
			log.ErrorH2("Failed to send email to %s: %v", currentCreds.Email, err)
		}
		log.InfoH2("Successfully sending email to %s", currentCreds.Email)
		currentCreds.IsEmailAlreadySent = true
	} else {
		log.ErrorH2("Email to %s already sended before", currentCreds.Email)
	}

	// get team info
	team, err := api.GetTeams()
	if err != nil {
		log.Error("%s", err.Error())
	}

	if err := api.JoinGame(config.GetEventId(), &gzapi.GameJoinModel{
		TeamId:     team[0].Id,
		InviteCode: config.GetInviteCode(),
	}); err != nil {
		log.Error("%s", err.Error())
	}
	log.InfoH2("Successfully joining team %s to game %s", team[0].Name, config.GetEventTitle())

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
