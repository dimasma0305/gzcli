package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/dimasma0305/gzcli/internal/log"
)

// isValidSlug returns true for safe slugs (lowercase letters, digits, hyphen, underscore)
func isValidSlug(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' && r != '_' {
			return false
		}
	}
	return s != ""
}

// isSafeConfigPath ensures configPath is absolute and under baseDir
func isSafeConfigPath(configPath, baseDir string) bool {
	// clean and make absolute
	p := filepath.Clean(configPath)
	if !filepath.IsAbs(p) {
		p = filepath.Join(baseDir, p)
	}
	rel, err := filepath.Rel(baseDir, p)
	if err != nil {
		return false
	}
	// disallow paths that escape baseDir via .. and ensure file exists is optional
	if strings.HasPrefix(rel, "..") {
		return false
	}
	return true
}

// Executor handles challenge lifecycle operations
type Executor struct {
	timeout time.Duration
}

// NewExecutor creates a new executor
func NewExecutor() *Executor {
	return &Executor{
		timeout: 10 * time.Minute, // Increased for build operations
	}
}

// Start starts a challenge
func (e *Executor) Start(challenge *ChallengeInfo) error {
	if challenge.Dashboard == nil {
		return fmt.Errorf("challenge has no dashboard configuration")
	}

	dashboard := challenge.Dashboard
	launcherType := LauncherType(dashboard.Type)

	switch launcherType {
	case LauncherTypeCompose:
		return e.startCompose(challenge, dashboard)
	case LauncherTypeDockerfile:
		return e.startDockerfile(challenge, dashboard)
	case LauncherTypeKubernetes:
		return e.startKubernetes(challenge, dashboard)
	default:
		return fmt.Errorf("unknown launcher type: %s", dashboard.Type)
	}
}

// Stop stops a challenge
func (e *Executor) Stop(challenge *ChallengeInfo) error {
	if challenge.Dashboard == nil {
		return fmt.Errorf("challenge has no dashboard configuration")
	}

	dashboard := challenge.Dashboard
	launcherType := LauncherType(dashboard.Type)

	switch launcherType {
	case LauncherTypeCompose:
		return e.stopCompose(challenge, dashboard)
	case LauncherTypeDockerfile:
		return e.stopDockerfile(challenge)
	case LauncherTypeKubernetes:
		return e.stopKubernetes(challenge, dashboard)
	default:
		return fmt.Errorf("unknown launcher type: %s", dashboard.Type)
	}
}

// Restart restarts a challenge (stop then start)
func (e *Executor) Restart(challenge *ChallengeInfo) error {
	log.InfoH2("Restarting challenge: %s", challenge.Name)

	if err := e.Stop(challenge); err != nil {
		log.Error("Stop failed during restart: %v", err)
		// Continue anyway - the service might not be running
	}

	// Small delay between stop and start
	time.Sleep(2 * time.Second)

	if err := e.Start(challenge); err != nil {
		return fmt.Errorf("start failed during restart: %w", err)
	}

	return nil
}

// randomizeComposePorts randomizes host ports in a compose file structure
// Returns the modified compose structure and allocated port mappings
func randomizeComposePorts(compose map[string]interface{}, usedDockerPorts map[int]bool) (map[string]interface{}, []string, error) {
	// Deep copy the compose structure to avoid modifying the original
	composeBytes, err := yaml.Marshal(compose)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal compose: %w", err)
	}

	var modifiedCompose map[string]interface{}
	if err := yaml.Unmarshal(composeBytes, &modifiedCompose); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal compose: %w", err)
	}

	services, ok := modifiedCompose["services"].(map[interface{}]interface{})
	if !ok {
		return modifiedCompose, []string{}, nil
	}

	var allocatedPorts []string
	allocatedHostPorts := make(map[int]bool)

	for _, serviceData := range services {
		serviceMap, ok := serviceData.(map[interface{}]interface{})
		if !ok {
			continue
		}

		portsList, ok := serviceMap["ports"].([]interface{})
		if !ok {
			continue
		}

		modifiedPorts := make([]interface{}, 0, len(portsList))

		for _, port := range portsList {
			portStr := fmt.Sprintf("%v", port)
			parts := strings.Split(portStr, ":")

			// Handle different port formats: "host:container", "container", "*:container"
			if len(parts) == 1 {
				// Just container port, no host port - skip randomization
				modifiedPorts = append(modifiedPorts, port)
				continue
			}

			containerPort := parts[len(parts)-1]
			hostPortStr := parts[0]

			// Skip if host port is "*" (wildcard) or empty
			if hostPortStr == "*" || hostPortStr == "" {
				modifiedPorts = append(modifiedPorts, port)
				continue
			}

			// Check if host port is already a number (explicit port)
			_, err := strconv.Atoi(hostPortStr)
			if err != nil {
				// Not a number, might be a variable or other format - skip
				modifiedPorts = append(modifiedPorts, port)
				continue
			}

			// Combine global used ports with local allocated ports
			excludedPorts := make(map[int]bool)
			for p := range usedDockerPorts {
				excludedPorts[p] = true
			}
			for p := range allocatedHostPorts {
				excludedPorts[p] = true
			}

			// Get a random free port on host
			randomHostPort, err := GetRandomPort(30000, 65535, excludedPorts)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to allocate random port: %w", err)
			}

			allocatedHostPorts[randomHostPort] = true
			newPortMapping := fmt.Sprintf("%d:%s", randomHostPort, containerPort)
			modifiedPorts = append(modifiedPorts, newPortMapping)

			allocatedPorts = append(allocatedPorts, newPortMapping)
			log.Info("Randomized port mapping: %s -> %s", portStr, newPortMapping)
		}

		serviceMap["ports"] = modifiedPorts
	}

	return modifiedCompose, allocatedPorts, nil
}

// ensureGitignorePattern ensures a pattern is in .gitignore in the given directory
func ensureGitignorePattern(dir, pattern string) error {
	gitignorePath := filepath.Join(dir, ".gitignore")

	// Read existing .gitignore if it exists
	var existing string
	//nolint:gosec // G304: Path is constructed from challenge directory
	if data, err := os.ReadFile(gitignorePath); err == nil {
		existing = string(data)
	}

	// Check if pattern already exists
	if strings.Contains(existing, pattern) {
		return nil // Already present
	}

	// Append pattern if .gitignore exists, otherwise create new
	var content string
	if existing != "" {
		if !strings.HasSuffix(existing, "\n") {
			existing += "\n"
		}
		content = existing + pattern + "\n"
	} else {
		content = pattern + "\n"
	}

	// Write .gitignore
	//nolint:gosec // G304: Path is constructed from challenge directory
	if err := os.WriteFile(gitignorePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}

	log.Debug("Added %s to .gitignore in %s", pattern, dir)
	return nil
}

// startCompose starts a Docker Compose challenge
func (e *Executor) startCompose(challenge *ChallengeInfo, dashboard *Dashboard) error {
	configPath := dashboard.Config
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(challenge.Cwd, configPath)
	}

	log.InfoH2("Starting Docker Compose: %s", challenge.Name)
	log.InfoH3("Config: %s, Project: %s", configPath, challenge.Slug)

	// Read and parse the compose file
	//nolint:gosec // G304: Reading challenge configuration files is intentional
	composeData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %w", err)
	}

	var compose map[string]interface{}
	if err := yaml.Unmarshal(composeData, &compose); err != nil {
		return fmt.Errorf("failed to parse compose file: %w", err)
	}

	// Get currently used ports on Docker host
	usedDockerPorts, err := GetDockerUsedPorts()
	if err != nil {
		log.Error("Failed to get used docker ports: %v", err)
		usedDockerPorts = make(map[int]bool)
	}

	// Randomize ports in the compose structure
	modifiedCompose, allocatedPorts, err := randomizeComposePorts(compose, usedDockerPorts)
	if err != nil {
		return fmt.Errorf("failed to randomize ports: %w", err)
	}

	// Create temporary compose file in the same directory
	composeDir := filepath.Dir(configPath)

	// Ensure temp file pattern is gitignored
	tempPattern := "docker-compose.*.tmp.yml"
	if err := ensureGitignorePattern(composeDir, tempPattern); err != nil {
		log.Error("Failed to update .gitignore: %v", err)
		// Continue anyway - not critical
	}

	tempFile, err := os.CreateTemp(composeDir, fmt.Sprintf("docker-compose.%s.tmp.yml", challenge.Slug))
	if err != nil {
		return fmt.Errorf("failed to create temp compose file: %w", err)
	}
	tempFilePath := tempFile.Name()

	// Ensure temp file is cleaned up
	defer func() {
		if err := os.Remove(tempFilePath); err != nil {
			log.Error("Failed to remove temp compose file %s: %v", tempFilePath, err)
		}
	}()

	// Write modified compose to temp file
	modifiedComposeData, err := yaml.Marshal(modifiedCompose)
	if err != nil {
		return fmt.Errorf("failed to marshal modified compose: %w", err)
	}

	if _, err := tempFile.Write(modifiedComposeData); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("failed to write temp compose file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp compose file: %w", err)
	}

	// Store allocated ports before starting
	challenge.SetAllocatedPorts(allocatedPorts)
	if len(allocatedPorts) > 0 {
		log.Info("Allocated port mappings: %v", allocatedPorts)
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	// Use the temp file for docker compose
	//nolint:gosec // G204: Docker commands with challenge config are intentional
	//nolint:gosec // G204: Docker commands with challenge config are intentional
	//nolint:gosec // G204: Docker commands with challenge config are intentional
	cmd := exec.CommandContext(ctx, "docker", "compose",
		"-f", tempFilePath,
		"-p", challenge.Slug,
		"up", "-d", "--build")
	cmd.Dir = challenge.Cwd

	// Capture output for debugging
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		// Clear allocated ports on failure
		challenge.SetAllocatedPorts(nil)
		log.Error("Docker Compose failed: %v", err)
		log.Error("Stdout: %s", stdout.String())
		log.Error("Stderr: %s", stderr.String())
		return fmt.Errorf("docker compose up failed: %w", err)
	}

	log.InfoH3("Docker Compose started successfully")
	log.Debug("Output: %s", stdout.String())
	return nil
}

// stopCompose stops a Docker Compose challenge
func (e *Executor) stopCompose(challenge *ChallengeInfo, dashboard *Dashboard) error {
	configPath := dashboard.Config
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(challenge.Cwd, configPath)
	}

	if !isSafeConfigPath(configPath, challenge.Cwd) {
		return fmt.Errorf("unsafe compose config path: %s", dashboard.Config)
	}

	log.InfoH2("Stopping Docker Compose: %s", challenge.Name)

	// Clear allocated ports
	challenge.SetAllocatedPorts(nil)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	//nolint:gosec // G204: Docker commands with challenge config are intentional and configPath is validated
	cmd := exec.CommandContext(ctx, "docker", "compose",
		"-f", configPath,
		"-p", challenge.Slug,
		"down", "--volumes")
	cmd.Dir = challenge.Cwd

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose down failed: %w\nOutput: %s", err, string(output))
	}

	log.InfoH3("Docker Compose stopped successfully")
	return nil
}

// startDockerfile starts a Dockerfile-based challenge
func (e *Executor) startDockerfile(challenge *ChallengeInfo, dashboard *Dashboard) error {
	configPath := dashboard.Config
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(challenge.Cwd, configPath)
	}

	log.InfoH2("Starting Dockerfile: %s", challenge.Name)

	// Build the image
	log.InfoH3("Building image: %s:latest", challenge.Slug)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	//nolint:gosec // G204: Docker commands with challenge config are intentional
	buildCmd := exec.CommandContext(ctx, "docker", "build",
		"-t", fmt.Sprintf("%s:latest", challenge.Slug),
		"-f", configPath,
		".")
	buildCmd.Dir = challenge.Cwd

	output, err := buildCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker build failed: %w\nOutput: %s", err, string(output))
	}

	// Start the container
	log.InfoH3("Starting container: %s", challenge.Slug)

	args := []string{"run", "-d", "--name", challenge.Slug}

	// Get currently used ports on Docker host
	usedDockerPorts, err := GetDockerUsedPorts()
	if err != nil {
		// Just log warning and continue with empty map (optimistic allocation)
		log.Error("Failed to get used docker ports: %v", err)
		usedDockerPorts = make(map[int]bool)
	}

	// Add port mappings with randomization
	allocatedPorts := make([]string, 0, len(dashboard.Ports))
	// Track allocated ports for this container to avoid duplicates
	allocatedHostPorts := make(map[int]bool)

	for _, portMap := range dashboard.Ports {
		// portMap could be "host:container" or "container" or "*:container"
		parts := strings.Split(portMap, ":")
		containerPort := parts[len(parts)-1] // Always the last part

		// Combine global used ports with local allocated ports
		excludedPorts := make(map[int]bool)
		for p := range usedDockerPorts {
			excludedPorts[p] = true
		}
		for p := range allocatedHostPorts {
			excludedPorts[p] = true
		}

		// Get a random free port on host, excluding already allocated ones
		hostPort, err := GetRandomPort(30000, 65535, excludedPorts)
		if err != nil {
			return fmt.Errorf("failed to allocate port: %w", err)
		}

		allocatedHostPorts[hostPort] = true
		mapping := fmt.Sprintf("%d:%s", hostPort, containerPort)
		args = append(args, "-p", mapping)
		allocatedPorts = append(allocatedPorts, mapping)
		log.Info("Allocated port mapping: %s", mapping)
	}

	// Store allocated ports
	challenge.SetAllocatedPorts(allocatedPorts)

	args = append(args, fmt.Sprintf("%s:latest", challenge.Slug))

	//nolint:gosec // G204: Docker commands with challenge config are intentional
	runCmd := exec.Command("docker", args...)
	runCmd.Dir = challenge.Cwd

	output, err = runCmd.CombinedOutput()
	if err != nil {
		// Clear allocated ports on failure
		challenge.SetAllocatedPorts(nil)
		return fmt.Errorf("docker run failed: %w\nOutput: %s", err, string(output))
	}

	log.InfoH3("Dockerfile container started successfully")
	return nil
}

// stopDockerfile stops a Dockerfile-based challenge
func (e *Executor) stopDockerfile(challenge *ChallengeInfo) error {
	log.InfoH2("Stopping Dockerfile container: %s", challenge.Name)

	// Clear allocated ports
	challenge.SetAllocatedPorts(nil)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	// Stop the container
	//nolint:gosec // G204: Docker commands with challenge config are intentional
	stopCmd := exec.CommandContext(ctx, "docker", "stop", challenge.Slug)
	if output, err := stopCmd.CombinedOutput(); err != nil {
		log.Error("docker stop failed: %v\nOutput: %s", err, string(output))
		// Continue to try removing
	}

	// Remove the container
	//nolint:gosec // G204: Docker commands with challenge config are intentional
	rmCmd := exec.Command("docker", "rm", challenge.Slug)
	output, err := rmCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker rm failed: %w\nOutput: %s", err, string(output))
	}

	log.InfoH3("Dockerfile container stopped and removed successfully")
	return nil
}

// startKubernetes starts a Kubernetes-based challenge
func (e *Executor) startKubernetes(challenge *ChallengeInfo, dashboard *Dashboard) error {
	configPath := dashboard.Config
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(challenge.Cwd, configPath)
	}

	if !isSafeConfigPath(configPath, challenge.Cwd) {
		return fmt.Errorf("unsafe kubernetes manifest path: %s", dashboard.Config)
	}

	if !isValidSlug(challenge.Slug) {
		return fmt.Errorf("invalid challenge slug: %s", challenge.Slug)
	}

	log.InfoH2("Starting Kubernetes: %s", challenge.Name)
	log.InfoH3("Manifest: %s", configPath)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	//nolint:gosec // G204: kubectl apply is intended; manifest path is validated above
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", configPath)
	cmd.Dir = challenge.Cwd

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl apply failed: %w\nOutput: %s", err, string(output))
	}

	log.InfoH3("Kubernetes resources created successfully")
	return nil
}

// stopKubernetes stops a Kubernetes-based challenge
func (e *Executor) stopKubernetes(challenge *ChallengeInfo, dashboard *Dashboard) error {
	configPath := dashboard.Config
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(challenge.Cwd, configPath)
	}

	if !isSafeConfigPath(configPath, challenge.Cwd) {
		return fmt.Errorf("unsafe kubernetes manifest path: %s", dashboard.Config)
	}

	if !isValidSlug(challenge.Slug) {
		return fmt.Errorf("invalid challenge slug: %s", challenge.Slug)
	}

	log.InfoH2("Stopping Kubernetes: %s", challenge.Name)

	// Clear allocated ports
	challenge.SetAllocatedPorts(nil)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	//nolint:gosec // G204: kubectl delete is intended; manifest path is validated above
	cmd := exec.CommandContext(ctx, "kubectl", "delete", "-f", configPath)
	cmd.Dir = challenge.Cwd

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl delete failed: %w\nOutput: %s", err, string(output))
	}

	log.InfoH3("Kubernetes resources deleted successfully")
	return nil
}

// CheckHealth checks if a challenge is running
func (e *Executor) CheckHealth(challenge *ChallengeInfo) (bool, error) {
	if challenge.Dashboard == nil {
		return false, fmt.Errorf("challenge has no dashboard configuration")
	}

	dashboard := challenge.Dashboard
	launcherType := LauncherType(dashboard.Type)

	switch launcherType {
	case LauncherTypeCompose:
		return e.checkHealthCompose(challenge)
	case LauncherTypeDockerfile:
		return e.checkHealthDockerfile(challenge)
	case LauncherTypeKubernetes:
		return e.checkHealthKubernetes(challenge)
	default:
		return false, fmt.Errorf("unknown launcher type: %s", dashboard.Type)
	}
}

// checkHealthCompose checks Docker Compose health
func (e *Executor) checkHealthCompose(challenge *ChallengeInfo) (bool, error) {
	if challenge.Dashboard == nil {
		return false, nil
	}

	configPath := challenge.Dashboard.Config
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(challenge.Cwd, configPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//nolint:gosec // G204: Docker commands with challenge config are intentional
	cmd := exec.CommandContext(ctx, "docker", "compose",
		"-f", configPath,
		"-p", challenge.Slug,
		"ps", "--format", "json")
	cmd.Dir = challenge.Cwd

	output, err := cmd.Output()
	if err != nil {
		return false, nil // Not running
	}

	// Parse JSON output
	var containers []map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(output))
	for decoder.More() {
		var container map[string]interface{}
		if err := decoder.Decode(&container); err != nil {
			continue
		}
		containers = append(containers, container)
	}

	// Check if any containers are running
	for _, container := range containers {
		if state, ok := container["State"].(string); ok {
			if strings.Contains(strings.ToLower(state), "running") {
				return true, nil
			}
		}
	}

	return false, nil
}

// checkHealthDockerfile checks Dockerfile container health
func (e *Executor) checkHealthDockerfile(challenge *ChallengeInfo) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//nolint:gosec // G204: Docker commands for health checks are intentional
	cmd := exec.CommandContext(ctx, "docker", "ps",
		"--filter", fmt.Sprintf("name=%s", challenge.Slug),
		"--format", "json")

	output, err := cmd.Output()
	if err != nil {
		return false, nil
	}

	return len(output) > 0, nil
}

// checkHealthKubernetes checks Kubernetes pod health
func (e *Executor) checkHealthKubernetes(challenge *ChallengeInfo) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//nolint:gosec // G204: Kubectl commands for health checks are intentional
	cmd := exec.CommandContext(ctx, "kubectl", "get", "pods",
		"-l", fmt.Sprintf("app=%s", challenge.Slug),
		"-o", "json")

	output, err := cmd.Output()
	if err != nil {
		return false, nil
	}

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return false, nil
	}

	// Check if items array is not empty
	if items, ok := result["items"].([]interface{}); ok {
		return len(items) > 0, nil
	}

	return false, nil
}
