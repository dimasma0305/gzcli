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

	body := GenerateEmailBody(realName, website, creds, false)

	if !strings.Contains(body, creds.Username) {
		t.Errorf("Email body does not contain username: %s", creds.Username)
	}
	if !strings.Contains(body, creds.Password) {
		t.Errorf("Email body does not contain password: %s", creds.Password)
	}
	if !strings.Contains(body, creds.TeamName) {
		t.Errorf("Email body does not contain team name: %s", creds.TeamName)
	}

	if !strings.Contains(body, "Team CTF") {
		t.Errorf("Email body does not contain Team CTF mode label")
	}
}

func TestGenerateEmailBody_SoloMode(t *testing.T) {
	creds := &TeamCreds{
		Username: "solo-user",
		Password: "solo-pass",
		TeamName: "Solo Team",
		Email:    "solo@example.com",
	}
	body := GenerateEmailBody("Solo User", "http://example.com", creds, true)

	if !strings.Contains(body, "Solo CTF") {
		t.Errorf("Email body does not contain Solo CTF mode label")
	}
	if !strings.Contains(body, "no team invitation code is required") {
		t.Errorf("Email body does not contain solo-specific instructions")
	}
	if strings.Contains(body, "copy your team invitation code") {
		t.Errorf("Email body should not contain team invitation instructions in solo mode")
	}
}
