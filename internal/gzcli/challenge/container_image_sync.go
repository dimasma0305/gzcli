package challenge

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dimasma0305/gzcli/internal/log"
)

func isContainerChallengeType(t string) bool {
	switch strings.TrimSpace(t) {
	case "StaticContainer", "DynamicContainer":
		return true
	default:
		return false
	}
}

// parseRegistryServerAddress returns:
// - repoPrefix: where images should be tagged/pushed (can include a namespace path)
// - loginServer: the registry host used for docker login (no namespace path)
func parseRegistryServerAddress(serverAddress string) (repoPrefix string, loginServer string) {
	s := strings.TrimSpace(serverAddress)
	s = strings.TrimSuffix(s, "/")
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	if s == "" {
		return "", ""
	}

	repoPrefix = s

	parts := strings.SplitN(s, "/", 2)
	loginServer = parts[0]
	if loginServer == "" {
		return "", ""
	}
	return repoPrefix, loginServer
}

func resolveDockerBuildContext(challengeCwd string, containerImageValue string) (contextDir string, dockerfile string) {
	// If containerImage points to an existing path (dir or Dockerfile), prefer it as build context.
	// This supports setups like: containerImage: ../../shared/docker-images/web-base
	if p := strings.TrimSpace(containerImageValue); p != "" {
		candidate := p
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(challengeCwd, candidate)
		}
		if st, statErr := os.Stat(candidate); statErr == nil {
			if st.IsDir() {
				return candidate, ""
			}
			// If it's a file, treat it as an explicit Dockerfile.
			return filepath.Dir(candidate), candidate
		}
	}

	// Otherwise, follow common challenge layouts.
	srcDockerfile := filepath.Join(challengeCwd, "src", "Dockerfile")
	if st, statErr := os.Stat(srcDockerfile); statErr == nil && !st.IsDir() {
		return filepath.Dir(srcDockerfile), srcDockerfile
	}

	rootDockerfile := filepath.Join(challengeCwd, "Dockerfile")
	if st, statErr := os.Stat(rootDockerfile); statErr == nil && !st.IsDir() {
		return challengeCwd, rootDockerfile
	}

	// Fallback: let docker decide; this will error clearly if no Dockerfile exists.
	return challengeCwd, ""
}

func containerImageResolvesToLocalPath(challengeCwd string, containerImageValue string) bool {
	p := strings.TrimSpace(containerImageValue)
	if p == "" {
		return false
	}
	candidate := p
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(challengeCwd, candidate)
	}
	_, err := os.Stat(candidate)
	return err == nil
}

func dockerBuild(ctx context.Context, dir string, dockerfile string, tag string) error {
	args := []string{"build", "-t", tag}
	if strings.TrimSpace(dockerfile) != "" {
		args = append(args, "-f", dockerfile)
	}
	args = append(args, ".")
	return runDocker(ctx, dir, args, "")
}

func dockerTag(ctx context.Context, src string, dst string) error {
	return runDocker(ctx, "", []string{"tag", src, dst}, "")
}

func dockerPush(ctx context.Context, image string) error {
	return runDocker(ctx, "", []string{"push", image}, "")
}

type registryLoginState struct {
	done bool
	err  error
}

var (
	registryLoginMu    sync.Mutex
	registryLoginCache = make(map[string]registryLoginState)
)

const (
	defaultDockerBuildTimeout = 20 * time.Minute
	defaultDockerPushTimeout  = 20 * time.Minute
	defaultDockerLoginTimeout = 2 * time.Minute
	defaultDockerTagTimeout   = 1 * time.Minute
)

func getDockerBuildTimeout() time.Duration {
	return getDockerTimeoutFromEnv("GZCLI_DOCKER_BUILD_TIMEOUT", defaultDockerBuildTimeout)
}

func getDockerPushTimeout() time.Duration {
	return getDockerTimeoutFromEnv("GZCLI_DOCKER_PUSH_TIMEOUT", defaultDockerPushTimeout)
}

func getDockerLoginTimeout() time.Duration {
	return getDockerTimeoutFromEnv("GZCLI_DOCKER_LOGIN_TIMEOUT", defaultDockerLoginTimeout)
}

func getDockerTagTimeout() time.Duration {
	return getDockerTimeoutFromEnv("GZCLI_DOCKER_TAG_TIMEOUT", defaultDockerTagTimeout)
}

func getDockerTimeoutFromEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	dur, err := time.ParseDuration(raw)
	if err != nil || dur <= 0 {
		log.Info("Invalid %s=%q, using default %s", key, raw, fallback)
		return fallback
	}

	return dur
}

func dockerLoginOnce(ctx context.Context, server string, username string, password string) error {
	server = strings.TrimSpace(server)
	if server == "" {
		return fmt.Errorf("docker login: empty server")
	}

	registryLoginMu.Lock()
	if st, ok := registryLoginCache[server]; ok && st.done {
		registryLoginMu.Unlock()
		return st.err
	}
	registryLoginMu.Unlock()

	// Prefer password-stdin to avoid leaking passwords in process args.
	args := []string{"login", server, "-u", username, "--password-stdin"}
	err := runDocker(ctx, "", args, password+"\n")

	registryLoginMu.Lock()
	registryLoginCache[server] = registryLoginState{done: true, err: err}
	registryLoginMu.Unlock()

	return err
}

func runDocker(ctx context.Context, dir string, args []string, stdin string) error {
	cmd := exec.CommandContext(ctx, "docker", args...)
	if strings.TrimSpace(dir) != "" {
		cmd.Dir = dir
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker %s failed (dir=%s): %w\n--- output ---\n%s",
			strings.Join(args, " "),
			dir,
			err,
			tailForDocker(out.String(), 200),
		)
	}

	// Keep docker output available in logs for debugging without being too noisy.
	if s := strings.TrimSpace(out.String()); s != "" {
		log.Debug("docker %s output:\n%s", strings.Join(args, " "), tailForDocker(s, 200))
	}
	return nil
}

func tailForDocker(s string, maxLines int) string {
	if maxLines <= 0 || s == "" {
		return ""
	}
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) <= maxLines {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[len(lines)-maxLines:], "\n")
}
