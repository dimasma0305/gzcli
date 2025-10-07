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

// GetData retrieves data from a URL or file path
func GetData(source string) ([]byte, error) {
	var output []byte
	var err error
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
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
	} else if strings.HasPrefix(source, "file://") || !strings.Contains(source, "://") {
		filePath := strings.TrimPrefix(source, "file://")
		output, err = os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("unsupported source prefix")
	}

	return output, nil
}

// ParseCSV parses CSV data and creates teams
func ParseCSV(data []byte, config ConfigInterface, credsCache []*TeamCreds, isSendEmail bool, createTeamFunc func(*TeamCreds, ConfigInterface, map[string]struct{}, map[string]struct{}, []*TeamCreds, bool, func(string, int, map[string]struct{}) (string, error)) (*TeamCreds, error), generateUsername func(string, int, map[string]struct{}) (string, error), setCache func(string, interface{}) error) error {
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
		colIndices[header] = i
	}

	// Ensure that the required headers are present
	requiredHeaders := []string{"RealName", "Email", "TeamName"}
	for _, header := range requiredHeaders {
		if _, ok := colIndices[header]; !ok {
			return errors.New("missing required header: " + header)
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

	for _, row := range records[1:] {
		realName := row[colIndices["RealName"]]
		email := row[colIndices["Email"]]
		teamName := row[colIndices["TeamName"]]

		// Create or update team and user based on the generated username
		creds, err := createTeamFunc(&TeamCreds{
			Username: realName,
			Email:    email,
			TeamName: teamName,
		}, config, existingTeamNames, uniqueUsernames, credsCache, isSendEmail, generateUsername)
		if err != nil {
			log.Error("%s", err.Error())
			continue
		}

		if creds != nil {
			// Merge credentials if already exist in cache
			if existingCreds, exists := credsCacheMap[creds.Email]; exists {
				// Update the existing credentials with new information if necessary
				existingCreds.Username = creds.Username
				existingCreds.Password = creds.Password
				existingCreds.TeamName = creds.TeamName
			} else {
				// Add new credentials to the list
				teamsCreds = append(teamsCreds, creds)
			}
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
