package server

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// equalSlicesIgnoreOrder compares two slices for equality regardless of element order
func equalSlicesIgnoreOrder(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create copies to avoid modifying original slices
	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))
	copy(aCopy, a)
	copy(bCopy, b)

	// Sort both slices
	sort.Strings(aCopy)
	sort.Strings(bCopy)

	// Compare sorted slices
	return reflect.DeepEqual(aCopy, bCopy)
}

func TestPortParser_ParseDockerfilePorts(t *testing.T) {
	tests := []struct {
		name       string
		dockerfile string
		want       []string
	}{
		{
			name: "single expose",
			dockerfile: `FROM nginx:alpine
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]`,
			want: []string{"*:80"},
		},
		{
			name: "multiple expose on one line",
			dockerfile: `FROM alpine
EXPOSE 80 443 8080
CMD ["sh"]`,
			want: []string{"*:80", "*:443", "*:8080"},
		},
		{
			name: "multiple expose lines",
			dockerfile: `FROM alpine
EXPOSE 80
EXPOSE 443
EXPOSE 8080/tcp
CMD ["sh"]`,
			want: []string{"*:80", "*:443", "*:8080"},
		},
		{
			name: "with comments",
			dockerfile: `FROM alpine
# This exposes HTTP
EXPOSE 80
# This exposes HTTPS
EXPOSE 443
CMD ["sh"]`,
			want: []string{"*:80", "*:443"},
		},
		{
			name: "no expose",
			dockerfile: `FROM alpine
CMD ["sh"]`,
			want: []string{},
		},
		{
			name: "case insensitive",
			dockerfile: `FROM alpine
expose 80
Expose 443
EXPOSE 8080
CMD ["sh"]`,
			want: []string{"*:80", "*:443", "*:8080"},
		},
	}

	pp := NewPortParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a separate temp directory for each subtest to avoid race conditions
			tmpDir := t.TempDir()

			// Create temp Dockerfile
			dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
			if err := os.WriteFile(dockerfilePath, []byte(tt.dockerfile), 0600); err != nil {
				t.Fatalf("Failed to create temp Dockerfile: %v", err)
			}

			got := pp.parseDockerfilePorts(dockerfilePath)

			// Handle empty slice comparison
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDockerfilePorts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPortParser_ParseComposePorts(t *testing.T) {
	tests := []struct {
		name    string
		compose string
		want    []string
	}{
		{
			name: "simple ports mapping",
			compose: `version: '3'
services:
  web:
    image: nginx
    ports:
      - "8080:80"
      - "8443:443"`,
			want: []string{"8080:80", "8443:443"},
		},
		{
			name: "mixed ports and expose",
			compose: `version: '3'
services:
  web:
    image: nginx
    ports:
      - "8080:80"
    expose:
      - "3000"`,
			want: []string{"8080:80", "*:3000"},
		},
		{
			name: "multiple services",
			compose: `version: '3'
services:
  web:
    image: nginx
    ports:
      - "8080:80"
  api:
    image: node
    ports:
      - "3000:3000"`,
			want: []string{"8080:80", "3000:3000"},
		},
		{
			name: "no ports",
			compose: `version: '3'
services:
  worker:
    image: alpine
    command: ["sleep", "infinity"]`,
			want: []string{},
		},
		{
			name: "numeric port format",
			compose: `version: '3'
services:
  web:
    image: nginx
    ports:
      - 8080:80
      - 3000:3000`,
			want: []string{"8080:80", "3000:3000"},
		},
	}

	pp := NewPortParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a separate temp directory for each subtest to avoid race conditions
			tmpDir := t.TempDir()

			// Create temp docker-compose.yml
			composePath := filepath.Join(tmpDir, "docker-compose.yml")
			if err := os.WriteFile(composePath, []byte(tt.compose), 0600); err != nil {
				t.Fatalf("Failed to create temp compose file: %v", err)
			}

			got := pp.parseComposePorts(composePath)

			// Handle empty slice comparison
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			// Use order-agnostic comparison since map iteration order is non-deterministic
			if !equalSlicesIgnoreOrder(got, tt.want) {
				t.Errorf("parseComposePorts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPortParser_ParseKubernetesPorts(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
		want     []string
	}{
		{
			name: "service with nodePort",
			manifest: `apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  type: NodePort
  ports:
    - port: 80
      nodePort: 30080
    - port: 443
      nodePort: 30443`,
			want: []string{"30080:80", "30443:443"},
		},
		{
			name: "service without nodePort",
			manifest: `apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  type: ClusterIP
  ports:
    - port: 80
    - port: 443`,
			want: []string{"*:80", "*:443"},
		},
		{
			name: "multiple documents",
			manifest: `apiVersion: v1
kind: Service
metadata:
  name: service1
spec:
  ports:
    - port: 80
      nodePort: 30080
---
apiVersion: v1
kind: Service
metadata:
  name: service2
spec:
  ports:
    - port: 3000`,
			want: []string{"30080:80", "*:3000"},
		},
		{
			name: "deployment only (no service)",
			manifest: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: web
        image: nginx`,
			want: []string{},
		},
	}

	pp := NewPortParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a separate temp directory for each subtest to avoid race conditions
			tmpDir := t.TempDir()

			// Create temp manifest
			manifestPath := filepath.Join(tmpDir, "manifest.yaml")
			if err := os.WriteFile(manifestPath, []byte(tt.manifest), 0600); err != nil {
				t.Fatalf("Failed to create temp manifest: %v", err)
			}

			got := pp.parseKubernetesPorts(manifestPath)

			// Handle empty slice comparison
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseKubernetesPorts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPortParser_ParsePorts(t *testing.T) {
	pp := NewPortParser()
	tmpDir := t.TempDir()

	// Create test Dockerfile
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	dockerfileContent := `FROM nginx
EXPOSE 80 443`
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0600); err != nil {
		t.Fatalf("Failed to create test Dockerfile: %v", err)
	}

	// Test with relative path
	ports := pp.ParsePorts("dockerfile", "Dockerfile", tmpDir)
	expected := []string{"*:80", "*:443"}

	if !reflect.DeepEqual(ports, expected) {
		t.Errorf("ParsePorts() = %v, want %v", ports, expected)
	}

	// Test with absolute path
	ports = pp.ParsePorts("dockerfile", dockerfilePath, "")
	if !reflect.DeepEqual(ports, expected) {
		t.Errorf("ParsePorts() with absolute path = %v, want %v", ports, expected)
	}

	// Test unknown type
	ports = pp.ParsePorts("unknown", "somefile", tmpDir)
	if len(ports) != 0 {
		t.Errorf("ParsePorts() with unknown type should return empty slice, got %v", ports)
	}
}

func TestPortParser_FileNotFound(t *testing.T) {
	pp := NewPortParser()

	// Test with non-existent files
	ports := pp.parseDockerfilePorts("/nonexistent/Dockerfile")
	if len(ports) != 0 {
		t.Errorf("Expected empty ports for non-existent file, got %v", ports)
	}

	ports = pp.parseComposePorts("/nonexistent/docker-compose.yml")
	if len(ports) != 0 {
		t.Errorf("Expected empty ports for non-existent file, got %v", ports)
	}

	ports = pp.parseKubernetesPorts("/nonexistent/manifest.yaml")
	if len(ports) != 0 {
		t.Errorf("Expected empty ports for non-existent file, got %v", ports)
	}
}
