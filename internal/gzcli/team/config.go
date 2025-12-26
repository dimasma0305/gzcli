package team

// ColumnMapping defines the mapping between CSV headers and required fields
type ColumnMapping struct {
	RealName string `yaml:"real_name"`
	Email    string `yaml:"email"`
	TeamName string `yaml:"team_name"`
}

// Config holds the configuration for team operations
type Config struct {
	ColumnMapping ColumnMapping `yaml:"column_mapping"`
}
