package uploadserver

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessUpload_Validation(t *testing.T) {
	const (
		event    = "EventVal"
		category = "Web"
	)

	// Helper to run a test case
	runCase := func(name string, cfg buildChallengeArchiveConfig, wantError string) {
		t.Run(name, func(t *testing.T) {
			_ = setupWorkspace(t, event, category)
			archive := buildChallengeArchive(t, cfg)

			file, err := os.Open(filepath.Clean(archive))
			if err != nil {
				t.Fatalf("failed to open archive: %v", err)
			}
			t.Cleanup(func() { _ = file.Close() })

			srv := newTestServer(t)
			err = srv.processUpload(context.Background(), event, category, file, "val.zip")

			if wantError == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", wantError)
				} else if !strings.Contains(err.Error(), wantError) {
					t.Errorf("expected error containing %q, got %q", wantError, err.Error())
				}
			}
		})
	}

	// 1. Missing Dashboard Config (Invalid value)
	runCase("MissingDashboardConfig", buildChallengeArchiveConfig{
		ChallengeYAML: `name: "D1"
author: "a"
type: "StaticAttachment"
value: 1
flags: ["f"]
dashboard:
  type: "grafana"
  config: "./src/docker-compose.yml"
`,
		IncludeSolver: true,
		DistFiles: map[string]string{".gitkeep": ""},
	}, "Dashboard config file not found")

	// 2. Valid Dashboard Config
	runCase("ValidDashboardConfig", buildChallengeArchiveConfig{
		ChallengeYAML: `name: "D2"
author: "a"
type: "StaticAttachment"
value: 1
flags: ["f"]
dashboard:
  type: "grafana"
  config: "./src/docker-compose.yml"
`,
		SrcFiles:      map[string]string{"docker-compose.yml": "services:\n  val:\n    image: nginx"},
		IncludeSolver: true,
		DistFiles:     map[string]string{".gitkeep": ""},
	}, "")

	// 3. DynamicContainer missing docker-compose.yml
	runCase("MissingDockerCompose", buildChallengeArchiveConfig{
		ChallengeYAML: `name: "D3"
author: "a"
type: "DynamicContainer"
value: 1
flags: ["f"]
container:
  flagTemplate: "f"
  containerImage: "i"
`,
		IncludeSolver: true,
		DistFiles: map[string]string{".gitkeep": ""},
	}, "Missing docker-compose.yml")

	// 4. StaticContainer missing Dockerfile (local image)
	runCase("MissingDockerfile", buildChallengeArchiveConfig{
		ChallengeYAML: `name: "D4"
author: "a"
type: "StaticContainer"
value: 1
flags: ["f"]
container:
  containerImage: "my-local-image"
`,
		IncludeSolver: true,
		DistFiles: map[string]string{".gitkeep": ""},
	}, "Missing Dockerfile")

	// 5. Exposed Port Mismatch
	runCase("ExposedPortMismatch", buildChallengeArchiveConfig{
		ChallengeYAML: `name: "D5"
author: "a"
type: "DynamicContainer"
value: 1
flags: ["f"]
container:
  flagTemplate: "f"
  containerImage: "i"
  exposePort: 9090
`,
		IncludeSolver: true,
		ExtraRootFiles: map[string]string{
			"docker-compose.yml": `services:
  web:
    image: nginx
    ports:
      - "80:80"
`,
		},
		DistFiles: map[string]string{".gitkeep": ""},
	}, "Exposed port 9090 not found")

	// 6. Valid Exposed Port
	runCase("ValidExposedPort", buildChallengeArchiveConfig{
		ChallengeYAML: `name: "D6"
author: "a"
type: "DynamicContainer"
value: 1
flags: ["f"]
container:
  flagTemplate: "f"
  containerImage: "i"
  exposePort: 80
`,
		IncludeSolver: true,
		ExtraRootFiles: map[string]string{
			"docker-compose.yml": `services:
  web:
    image: nginx
    ports:
      - "80:80"
`,
		},
		DistFiles: map[string]string{".gitkeep": ""},
	}, "")

    // 7. Missing Build Resource (Only checked if Dashboard is present)
    runCase("MissingBuildResourceWithPath", buildChallengeArchiveConfig{
        ChallengeYAML: `name: "D7"
author: "a"
type: "DynamicContainer"
value: 1
flags: ["f"]
dashboard:
  type: "grafana"
  config: "./src/docker-compose.yml"
container:
    flagTemplate: "f"
    containerImage: "i"
`,
        IncludeSolver: true,
        SrcFiles: map[string]string{
            "docker-compose.yml": "services:\n  dummy:\n    image: nginx",
        },
        ExtraRootFiles: map[string]string{
            "docker-compose.yml": `services:
  web:
    build: .
`,
            "Dockerfile": `FROM alpine
COPY missing.txt /app/
`,
        },
        DistFiles: map[string]string{".gitkeep": ""},
    }, "File not found in build context: missing.txt")

    // 8. Missing Build Resource Ignored if no Dashboard
    runCase("MissingBuildResourceIgnored", buildChallengeArchiveConfig{
        ChallengeYAML: `name: "D8"
author: "a"
type: "DynamicContainer"
value: 1
flags: ["f"]
container:
    flagTemplate: "f"
    containerImage: "i"
`,
        IncludeSolver: true,
        ExtraRootFiles: map[string]string{
            "docker-compose.yml": `services:
  web:
    build: .
`,
            "Dockerfile": `FROM alpine
COPY missing.txt /app/
`,
        },
        DistFiles: map[string]string{".gitkeep": ""},
    }, "")
	// 9. Invalid Script
	runCase("InvalidScript", buildChallengeArchiveConfig{
		ChallengeYAML: `name: "D9"
author: "a"
type: "DynamicContainer"
value: 1
flags: ["f"]
container:
  flagTemplate: "f"
  containerImage: "i"
scripts:
  start: "echo hello"
`,
		IncludeSolver: true,
		DistFiles: map[string]string{".gitkeep": ""},
		ExtraRootFiles: map[string]string{
			"docker-compose.yml": "services:\n  web:\n    image: nginx\n",
		},
	}, "Invalid 'start' script")

	// 10. Valid Script
	runCase("ValidScript", buildChallengeArchiveConfig{
		ChallengeYAML: `name: "D10"
author: "a"
type: "DynamicContainer"
value: 1
flags: ["f"]
container:
  flagTemplate: "f"
  containerImage: "i"
scripts:
  start: "cd src && docker build -t {{.slug}} ."
`,
		IncludeSolver: true,
		DistFiles: map[string]string{".gitkeep": ""},
		ExtraRootFiles: map[string]string{
			"docker-compose.yml": "services:\n  web:\n    image: nginx\n",
		},
	}, "")
	// 11. Privileged Service Rejection
	runCase("PrivilegedService", buildChallengeArchiveConfig{
		ChallengeYAML: `name: "D11"
author: "a"
type: "DynamicContainer"
value: 1
flags: ["f"]
container:
  flagTemplate: "f"
  containerImage: "i"
`,
		IncludeSolver: true,
		DistFiles: map[string]string{".gitkeep": ""},
		ExtraRootFiles: map[string]string{
			"docker-compose.yml": "services:\n  web:\n    image: nginx\n",
		},
		SrcFiles: map[string]string{
			"docker-compose.yml": `services:
  app:
    image: alpine
    privileged: true
`,
		},
	}, "uses privileged mode")

	// 12. Non-Privileged Service Allowed
	runCase("NonPrivilegedService", buildChallengeArchiveConfig{
		ChallengeYAML: `name: "D12"
author: "a"
type: "DynamicContainer"
value: 1
flags: ["f"]
container:
  flagTemplate: "f"
  containerImage: "i"
`,
		IncludeSolver: true,
		DistFiles: map[string]string{".gitkeep": ""},
		ExtraRootFiles: map[string]string{
			"docker-compose.yml": "services:\n  web:\n    image: nginx\n",
		},
		SrcFiles: map[string]string{
			"docker-compose.yml": `services:
  app:
    image: alpine
    privileged: false
`,
		},
	}, "")
}
