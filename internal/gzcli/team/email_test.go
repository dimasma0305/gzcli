package team

import (
	"strings"
	"testing"
)

func TestGenerateEmailBody(t *testing.T) {
	creds := &TeamCreds{
		Username: "testuser",
		Password: "testpassword",
		TeamName: "Test Team",
		Email:    "test@example.com",
	}
	realName := "Test User"
	website := "http://example.com"

	body := GenerateEmailBody(realName, website, creds)

	expectedWarning := "IMPORTANT: Do not change the account username and password."
	if !strings.Contains(body, expectedWarning) {
		t.Errorf("Email body does not contain warning message: %s", expectedWarning)
	}

	if !strings.Contains(body, creds.Username) {
		t.Errorf("Email body does not contain username: %s", creds.Username)
	}
	if !strings.Contains(body, creds.Password) {
		t.Errorf("Email body does not contain password: %s", creds.Password)
	}
	if !strings.Contains(body, creds.TeamName) {
		t.Errorf("Email body does not contain team name: %s", creds.TeamName)
	}
}
