package team

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/dimasma0305/gzcli/internal/log"
)

// CommunicationOptions configures a single global communication contact
// applied to all generated team credentials from one CSV import.
type CommunicationOptions struct {
	Type string
	Link string
}

// GetData retrieves data from a URL or file path
func GetData(source string) ([]byte, error) {
	var output []byte
	var err error
	switch {
	case strings.HasPrefix(source, "http://"), strings.HasPrefix(source, "https://"):
		//nolint:gosec // G107: URL is validated and comes from user config
		resp, err := http.Get(source)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("failed to fetch data from URL")
		}

		output, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	case strings.HasPrefix(source, "file://"), !strings.Contains(source, "://"):
		filePath := strings.TrimPrefix(source, "file://")
		//nolint:gosec // G304: File path comes from user config
		output, err = os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported source prefix")
	}

	return output, nil
}

// ParseCSV parses CSV data and creates teams
func ParseCSV(data []byte, config ConfigInterface, teamConfig *Config, credsCache []*TeamCreds, isSendEmail bool, createTeamFunc func(*TeamCreds, ConfigInterface, map[string]struct{}, map[string]struct{}, []*TeamCreds, bool, func(string, int, map[string]struct{}) (string, error)) (*TeamCreds, error), generateUsername func(string, int, map[string]struct{}) (string, error), setCache func(string, interface{}) error, communicationOptions ...CommunicationOptions) error {
	reader := csv.NewReader(strings.NewReader(string(data)))

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV data: %v", err)
	}

	if len(records) == 0 {
		return errors.New("CSV is empty")
	}

	// Map to store the column indices for each header
	colIndices := make(map[string]int)

	// Assume the first row contains headers
	headers := records[0]
	for i, header := range headers {
		colIndices[strings.TrimSpace(header)] = i
	}

	// Validate required columns based on mapping
	requiredMappings := map[string]string{
		"RealName": teamConfig.ColumnMapping.RealName,
		"Email":    teamConfig.ColumnMapping.Email,
		"TeamName": teamConfig.ColumnMapping.TeamName,
	}

	for field, header := range requiredMappings {
		if _, ok := colIndices[header]; !ok {
			return fmt.Errorf("missing required header for %s: %s", field, header)
		}
	}

	// Maps for storing unique usernames and existing team names
	uniqueUsernames := make(map[string]struct{})
	existingTeamNames := make(map[string]struct{})

	// Create a map for quick lookup of existing credentials by email
	credsCacheMap := make(map[string]*TeamCreds)
	for _, creds := range credsCache {
		credsCacheMap[creds.Email] = creds
	}

	// List to hold the merged team credentials
	teamsCreds := make([]*TeamCreds, 0, len(records)-1)
	globalCommunication := CommunicationOptions{}
	if len(communicationOptions) > 0 {
		globalCommunication = communicationOptions[0]
	}

	for _, row := range records[1:] {
		realName := row[colIndices[teamConfig.ColumnMapping.RealName]]
		email := row[colIndices[teamConfig.ColumnMapping.Email]]
		teamName := row[colIndices[teamConfig.ColumnMapping.TeamName]]

		// Parse events if column mapping exists
		var events []string
		if teamConfig.ColumnMapping.Events != "" {
			if idx, ok := colIndices[teamConfig.ColumnMapping.Events]; ok {
				rawEvents := row[idx]
				if rawEvents != "" {
					// Split by comma and trim whitespace
					parts := strings.Split(rawEvents, ",")
					for _, part := range parts {
						if trimmed := strings.TrimSpace(part); trimmed != "" {
							events = append(events, trimmed)
						}
					}
				}
			}
		}

		// Create or update team and user based on the generated username
		creds, err := createTeamFunc(&TeamCreds{
			Username:          realName,
			Email:             email,
			TeamName:          teamName,
			CommunicationType: globalCommunication.Type,
			CommunicationLink: globalCommunication.Link,
			Events:            events,
		}, config, existingTeamNames, uniqueUsernames, credsCache, isSendEmail, generateUsername)
		if creds != nil {
			// Merge credentials if already exist in cache
			if existingCreds, exists := credsCacheMap[creds.Email]; exists {
				// Update the existing credentials with new information if necessary
				existingCreds.Username = creds.Username
				existingCreds.Password = creds.Password
				existingCreds.TeamName = creds.TeamName
				existingCreds.CommunicationType = creds.CommunicationType
				existingCreds.CommunicationLink = creds.CommunicationLink
			} else {
				// Add new credentials to the list
				teamsCreds = append(teamsCreds, creds)
			}
		}

		if err != nil {
			log.Error("%s", err.Error())
			continue
		}
	}

	// Add all credentials from the cache that were not updated
	for _, creds := range credsCacheMap {
		teamsCreds = append(teamsCreds, creds)
	}

	// Save the merged credentials to cache
	if err := setCache("teams_creds", teamsCreds); err != nil {
		return err
	}

	return nil
}
