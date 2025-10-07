package team

// TeamCreds stores team credentials
type TeamCreds struct {
	Username           string `json:"username" yaml:"username"`
	Password           string `json:"password" yaml:"password"`
	Email              string `json:"email" yaml:"email"`
	TeamName           string `json:"team_name" yaml:"team_name"`
	IsEmailAlreadySent bool   `json:"is_email_already_sent" yaml:"is_email_already_sent"`
	IsTeamCreated      bool   `json:"is_team_created" yaml:"is_team_created"`
}
