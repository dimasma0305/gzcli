package team

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

// GenerateEmailBody generates the HTML body for the email
func GenerateEmailBody(realName, website string, creds *TeamCreds) string {
	return fmt.Sprintf(`
	&nbsp;
	<html>
	<head>
		<style>
			body {
				font-family: Arial, sans-serif;
				line-height: 1.6;
				color: #333;
			}
			.block {
				max-width: 600px;
				margin: 0 auto;
				padding: 20px;
				border: 1px solid #eaeaea;
				border-radius: 5px;
				background-color: #f9f9f9;
			}
			h1 {
				color: #333;
			}
			.creds {
				margin-bottom: 20px;
			}
			.creds p {
				margin: 5px 0;
			}
			.cta {
				text-align: center;
				margin-top: 20px;
			}
			.cta a {
				display: inline-block;
				padding: 10px 20px;
				text-decoration: none;
				color: white;
				background-color: #007BFF;
				border-radius: 5px;
			}
			.cta a:hover {
				background-color: #0056b3;
			}
			.warning {
				color: #721c24;
				background-color: #f8d7da;
				border-color: #f5c6cb;
				padding: 10px;
				border-radius: 5px;
				margin-bottom: 20px;
				font-weight: bold;
			}
		</style>
	</head>
	<body>
		<div class="block">
		<h1>Hello %s,</h1>
		&nbsp;
		<div class="warning">
			IMPORTANT: Do not change the account username and password.
		</div>
		<div class="creds">
			<p>Here are your team credentials:</p>
			&nbsp;
			<p><strong>Username:</strong> %s</p>
			<p><strong>Password:</strong> %s</p>
			<p><strong>Team Name:</strong> %s</p>
			<p><strong>Website:</strong> <a href="%s">%s</a></p>
		</div>
		&nbsp;
		<p>After logging in with your credentials, you can copy your team invitation code from the /teams page, and then share it with your team members.</p>
		&nbsp;
		<p>Make sure to notify your team members to register first and then use the invitation code on the /team page.</p>
		&nbsp;
		<p>Once all your team members have joined, you can navigate to the /games page and request to join the game. The admin will verify your request, and you just need to wait for the CTF to start.</p>
		&nbsp;
		<div class="cta">
			<a href="%s">Go to Website</a>
		</div>
		&nbsp;
		</div>
	</body>
	</html>
	`,
		realName, creds.Username, creds.Password, creds.TeamName, website, website, website,
	)
}

// SendEmail sends the team credentials to the specified email address using gomail
func SendEmail(realName string, website string, creds *TeamCreds, appsettings AppSettingsInterface) error {
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

	htmlBody := GenerateEmailBody(realName, website, creds)

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
