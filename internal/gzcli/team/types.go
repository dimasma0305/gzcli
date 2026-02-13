package team

// TeamCreds stores team credentials
//
//nolint:revive // TeamCreds type name kept for package clarity and backward compatibility
type TeamCreds struct {
	Username           string   `json:"username" yaml:"username"`
	Password           string   `json:"password" yaml:"password"`
	Email              string   `json:"email" yaml:"email"`
	TeamName           string   `json:"team_name" yaml:"team_name"`
	CommunicationType  string   `json:"communication_type,omitempty" yaml:"communication_type,omitempty"`
	CommunicationLink  string   `json:"communication_link,omitempty" yaml:"communication_link,omitempty"`
	IsEmailAlreadySent bool     `json:"is_email_already_sent" yaml:"is_email_already_sent"`
	IsTeamCreated      bool     `json:"is_team_created" yaml:"is_team_created"`
	Events             []string `json:"events" yaml:"events"`
}
