package server

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/dimasma0305/gzcli/internal/log"
)

// loadEnvFile loads environment variables from a .env file
func loadEnvFile(envFilePath string) map[string]string {
	envVars := make(map[string]string)

	//nolint:gosec // G304: Reading env files is intentional
	file, err := os.Open(envFilePath)
	if err != nil {
		// .env file is optional, don't error out
		return envVars
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Error("Failed to close env file %s: %v", envFilePath, cerr)
		}
	}()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Debug("Skipping invalid line %d in %s: %s", lineNum, envFilePath, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		envVars[key] = value
	}

	if err := scanner.Err(); err != nil {
		log.Error("Error reading env file %s: %v", envFilePath, err)
	}

	return envVars
}

// expandEnvVarsWithMap expands environment variables in a string using both system env and provided map
// Supports both ${VAR} and $VAR formats
// Priority: provided map > system environment
func expandEnvVarsWithMap(s string, envMap map[string]string) string {
	return os.Expand(s, func(key string) string {
		// Check custom env map first
		if val, ok := envMap[key]; ok {
			return val
		}
		// Fall back to system environment
		return os.Getenv(key)
	})
}

// PortParser extracts port information from configuration files
type PortParser struct{}

// NewPortParser creates a new port parser
func NewPortParser() *PortParser {
	return &PortParser{}
}

// ParsePorts extracts ports from a configuration file based on launcher type
func (pp *PortParser) ParsePorts(launcherType, configPath, cwd string) []string {
	// Make absolute path
	if !strings.HasPrefix(configPath, "/") {
		configPath = filepath.Join(cwd, configPath)
	}

	switch LauncherType(launcherType) {
	case LauncherTypeCompose:
		return pp.parseComposePorts(configPath)
	case LauncherTypeDockerfile:
		return pp.parseDockerfilePorts(configPath)
	case LauncherTypeKubernetes:
		return pp.parseKubernetesPorts(configPath)
	default:
		return []string{}
	}
}

// parseComposePorts parses ports from docker-compose.yml
func (pp *PortParser) parseComposePorts(configPath string) []string {
	//nolint:gosec // G304: Reading challenge configuration files is intentional
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Error("Failed to read compose file %s: %v", configPath, err)
		return []string{}
	}

	var compose map[string]interface{}
	if err := yaml.Unmarshal(data, &compose); err != nil {
		log.Error("Failed to parse compose file %s: %v", configPath, err)
		return []string{}
	}

	// Load environment variables from .env file (Docker Compose convention)
	composeDir := filepath.Dir(configPath)
	defaultEnvFile := filepath.Join(composeDir, ".env")
	envVars := loadEnvFile(defaultEnvFile)

	// Also load env_file entries from services
	if services, ok := compose["services"].(map[interface{}]interface{}); ok {
		for _, serviceData := range services {
			if serviceMap, ok := serviceData.(map[interface{}]interface{}); ok {
				// Load env_file if specified
				if envFile, ok := serviceMap["env_file"]; ok {
					envFiles := []string{}
					switch v := envFile.(type) {
					case string:
						envFiles = append(envFiles, v)
					case []interface{}:
						for _, ef := range v {
							if efStr, ok := ef.(string); ok {
								envFiles = append(envFiles, efStr)
							}
						}
					}

					// Load each env file
					for _, ef := range envFiles {
						envPath := ef
						if !filepath.IsAbs(ef) {
							envPath = filepath.Join(composeDir, ef)
						}
						fileEnvVars := loadEnvFile(envPath)
						// Merge into envVars (later files override earlier ones)
						for k, v := range fileEnvVars {
							envVars[k] = v
						}
					}
				}
			}
		}
	}

	var ports []string

	// Extract ports from services
	if services, ok := compose["services"].(map[interface{}]interface{}); ok {
		for serviceName, serviceData := range services {
			if serviceMap, ok := serviceData.(map[interface{}]interface{}); ok {
				// Check for ports array
				if portsList, ok := serviceMap["ports"].([]interface{}); ok {
					for _, port := range portsList {
						portStr := fmt.Sprintf("%v", port)
						// Expand environment variables in port mappings
						portStr = expandEnvVarsWithMap(portStr, envVars)
						ports = append(ports, portStr)
					}
				}

				// Also check expose
				if exposeList, ok := serviceMap["expose"].([]interface{}); ok {
					for _, port := range exposeList {
						portStr := fmt.Sprintf("%v", port)
						// Expand environment variables
						portStr = expandEnvVarsWithMap(portStr, envVars)
						// Expose without mapping, show as exposed only
						ports = append(ports, fmt.Sprintf("*:%s", portStr))
					}
				}

				log.Debug("Service %v: found %d port(s)", serviceName, len(ports))
			}
		}
	}

	return ports
}

// parseDockerfilePorts parses EXPOSE directives from Dockerfile
func (pp *PortParser) parseDockerfilePorts(configPath string) []string {
	//nolint:gosec // G304: Reading challenge configuration files is intentional
	file, err := os.Open(configPath)
	if err != nil {
		log.Error("Failed to open Dockerfile %s: %v", configPath, err)
		return []string{}
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Error("Failed to close file %s: %v", configPath, cerr)
		}
	}()

	var ports []string
	exposeRegex := regexp.MustCompile(`(?i)^EXPOSE\s+(.+)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Check for EXPOSE directive
		if matches := exposeRegex.FindStringSubmatch(line); len(matches) > 1 {
			// Parse port(s) - can be space-separated
			portsPart := strings.TrimSpace(matches[1])
			// Expand environment variables in port definitions (only system env for Dockerfile)
			portsPart = expandEnvVarsWithMap(portsPart, nil)
			portFields := strings.Fields(portsPart)

			for _, portField := range portFields {
				// Remove protocol suffix if present (e.g., "80/tcp" -> "80")
				port := strings.Split(portField, "/")[0]
				// Show as exposed (no external mapping specified)
				ports = append(ports, fmt.Sprintf("*:%s", port))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error("Error reading Dockerfile %s: %v", configPath, err)
		return []string{}
	}

	return ports
}

// parseKubernetesPorts parses ports from Kubernetes manifest
func (pp *PortParser) parseKubernetesPorts(configPath string) []string {
	//nolint:gosec // G304: Reading challenge configuration files is intentional
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Error("Failed to read Kubernetes manifest %s: %v", configPath, err)
		return []string{}
	}

	var ports []string

	// K8s manifests can contain multiple documents
	documents := strings.Split(string(data), "\n---\n")

	for _, doc := range documents {
		var manifest map[string]interface{}
		if err := yaml.Unmarshal([]byte(doc), &manifest); err != nil {
			continue
		}

		// Check if it's a Service
		if kind, ok := manifest["kind"].(string); ok && kind == "Service" {
			if spec, ok := manifest["spec"].(map[interface{}]interface{}); ok {
				if portsList, ok := spec["ports"].([]interface{}); ok {
					for _, portData := range portsList {
						if portMap, ok := portData.(map[interface{}]interface{}); ok {
							// Extract port and nodePort
							var port, nodePort interface{}
							var hasPort, hasNodePort bool

							if p, ok := portMap["port"]; ok {
								port = p
								hasPort = true
							}

							if np, ok := portMap["nodePort"]; ok {
								nodePort = np
								hasNodePort = true
							}

							if hasPort && hasNodePort {
								portMapping := fmt.Sprintf("%v:%v", nodePort, port)
								// Expand environment variables in port mappings (only system env for K8s)
								portMapping = expandEnvVarsWithMap(portMapping, nil)
								ports = append(ports, portMapping)
							} else if hasPort {
								portMapping := fmt.Sprintf("*:%v", port)
								// Expand environment variables (only system env for K8s)
								portMapping = expandEnvVarsWithMap(portMapping, nil)
								ports = append(ports, portMapping)
							}
						}
					}
				}
			}
		}
	}

	return ports
}
