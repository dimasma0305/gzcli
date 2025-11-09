package uploadserver

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/fileutil"
)

const sampleChallengeYAML = `name: "Upload Sample"
author: "tester"
description: "simple challenge"
type: "StaticAttachment"
value: 50
flags:
  - "flag{TEST}"
scripts:
  start: echo start
  stop: echo stop
`

func TestProcessUpload_Success(t *testing.T) {
	const (
		event    = "TestEvent"
		category = "Web"
	)

	workspace := setupWorkspace(t, event, category)
	archive := buildChallengeArchive(t, true, "initial solver")

	file, err := os.Open(filepath.Clean(archive)) // #nosec G304 -- archive resides in a controlled temp directory
	if err != nil {
		t.Fatalf("failed to open archive: %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })

	srv := newTestServer(t)

	if err := srv.processUpload(context.Background(), event, category, file, "challenge.zip"); err != nil {
		t.Fatalf("processUpload returned error: %v", err)
	}

	dest := filepath.Join(workspace, "events", event, category, "uploadsample")
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("expected challenge directory %s to exist: %v", dest, err)
	}

	if _, err := os.Stat(filepath.Join(dest, "challenge.yml")); err != nil {
		t.Fatalf("expected challenge.yml in destination: %v", err)
	}

	if info, err := os.Stat(filepath.Join(dest, "solver")); err != nil || !info.IsDir() {
		t.Fatalf("expected solver directory in destination: %v (info: %v)", err, info)
	}
}

func TestProcessUpload_MissingChallengeYML(t *testing.T) {
	const (
		event    = "EventOne"
		category = "Web"
	)

	_ = setupWorkspace(t, event, category)
	archive := buildArchiveWithoutChallengeYML(t)

	file, err := os.Open(filepath.Clean(archive)) // #nosec G304 -- archive resides in a controlled temp directory
	if err != nil {
		t.Fatalf("failed to open archive: %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })

	srv := newTestServer(t)

	err = srv.processUpload(context.Background(), event, category, file, "missing.zip")
	if !errors.Is(err, errNoChallengeYML) {
		t.Fatalf("expected errNoChallengeYML, got %v", err)
	}
}

func TestProcessUpload_MissingSolver(t *testing.T) {
	const (
		event    = "EventTwo"
		category = "Web"
	)

	_ = setupWorkspace(t, event, category)
	archive := buildChallengeArchive(t, false)

	file, err := os.Open(filepath.Clean(archive)) // #nosec G304 -- archive resides in a controlled temp directory
	if err != nil {
		t.Fatalf("failed to open archive: %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })

	srv := newTestServer(t)

	err = srv.processUpload(context.Background(), event, category, file, "nosolver.zip")
	if !errors.Is(err, errMissingSolver) {
		t.Fatalf("expected errMissingSolver, got %v", err)
	}
}

func TestProcessUpload_ReplacesExistingChallenge(t *testing.T) {
	const (
		event    = "EventReplace"
		category = "Web"
	)

	workspace := setupWorkspace(t, event, category)

	archiveV1 := buildChallengeArchive(t, true, "v1")

	file1, err := os.Open(filepath.Clean(archiveV1)) // #nosec G304 -- archive resides in a controlled temp directory
	if err != nil {
		t.Fatalf("failed to open archive v1: %v", err)
	}
	t.Cleanup(func() { _ = file1.Close() })

	srv := newTestServer(t)

	if err := srv.processUpload(context.Background(), event, category, file1, "challenge-v1.zip"); err != nil {
		t.Fatalf("processUpload v1 error: %v", err)
	}
	_ = file1.Close()

	archiveV2 := buildChallengeArchive(t, true, "v2")

	file2, err := os.Open(filepath.Clean(archiveV2)) // #nosec G304 -- archive resides in a controlled temp directory
	if err != nil {
		t.Fatalf("failed to open archive v2: %v", err)
	}
	t.Cleanup(func() { _ = file2.Close() })

	if err := srv.processUpload(context.Background(), event, category, file2, "challenge-v2.zip"); err != nil {
		t.Fatalf("processUpload v2 error: %v", err)
	}

	dest := filepath.Join(workspace, "events", event, category, "uploadsample")

	solverReadme := filepath.Clean(filepath.Join(dest, "solver", "README.md"))
	content, err := os.ReadFile(solverReadme) // #nosec G304 -- path points to controlled workspace directory
	if err != nil {
		t.Fatalf("failed reading updated solver: %v", err)
	}
	if got := strings.TrimSpace(string(content)); got != "v2" {
		t.Fatalf("expected updated solver content 'v2', got %q", got)
	}

	if _, err := os.Stat(filepath.Join(dest, "src", "old.txt")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected old.txt to be removed, err=%v", err)
	}
}

func TestWriteTemplateArchive(t *testing.T) {
	for _, tpl := range challengeTemplates {
		tpl := tpl
		t.Run(tpl.Slug, func(t *testing.T) {
			var buf bytes.Buffer
			if err := writeTemplateArchive(&buf, tpl); err != nil {
				t.Fatalf("writeTemplateArchive error: %v", err)
			}

			readerAt := bytes.NewReader(buf.Bytes())
			zr, err := zip.NewReader(readerAt, int64(buf.Len()))
			if err != nil {
				t.Fatalf("failed to open generated archive: %v", err)
			}

			wantChallenge := tpl.Slug + "/challenge.yml"
			wantSolverDir := tpl.Slug + "/solver/"
			found := false
			hasSolver := false
			for _, file := range zr.File {
				if file.Name == wantChallenge {
					found = true
				}
				if file.Name == wantSolverDir {
					hasSolver = true
				}
			}
			if !found {
				t.Fatalf("expected archive to include %s", wantChallenge)
			}
			if !hasSolver {
				t.Fatalf("expected archive to include %s", wantSolverDir)
			}
		})
	}
}

func newTestServer(t *testing.T) *server {
	t.Helper()
	srv, err := newServer(Options{
		Host: "localhost",
		Port: 8090,
	})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	return srv
}

//nolint:unparam // helper accepts category for future tests even if current cases reuse the same value
func setupWorkspace(t *testing.T, event, category string) string {
	t.Helper()

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	workspace := t.TempDir()

	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	if err := os.MkdirAll(filepath.Join("events", event, category), 0o750); err != nil {
		t.Fatalf("failed to create events directory: %v", err)
	}

	return workspace
}

func buildChallengeArchive(t *testing.T, includeSolver bool, solverContent ...string) string {
	t.Helper()

	root := t.TempDir()
	challengeDir := filepath.Join(root, "challenge")

	if err := os.MkdirAll(challengeDir, 0o750); err != nil {
		t.Fatalf("failed to create challenge directory: %v", err)
	}

	if err := os.WriteFile(filepath.Join(challengeDir, "challenge.yml"), []byte(sampleChallengeYAML), 0o600); err != nil {
		t.Fatalf("failed to write challenge.yml: %v", err)
	}

	if includeSolver {
		content := "default solver"
		if len(solverContent) > 0 && solverContent[0] != "" {
			content = solverContent[0]
		}
		solverDir := filepath.Join(challengeDir, "solver")
		if err := os.MkdirAll(solverDir, 0o750); err != nil {
			t.Fatalf("failed to create solver directory: %v", err)
		}
		solverFile := filepath.Join(solverDir, "README.md")
		if err := os.WriteFile(solverFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write solver file: %v", err)
		}
	}

	archivePath := filepath.Join(root, "challenge.zip")
	if err := fileutil.ZipSource(challengeDir, archivePath); err != nil {
		t.Fatalf("failed to zip challenge: %v", err)
	}

	return archivePath
}

func buildArchiveWithoutChallengeYML(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	challengeDir := filepath.Join(root, "challenge")

	if err := os.MkdirAll(filepath.Join(challengeDir, "solver"), 0o750); err != nil {
		t.Fatalf("failed to create solver directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(challengeDir, "solver", "README.md"), []byte("content"), 0o600); err != nil {
		t.Fatalf("failed to write solver file: %v", err)
	}

	if err := os.WriteFile(filepath.Join(challengeDir, "placeholder.txt"), []byte("no challenge yaml"), 0o600); err != nil {
		t.Fatalf("failed to write placeholder: %v", err)
	}

	archivePath := filepath.Join(root, "challenge.zip")
	if err := fileutil.ZipSource(challengeDir, archivePath); err != nil {
		t.Fatalf("failed to zip archive: %v", err)
	}

	return archivePath
}
