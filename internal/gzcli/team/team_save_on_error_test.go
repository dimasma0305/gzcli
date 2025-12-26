package team

import (
	"fmt"
	"testing"
)

func TestParseCSV_SaveCredsOnError(t *testing.T) {
	csvData := []byte(`RealName,Email,TeamName
Error User,error@example.com,Error Team`)

	config := &mockConfig{
		url:     "http://test.com",
		eventId: 1,
	}

	credsSaved := false
	setCache := func(_ string, value interface{}) error {
		credsList, ok := value.([]*TeamCreds)
		if !ok {
			return nil
		}
		for _, c := range credsList {
			if c.Email == "error@example.com" {
				credsSaved = true
			}
		}
		return nil
	}

	// Function that returns creds AND an error
	createTeamFunc := func(creds *TeamCreds, _ ConfigInterface, _, _ map[string]struct{}, _ []*TeamCreds, _ bool, _ func(string, int, map[string]struct{}) (string, error)) (*TeamCreds, error) {
		return &TeamCreds{
			Username: creds.Username,
			Email:    creds.Email,
			TeamName: creds.TeamName,
			Password: "generated_password",
		}, fmt.Errorf("intentional failure")
	}

	generateUsername := func(name string, _ int, _ map[string]struct{}) (string, error) {
		return name, nil
	}

	// Should not return error, just log it. But ParseCSV returns nil if iteration finishes.
	// We want to check if setCache was called with the creds.
	err := ParseCSV(
		csvData,
		config,
		&Config{ColumnMapping: ColumnMapping{RealName: "RealName", Email: "Email", TeamName: "TeamName"}},
		[]*TeamCreds{},
		false,
		createTeamFunc,
		generateUsername,
		setCache,
	)

	if err != nil {
		t.Errorf("ParseCSV() returned error: %v", err)
	}

	if !credsSaved {
		t.Error("ParseCSV() did not save credentials for user that encountered an error")
	}
}
