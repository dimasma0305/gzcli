package challenge

import (
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// ChallengeYaml represents a challenge configuration from YAML
type ChallengeYaml struct {
	Name        string                 `yaml:"name"`
	Author      string                 `yaml:"author"`
	Description string                 `yaml:"description"`
	Flags       []string               `yaml:"flags"`
	Value       int                    `yaml:"value"`
	Provide     *string                `yaml:"provide,omitempty"`
	Visible     *bool                  `yaml:"visible"`
	Type        string                 `yaml:"type"`
	Hints       []string               `yaml:"hints"`
	Container   Container              `yaml:"container"`
	Scripts     map[string]ScriptValue `yaml:"scripts"`
	Dashboard   *Dashboard             `yaml:"dashboard,omitempty"`
	Category    string                 `yaml:"-"`
	Cwd         string                 `yaml:"-"`
}

// Container represents container configuration
type Container struct {
	FlagTemplate         string `yaml:"flagTemplate"`
	ContainerImage       string `yaml:"containerImage"`
	MemoryLimit          int    `yaml:"memoryLimit"`
	CpuCount             int    `yaml:"cpuCount"`
	StorageLimit         int    `yaml:"storageLimit"`
	ContainerExposePort  int    `yaml:"containerExposePort"`
	EnableTrafficCapture bool   `yaml:"enableTrafficCapture"`
}

// ScriptConfig represents a script configuration with interval and execute parameters
type ScriptConfig struct {
	Execute  string        `yaml:"execute,omitempty"`
	Interval time.Duration `yaml:"interval,omitempty"`
}

// ScriptValue holds either a simple command string or a complex ScriptConfig
type ScriptValue struct {
	Simple  string
	Complex *ScriptConfig
}

// IsSimple returns true if this is a simple string command
func (sv *ScriptValue) IsSimple() bool {
	return sv.Simple != ""
}

// GetCommand returns the command to execute
func (sv *ScriptValue) GetCommand() string {
	if sv.IsSimple() {
		return sv.Simple
	}
	if sv.Complex != nil {
		return sv.Complex.Execute
	}
	return ""
}

// GetInterval returns the execution interval for complex scripts
func (sv *ScriptValue) GetInterval() time.Duration {
	if sv.Complex != nil {
		return sv.Complex.Interval
	}
	return 0
}

// HasInterval returns true if this script has an interval configured
func (sv *ScriptValue) HasInterval() bool {
	return sv.Complex != nil && sv.Complex.Interval > 0
}

// Dashboard represents dashboard configuration
type Dashboard struct {
	Compose                  string `yaml:"compose"`
	ChallengeDurationMinutes int    `yaml:"challengeDurationMinutes"`
	ResetTimerMinutes        int    `yaml:"resetTimerMinutes"`
	RestartCooldownMinutes   int    `yaml:"restartCooldownMinutes"`
}

// Config represents the application configuration
type Config struct {
	Url   string      `yaml:"url"`
	Creds gzapi.Creds `yaml:"creds"`
	Event gzapi.Game  `yaml:"event"`
}

// AppSettings represents application settings
type AppSettings struct {
	ContainerProvider struct {
		PublicEntry string `json:"PublicEntry"`
	} `json:"ContainerProvider"`
	EmailConfig struct {
		UserName string `json:"UserName"`
		Password string `json:"Password"`
		Smtp     struct {
			Host string `json:"Host"`
			Port int    `json:"Port"`
		} `json:"Smtp"`
	} `json:"EmailConfig"`
}
