package uploadserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
)

// ValidationError represents a detailed validation error
type ValidationError struct {
	What     string
	Where    string
	HowToFix string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("Validation Error:\n  What: %s\n  Where: %s\n  How to Fix: %s", e.What, e.Where, e.HowToFix)
}

func validateUploadChallenge(root string, chall config.ChallengeYaml) error {
	if err := ensureDashboardConfigExists(root, chall); err != nil {
		return err
	}

	if chall.Dashboard != nil {
		if err := validateDockerBuildResources(root); err != nil {
			return err
		}
	}

	if err := validateContainerSource(root, chall); err != nil {
		return err
	}

	if err := validateExposedPort(root, chall); err != nil {
		return err
	}

	if err := validateDockerComposePrivileged(root); err != nil {
		return err
	}

	if err := validateChallengeScripts(chall); err != nil {
		return err
	}

	return nil
}

func ensureDashboardConfigExists(root string, chall config.ChallengeYaml) error {
	if chall.Dashboard == nil || chall.Dashboard.Config == "" {
		return nil
	}

	if chall.Dashboard.Config != "./src/docker-compose.yml" {
		return &ValidationError{
			What:     "Invalid dashboard config path",
			Where:    "challenge.yml (dashboard.config)",
			HowToFix: "Dashboard config must be set to './src/docker-compose.yml'",
		}
	}

	configPath := filepath.Join(root, chall.Dashboard.Config)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ValidationError{
			What:     fmt.Sprintf("Dashboard config file not found: %s", chall.Dashboard.Config),
			Where:    "challenge.yml (dashboard.config)",
			HowToFix: "Ensure the file specified in 'dashboard.config' exists in the challenge archive.",
		}
	}
	return nil
}

func validateContainerSource(root string, chall config.ChallengeYaml) error {
	switch chall.Type {
	case "DynamicContainer":
		dcPath := filepath.Join(root, "docker-compose.yml")
		if _, err := os.Stat(dcPath); os.IsNotExist(err) {
			return &ValidationError{
				What:     "Missing docker-compose.yml",
				Where:    "Challenge Root",
				HowToFix: "DynamicContainer challenges require a docker-compose.yml file.",
			}
		}
	case "StaticContainer":
		// For StaticContainer, if no image is provided or it looks local, check for Dockerfile
		// Simpler check: If provide is missing or implies build, we might expect Dockerfile.
		// Use a basic check: if Dockerfile exists, good. If not, and image is local-like, warn/error?
		// Given standard usage, let's checking if Dockerfile exists in root if not using a clear remote image.
		if !strings.Contains(chall.Container.ContainerImage, "/") {
			dfPath := filepath.Join(root, "Dockerfile")
			if _, err := os.Stat(dfPath); os.IsNotExist(err) {
				// Also check src/Dockerfile as per some templates
				srcDfPath := filepath.Join(root, "src/Dockerfile")
				if _, err := os.Stat(srcDfPath); os.IsNotExist(err) {
					// This is heuristic, so maybe skip strict error for now to avoid false positives,
					// but user asked for checks. Let's return error if we are fairly sure.
					// If containerImage is just a name (no dots, no slashes), it's likely a local build tag.
					// e.g. "my-challenge".
					return &ValidationError{
						What:     "Missing Dockerfile for StaticContainer",
						Where:    "Challenge Root or src/",
						HowToFix: "Include a Dockerfile to build your container, or specify a full image name (e.g. registry/image:tag).",
					}
				}
			}
		}
	}
	return nil
}

func validateExposedPort(root string, chall config.ChallengeYaml) error {
	if chall.Container.ContainerExposePort == 0 || chall.Type != "DynamicContainer" {
		return nil
	}

	dcPath := filepath.Join(root, "docker-compose.yml")
	if _, err := os.Stat(dcPath); os.IsNotExist(err) {
		return nil // Handled by validateContainerSource
	}

	//nolint:gosec // Validating user-provided challenge directory
	content, err := os.ReadFile(dcPath)
	if err != nil {
		return fmt.Errorf("failed to read docker-compose.yml: %w", err)
	}

	var compose struct {
		Services map[string]struct {
			Ports  []interface{} `yaml:"ports"`
			Expose []interface{} `yaml:"expose"`
		} `yaml:"services"`
	}

	if err := yaml.Unmarshal(content, &compose); err != nil {
		return fmt.Errorf("failed to parse docker-compose.yml: %w", err)
	}

	targetPort := fmt.Sprintf("%d", chall.Container.ContainerExposePort)
	found := false

	for _, service := range compose.Services {
		for _, p := range service.Ports {
			str := fmt.Sprintf("%v", p)
			if strings.Contains(str, ":"+targetPort) || strings.HasSuffix(str, targetPort) || str == targetPort {
				found = true
				break
			}
		}
		if found {
			break
		}
		for _, e := range service.Expose {
			str := fmt.Sprintf("%v", e)
			if str == targetPort {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return &ValidationError{
			What:     fmt.Sprintf("Exposed port %d not found in docker-compose.yml", chall.Container.ContainerExposePort),
			Where:    "docker-compose.yml (services.ports/expose)",
			HowToFix: fmt.Sprintf("Ensure one of your services exposes port %d in the 'ports' or 'expose' section.", chall.Container.ContainerExposePort),
		}
	}

	return nil
}

func validateDockerBuildResources(root string) error {
	dcPath := filepath.Join(root, "docker-compose.yml")
	if _, err := os.Stat(dcPath); os.IsNotExist(err) {
		return nil
	}

	//nolint:gosec // Validating user-provided challenge directory
	content, err := os.ReadFile(dcPath)
	if err != nil {
		return fmt.Errorf("failed to read docker-compose.yml: %w", err)
	}

	var compose struct {
		Services map[string]struct {
			Build interface{} `yaml:"build"`
		} `yaml:"services"`
	}

	if err := yaml.Unmarshal(content, &compose); err != nil {
		return fmt.Errorf("failed to parse docker-compose.yml: %w", err)
	}

	for name, service := range compose.Services {
		if service.Build == nil {
			continue
		}

		contextPath := "."
		dockerfilePath := "Dockerfile"

		switch b := service.Build.(type) {
		case string:
			contextPath = b
		case map[interface{}]interface{}:
			if c, ok := b["context"].(string); ok {
				contextPath = c
			}
			if d, ok := b["dockerfile"].(string); ok {
				dockerfilePath = d
			}
		}

		absContextPath := filepath.Join(root, contextPath)
		absDockerfilePath := filepath.Join(absContextPath, dockerfilePath)

		if err := validateDockerfileResources(absContextPath, absDockerfilePath); err != nil {
			if vErr, ok := err.(*ValidationError); ok {
				// Enhance error with service info
				vErr.Where = fmt.Sprintf("Service '%s' -> %s", name, vErr.Where)
			}
			return err
		}
	}

	return nil
}

func validateDockerfileResources(contextPath, dockerfilePath string) error {
	//nolint:gosec // Validating user-provided challenge directory
	f, err := os.Open(dockerfilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ValidationError{
				What:     fmt.Sprintf("Dockerfile not found: %s", filepath.Base(dockerfilePath)),
				Where:    "docker-compose.yml build configuration",
				HowToFix: "Ensure the Dockerfile exists at the specified path relative to the build context.",
			}
		}
		return err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	lineNum := 0

	// Pre-compile regex for COPY/ADD
	// Matches: COPY <src>... <dest>
	// Flag --from is ignored for file check as it refers to other stages/images
	instrRegex := regexp.MustCompile(`^(?i)(COPY|ADD)\s+(.*)`)

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Handle multi-line support? Dockerfile commands can span lines with \
		// For simplicity, we assume single line or check first part.
		// Real parsing is complex. We will try to catch common cases.

		matches := instrRegex.FindStringSubmatch(line)
		if len(matches) < 3 {
			continue
		}

		args := strings.TrimSpace(matches[2])
		// Check for flags like --from=...
		if strings.HasPrefix(args, "--") {
			continue // Skip validatinig COPY --from as it copies from other stage
		}

		// Parse arguments
		var sources []string

		if strings.HasPrefix(args, "[") {
			// JSON array format: ["src", "dest"]
			var parsed []string
			if err := json.Unmarshal([]byte(args), &parsed); err == nil && len(parsed) >= 2 {
				sources = parsed[:len(parsed)-1]
			}
		} else {
			// Space separated
			// This is fragile dealing with spaces in paths, but standard for simple Dockerfiles
			parts := strings.Fields(args)
			if len(parts) >= 2 {
				sources = parts[:len(parts)-1]
			}
		}

		for _, src := range sources {
			// Skip wildcards and URLs
			if strings.ContainsAny(src, "*?[]") || strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
				continue
			}

			srcPath := filepath.Join(contextPath, src)
			if _, err := os.Stat(srcPath); os.IsNotExist(err) {
				return &ValidationError{
					What:     fmt.Sprintf("File not found in build context: %s", src),
					Where:    fmt.Sprintf("%s:%d (%s)", filepath.Base(dockerfilePath), lineNum, line),
					HowToFix: fmt.Sprintf("Ensure '%s' exists in '%s' or check .dockerignore rules.", src, filepath.Base(contextPath)),
				}
			}
		}
	}

	return nil
}

func validateChallengeScripts(chall config.ChallengeYaml) error {
	if chall.Scripts == nil {
		return nil
	}

	if err := validateStartScript(chall.Scripts["start"]); err != nil {
		return err
	}
	if err := validateStopScript(chall.Scripts["stop"]); err != nil {
		return err
	}
	if err := validateRestartScript(chall.Scripts["restart"]); err != nil {
		return err
	}

	return nil
}

func validateStartScript(sv config.ScriptValue) error {
	cmd := strings.TrimSpace(sv.GetCommand())
	if cmd == "" {
		return nil
	}

	allowedStart1 := "cd src && docker build -t {{.slug}} ."
	allowedStart2 := "cd src && docker compose -p {{.slug}} up --build -d"

	if cmd != allowedStart1 && cmd != allowedStart2 {
		return &ValidationError{
			What:     "Invalid 'start' script",
			Where:    "challenge.yml (scripts.start)",
			HowToFix: fmt.Sprintf("Start script must be either:\n  1. %s\n  2. %s", allowedStart1, allowedStart2),
		}
	}
	return nil
}

func validateStopScript(sv config.ScriptValue) error {
	cmd := strings.TrimSpace(sv.GetCommand())
	if cmd == "" {
		return nil
	}

	allowedStop := "cd src && docker compose -p {{.slug}} down --volumes"

	if cmd != allowedStop {
		return &ValidationError{
			What:     "Invalid 'stop' script",
			Where:    "challenge.yml (scripts.stop)",
			HowToFix: fmt.Sprintf("Stop script must be: %s", allowedStop),
		}
	}
	return nil
}

func validateRestartScript(sv config.ScriptValue) error {
	cmd := strings.TrimSpace(sv.GetCommand())
	if cmd == "" {
		return nil
	}

	allowedRestart := "cd src && docker compose -p {{.slug}} restart"

	if cmd != allowedRestart {
		return &ValidationError{
			What:     "Invalid 'restart' script",
			Where:    "challenge.yml (scripts.restart)",
			HowToFix: fmt.Sprintf("Restart script must be: %s", allowedRestart),
		}
	}
	return nil
}

func validateDockerComposePrivileged(root string) error {
	dcPath := filepath.Join(root, "src/docker-compose.yml")
	if _, err := os.Stat(dcPath); os.IsNotExist(err) {
		return nil
	}

	//nolint:gosec // Validating user-provided challenge directory
	content, err := os.ReadFile(dcPath)
	if err != nil {
		return fmt.Errorf("failed to read src/docker-compose.yml: %w", err)
	}

	var compose struct {
		Services map[string]struct {
			Privileged bool `yaml:"privileged"`
		} `yaml:"services"`
	}

	if err := yaml.Unmarshal(content, &compose); err != nil {
		// Just log error or ignore if structure is very different?
		// Better to fail if it's not valid yaml if we expect it to be one.
		return fmt.Errorf("failed to parse src/docker-compose.yml: %w", err)
	}

	for name, service := range compose.Services {
		if service.Privileged {
			return &ValidationError{
				What:     fmt.Sprintf("Service '%s' uses privileged mode", name),
				Where:    "src/docker-compose.yml (services.privileged)",
				HowToFix: "Remove 'privileged: true' from the service configuration. Privileged mode is not allowed.",
			}
		}
	}

	return nil
}
