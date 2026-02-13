package team

import (
	"fmt"
	"strings"

	"gopkg.in/gomail.v2"
)

// DetectCommunicationType infers the platform name from a communication link.
func DetectCommunicationType(link string) string {
	lower := strings.ToLower(strings.TrimSpace(link))
	if lower == "" {
		return ""
	}

	switch {
	case strings.Contains(lower, "wa.me/"),
		strings.Contains(lower, "chat.whatsapp.com/"),
		strings.Contains(lower, "whatsapp.com/"):
		return "WhatsApp"
	case strings.Contains(lower, "discord.gg/"),
		strings.Contains(lower, "discord.com/"),
		strings.Contains(lower, "discordapp.com/"):
		return "Discord"
	case strings.Contains(lower, "slack.com/"),
		strings.Contains(lower, ".slack.com/"):
		return "Slack"
	default:
		return ""
	}
}

// GenerateEmailBody generates the HTML body for the email
func GenerateEmailBody(realName, website string, creds *TeamCreds, isSolo bool) string {
	modeLabel := "Team CTF"
	modeInstructions := `
		<p>After logging in, open the <strong>/teams</strong> page to copy your team invitation code.</p>
		<p>Ask teammates to register first, then join from the <strong>/team</strong> page using that code.</p>
		<p>Your team has already been joined to the event automatically. Go to <strong>/games</strong> to verify status and prepare.</p>
	`

	if isSolo {
		modeLabel = "Solo CTF"
		modeInstructions = `
		<p>This event is configured as <strong>Solo CTF</strong>, so no team invitation code is required.</p>
		<p>Your account has already been joined to the event automatically. Go to <strong>/games</strong> to verify status and prepare.</p>
	`
	}

	communicationSection := ""
	communicationType := strings.TrimSpace(creds.CommunicationType)
	communicationLink := strings.TrimSpace(creds.CommunicationLink)
	if communicationLink != "" {
		if communicationType == "" {
			communicationType = DetectCommunicationType(communicationLink)
		}
		if communicationType == "" {
			communicationType = "Communication"
		}

		normalizedLink := communicationLink
		if !strings.Contains(normalizedLink, "://") {
			normalizedLink = "https://" + normalizedLink
		}

		communicationSection = fmt.Sprintf(
			`<p><strong>%s:</strong> <a href="%s">%s</a></p>`,
			communicationType,
			normalizedLink,
			communicationLink,
		)
	}

	return fmt.Sprintf(`
	<html>
	<head>
		<style>
			body {
				font-family: Arial, sans-serif;
				line-height: 1.6;
				color: #1f2937;
				background-color: #f3f4f6;
				margin: 0;
				padding: 24px 12px;
			}
			.block {
				max-width: 600px;
				margin: 0 auto;
				padding: 24px;
				border: 1px solid #e5e7eb;
				border-radius: 10px;
				background-color: #ffffff;
				box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
			}
			h1 {
				color: #111827;
				margin-top: 0;
				margin-bottom: 8px;
			}
			.subtitle {
				margin-top: 0;
				margin-bottom: 20px;
				color: #4b5563;
			}
			.mode {
				display: inline-block;
				margin-bottom: 16px;
				padding: 6px 10px;
				border-radius: 999px;
				background-color: #ecfeff;
				color: #0f766e;
				font-size: 13px;
				font-weight: 700;
			}
			.creds {
				margin-bottom: 18px;
				padding: 12px;
				border: 1px solid #e5e7eb;
				border-radius: 8px;
				background-color: #f9fafb;
			}
			.creds p {
				margin: 5px 0;
			}
			.steps p {
				margin: 10px 0;
			}
			.cta {
				text-align: center;
				margin-top: 24px;
			}
			.cta a {
				display: inline-block;
				padding: 10px 18px;
				text-decoration: none;
				color: white;
				background-color: #2563eb;
				border-radius: 5px;
				font-weight: 600;
			}
			.cta a:hover {
				background-color: #1d4ed8;
			}
		</style>
	</head>
	<body>
		<div class="block">
		<h1>Hello %s,</h1>
		<p class="subtitle">Your account has been created successfully.</p>
		<div class="mode">%s</div>
		<div class="creds">
			<p><strong>Credentials</strong></p>
			<p><strong>Username:</strong> %s</p>
			<p><strong>Password:</strong> %s</p>
			<p><strong>Team Name:</strong> %s</p>
			<p><strong>Website:</strong> <a href="%s">%s</a></p>
			%s
		</div>
		<div class="steps">
			%s
		</div>
		<p>If anything looks wrong, reply to this email so we can help quickly.</p>
		<div class="cta">
			<a href="%s">Go to Website</a>
		</div>
		</div>
	</body>
	</html>
	`,
		realName, modeLabel, creds.Username, creds.Password, creds.TeamName, website, website, communicationSection, modeInstructions, website,
	)
}

// SendEmail sends the team credentials to the specified email address using gomail
func SendEmail(realName string, website string, creds *TeamCreds, appsettings AppSettingsInterface, isSolo bool) error {
	emailConfig := appsettings.GetEmailConfig()
	smtp := emailConfig.SMTP

	// Extract the necessary fields from the emailConfig map
	smtpHost := smtp.Host
	smtpPort := smtp.Port
	smtpUsername := emailConfig.UserName
	smtpPassword := emailConfig.Password

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUsername)
	m.SetHeader("To", creds.Email)
	m.SetHeader("Subject", "Your Team Credentials")

	htmlBody := GenerateEmailBody(realName, website, creds, isSolo)

	// Set the email body as HTML
	m.SetBody("text/html", htmlBody)

	// Dial the SMTP server
	d := gomail.NewDialer(smtpHost, smtpPort, smtpUsername, smtpPassword)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
