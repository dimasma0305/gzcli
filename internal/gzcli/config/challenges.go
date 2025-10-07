//nolint:revive // Config constants and field names match project structure
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/utils"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	CHALLENGE_CATEGORY = []string{
		"Misc", "Crypto", "Pwn",
		"Web", "Reverse", "Blockchain",
		"Forensics", "Hardware", "Mobile", "PPC",
		"OSINT", "Game Hacking", "AI", "Pentest",
	}
	challengeFileRegex = regexp.MustCompile(`challenge\.(yaml|yml)$`)
	slugRegex          = regexp.MustCompile(`[^a-z0-9_]+`)
)

// Cache for parsed URL host
var hostCache struct {
	host string
	once sync.Once
}

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

// UnmarshalYAML implements custom YAML unmarshaling for ScriptValue
func (sv *ScriptValue) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// First try to unmarshal as a simple string
	var simpleScript string
	if err := unmarshal(&simpleScript); err == nil {
		sv.Simple = simpleScript
		sv.Complex = nil
		return nil
	}

	// If that fails, try to unmarshal as a complex object
	var complexScript ScriptConfig
	if err := unmarshal(&complexScript); err == nil {
		sv.Simple = ""
		sv.Complex = &complexScript
		return nil
	} else {
		return fmt.Errorf("script value must be either a string or an object with 'execute' and 'interval' fields")
	}
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

func generateSlug(challengeConf ChallengeYaml) string {
	var b strings.Builder
	b.Grow(len(challengeConf.Category) + len(challengeConf.Name) + 1)

	b.WriteString(strings.ToLower(challengeConf.Category))
	b.WriteByte('_')
	b.WriteString(strings.ToLower(challengeConf.Name))

	slug := strings.ReplaceAll(b.String(), " ", "_")
	return slugRegex.ReplaceAllString(slug, "")
}

func GetChallengesYaml(config *Config) ([]ChallengeYaml, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Pre-parse URL once
	hostCache.once.Do(func() {
		hostCache.host = config.Appsettings.ContainerProvider.PublicEntry
	})

	var wg sync.WaitGroup
	challengeChan := make(chan ChallengeYaml)
	errChan := make(chan error, 1)
	resultChan := make(chan []ChallengeYaml)

	// Start result collector
	go func() {
		var challenges []ChallengeYaml
		for c := range challengeChan {
			challenges = append(challenges, c)
		}
		resultChan <- challenges
	}()

	// Process categories in parallel
	for _, category := range CHALLENGE_CATEGORY {
		wg.Add(1)
		go func(category string) {
			defer wg.Done()
			categoryPath := filepath.Join(dir, category)

			if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
				return
			}

			err := filepath.Walk(categoryPath, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() || !challengeFileRegex.MatchString(info.Name()) {
					return err
				}

				//nolint:gosec // G304: File paths come from validated challenges directory
				content, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("reading file error: %w", err)
				}

				var challenge ChallengeYaml
				if err := utils.ParseYamlFromBytes(content, &challenge); err != nil {
					return fmt.Errorf("yaml parse error: %w %s", err, path)
				}

				challenge.Category = category
				challenge.Cwd = filepath.Dir(path)

				if category == "Game Hacking" {
					challenge.Category = "Reverse"
					challenge.Name = "[Game Hacking] " + challenge.Name
				}

				t, err := template.New("chall").Parse(string(content))
				if err != nil {
					log.ErrorH2("template error: %v", err)
					return nil
				}

				var buf bytes.Buffer
				err = t.Execute(&buf, map[string]string{
					"host": hostCache.host,
					"slug": generateSlug(challenge),
				})
				if err != nil {
					return fmt.Errorf("template execution error: %w", err)
				}

				if err := utils.ParseYamlFromBytes(buf.Bytes(), &challenge); err != nil {
					return fmt.Errorf("yaml parse error: %w %s", err, path)
				}

				select {
				case challengeChan <- challenge:
				case <-errChan:
				}
				return nil
			})

			if err != nil {
				select {
				case errChan <- fmt.Errorf("category %s: %w ", category, err):
				default:
				}
			}
		}(category)
	}

	go func() {
		wg.Wait()
		close(challengeChan)
	}()

	select {
	case err := <-errChan:
		close(errChan)
		return nil, err
	case challenges := <-resultChan:
		return challenges, nil
	}
}
