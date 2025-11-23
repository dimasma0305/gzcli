package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dimasma0305/gzcli/internal/log"
)

// GetDockerUsedPorts returns a map of ports currently used by Docker containers on the host
func GetDockerUsedPorts() (map[int]bool, error) {
	// docker ps -a --format "{{.Ports}}"
	// Output format examples:
	// 0.0.0.0:3000->80/tcp, :::3000->80/tcp
	// 0.0.0.0:80->80/tcp
	// 80/tcp, 443/tcp (no host binding)
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Ports}}")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list docker ports: %w", err)
	}

	usedPorts := make(map[int]bool)

	// Regex to find host ports: Look for patterns like ":<port>->"
	// Matches: 0.0.0.0:3000->, :::3000->, :3000->
	re := regexp.MustCompile(`:(\d+)->`)

	// We can process the whole output or line by line. Line by line is safer.
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		matches := re.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) == 2 {
				portStr := match[1]
				port, err := strconv.Atoi(portStr)
				if err == nil {
					usedPorts[port] = true
				}
			}
		}
	}

	log.Debug("Found %d ports used by Docker", len(usedPorts))
	return usedPorts, nil
}

// GetComposePortMappings extracts port mappings from Docker Compose containers
// Returns a slice of port mappings in "host:container" format
func GetComposePortMappings(configPath, projectName, cwd string) ([]string, error) {
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(cwd, configPath)
	}

	cmd := exec.Command("docker", "compose",
		"-f", configPath,
		"-p", projectName,
		"ps", "--format", "json")
	cmd.Dir = cwd

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list compose containers: %w", err)
	}

	var portMappings []string
	output := out.String()

	// Parse JSON output - each line is a JSON object
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var container struct {
			ID    string `json:"ID"`
			Ports string `json:"Ports"`
		}

		if err := json.Unmarshal([]byte(line), &container); err != nil {
			log.Debug("Failed to parse container JSON: %v", err)
			continue
		}

		if container.Ports == "" {
			continue
		}

		// Parse port string format: "0.0.0.0:30000->80/tcp, :::30000->80/tcp"
		// Extract mappings like "30000->80"
		re := regexp.MustCompile(`:(\d+)->(\d+)/`)
		matches := re.FindAllStringSubmatch(container.Ports, -1)

		for _, match := range matches {
			if len(match) == 3 {
				hostPort := match[1]
				containerPort := match[2]
				mapping := fmt.Sprintf("%s:%s", hostPort, containerPort)
				portMappings = append(portMappings, mapping)
			}
		}
	}

	log.Debug("Extracted %d port mappings from compose project %s", len(portMappings), projectName)
	return portMappings, nil
}
