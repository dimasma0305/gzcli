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

func TestGenerateEmailBody_WithGlobalCommunicationLink(t *testing.T) {
	creds := &TeamCreds{
		Username:          "testuser",
		Password:          "testpassword",
		TeamName:          "Test Team",
		Email:             "test@example.com",
		CommunicationType: "Discord",
		CommunicationLink: "discord.gg/team-chat",
	}

	body := GenerateEmailBody("Test User", "http://example.com", creds, false)

	if !strings.Contains(body, "Discord") {
		t.Errorf("Email body does not contain communication type")
	}
	if !strings.Contains(body, "discord.gg/team-chat") {
		t.Errorf("Email body does not contain communication link")
	}
	if !strings.Contains(body, "https://discord.gg/team-chat") {
		t.Errorf("Email body does not contain normalized communication URL")
	}
}

func TestGenerateEmailBody_AutoDetectCommunicationType(t *testing.T) {
	tests := []struct {
		name         string
		link         string
		expectedType string
	}{
		{
			name:         "detect discord",
			link:         "https://discord.gg/team-chat",
			expectedType: "Discord",
		},
		{
			name:         "detect whatsapp",
			link:         "https://chat.whatsapp.com/ABC123",
			expectedType: "WhatsApp",
		},
		{
			name:         "detect slack",
			link:         "https://workspace.slack.com/archives/C123",
			expectedType: "Slack",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := &TeamCreds{
				Username:          "testuser",
				Password:          "testpassword",
				TeamName:          "Test Team",
				Email:             "test@example.com",
				CommunicationType: "",
				CommunicationLink: tt.link,
			}

			body := GenerateEmailBody("Test User", "http://example.com", creds, false)
			if !strings.Contains(body, tt.expectedType) {
				t.Errorf("Expected auto-detected communication type %q in email body", tt.expectedType)
			}
		})
	}
}
