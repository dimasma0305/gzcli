package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/dimasma0305/gzcli/internal/log"
)

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

// startCompose starts a Docker Compose challenge
func (e *Executor) startCompose(challenge *ChallengeInfo, dashboard *Dashboard) error {
	configPath := dashboard.Config
	if !strings.HasPrefix(configPath, "/") {
		configPath = fmt.Sprintf("%s/%s", challenge.Cwd, configPath)
	}

	log.InfoH2("Starting Docker Compose: %s", challenge.Name)
	log.InfoH3("Config: %s, Project: %s", configPath, challenge.Slug)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	//nolint:gosec // G204: Docker commands with challenge config are intentional
	cmd := exec.CommandContext(ctx, "docker", "compose",
		"-f", configPath,
		"-p", challenge.Slug,
		"up", "-d", "--build")
	cmd.Dir = challenge.Cwd

	// Capture output for debugging
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
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
	if !strings.HasPrefix(configPath, "/") {
		configPath = fmt.Sprintf("%s/%s", challenge.Cwd, configPath)
	}

	log.InfoH2("Stopping Docker Compose: %s", challenge.Name)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	//nolint:gosec // G204: Docker commands with challenge config are intentional
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
	if !strings.HasPrefix(configPath, "/") {
		configPath = fmt.Sprintf("%s/%s", challenge.Cwd, configPath)
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

	// Add port mappings
	for _, portMap := range dashboard.Ports {
		args = append(args, "-p", portMap)
	}

	args = append(args, fmt.Sprintf("%s:latest", challenge.Slug))

	//nolint:gosec // G204: Docker commands with challenge config are intentional
	runCmd := exec.Command("docker", args...)
	runCmd.Dir = challenge.Cwd

	output, err = runCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker run failed: %w\nOutput: %s", err, string(output))
	}

	log.InfoH3("Dockerfile container started successfully")
	return nil
}

// stopDockerfile stops a Dockerfile-based challenge
func (e *Executor) stopDockerfile(challenge *ChallengeInfo) error {
	log.InfoH2("Stopping Dockerfile container: %s", challenge.Name)

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
	if !strings.HasPrefix(configPath, "/") {
		configPath = fmt.Sprintf("%s/%s", challenge.Cwd, configPath)
	}

	log.InfoH2("Starting Kubernetes: %s", challenge.Name)
	log.InfoH3("Manifest: %s", configPath)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	//nolint:gosec // G204: Kubectl commands with challenge config are intentional
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
	if !strings.HasPrefix(configPath, "/") {
		configPath = fmt.Sprintf("%s/%s", challenge.Cwd, configPath)
	}

	log.InfoH2("Stopping Kubernetes: %s", challenge.Name)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	//nolint:gosec // G204: Kubectl commands with challenge config are intentional
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
	if !strings.HasPrefix(configPath, "/") {
		configPath = fmt.Sprintf("%s/%s", challenge.Cwd, configPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//nolint:gosec // G204: Docker commands for health checks are intentional
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
