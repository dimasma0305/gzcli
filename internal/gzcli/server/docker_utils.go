package server

import (
	"bytes"
	"fmt"
	"os/exec"
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
