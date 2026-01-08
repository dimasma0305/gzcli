package uploadserver

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path"
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
`

const sampleChallengeProvideDist = sampleChallengeYAML + `
provide: "./dist"
`

func TestProcessUpload_Success(t *testing.T) {
	const (
		event    = "TestEvent"
		category = "Web"
	)

	workspace := setupWorkspace(t, event, category)
	archive := buildChallengeArchive(t, buildChallengeArchiveConfig{
		ChallengeYAML: sampleChallengeYAML,
		IncludeSolver: true,
		SolverReadme:  "initial solver",
		SrcFiles: map[string]string{
			"README.md": "source file",
		},
	})

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
	archive := buildChallengeArchive(t, buildChallengeArchiveConfig{
		IncludeSolver: false,
	})

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

	archiveV1 := buildChallengeArchive(t, buildChallengeArchiveConfig{
		IncludeSolver: true,
		SolverReadme:  "v1",
	})

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

	archiveV2 := buildChallengeArchive(t, buildChallengeArchiveConfig{
		IncludeSolver: true,
		SolverReadme:  "v2",
	})

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

func TestProcessUpload_InvalidRootContents(t *testing.T) {
	const (
		event    = "EventInvalid"
		category = "Web"
	)

	_ = setupWorkspace(t, event, category)
	archive := buildChallengeArchive(t, buildChallengeArchiveConfig{
		IncludeSolver: true,
		ExtraRootFiles: map[string]string{
			"README.txt": "unexpected",
		},
	})

	file, err := os.Open(filepath.Clean(archive)) // #nosec G304
	if err != nil {
		t.Fatalf("failed to open archive: %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })

	srv := newTestServer(t)

	err = srv.processUpload(context.Background(), event, category, file, "invalid.zip")
	if !errors.Is(err, errInvalidRootContents) {
		t.Fatalf("expected errInvalidRootContents, got %v", err)
	}
}

func TestProcessUpload_EmptyDistProvided(t *testing.T) {
	const (
		event    = "EventDist"
		category = "Web"
	)

	_ = setupWorkspace(t, event, category)
	archive := buildChallengeArchive(t, buildChallengeArchiveConfig{
		ChallengeYAML: sampleChallengeProvideDist,
		IncludeSolver: true,
	})

	file, err := os.Open(filepath.Clean(archive)) // #nosec G304
	if err != nil {
		t.Fatalf("failed to open archive: %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })

	srv := newTestServer(t)

	err = srv.processUpload(context.Background(), event, category, file, "emptydist.zip")
	if !errors.Is(err, errEmptyDistProvided) {
		t.Fatalf("expected errEmptyDistProvided, got %v", err)
	}
}

func TestProcessUpload_DefaultChallengeYAML(t *testing.T) {
	const (
		event    = "EventTemplate"
		category = "Web"
	)

	templateContent, err := fs.ReadFile(templateFS, path.Join(templateStaticAttachmentPath, "challenge.yml"))
	if err != nil {
		t.Fatalf("failed to read template: %v", err)
	}

	_ = setupWorkspace(t, event, category)
	archive := buildChallengeArchive(t, buildChallengeArchiveConfig{
		ChallengeYAML: string(templateContent),
		IncludeSolver: true,
		DistFiles: map[string]string{
			"placeholder.txt": "content",
		},
	})

	file, err := os.Open(filepath.Clean(archive)) // #nosec G304
	if err != nil {
		t.Fatalf("failed to open archive: %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })

	srv := newTestServer(t)

	err = srv.processUpload(context.Background(), event, category, file, "template.zip")
	if !errors.Is(err, errChallengeTemplateUnchanged) {
		t.Fatalf("expected errChallengeTemplateUnchanged, got %v", err)
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

type buildChallengeArchiveConfig struct {
	ChallengeYAML  string
	IncludeSolver  bool
	SolverReadme   string
	SolverFiles    map[string]string
	DistFiles      map[string]string
	SrcFiles       map[string]string
	ExtraRootFiles map[string]string
	ExtraRootDirs  []string
}

func buildChallengeArchive(t *testing.T, cfg buildChallengeArchiveConfig) string {
	t.Helper()

	root := t.TempDir()
	challengeDir := filepath.Join(root, "challenge")

	if err := os.MkdirAll(challengeDir, 0o750); err != nil {
		t.Fatalf("failed to create challenge directory: %v", err)
	}

	challengeContent := cfg.ChallengeYAML
	if challengeContent == "" {
		challengeContent = sampleChallengeYAML
	}
	if err := os.WriteFile(filepath.Join(challengeDir, "challenge.yml"), []byte(challengeContent), 0o600); err != nil {
		t.Fatalf("failed to write challenge.yml: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(challengeDir, "dist"), 0o750); err != nil {
		t.Fatalf("failed to create dist directory: %v", err)
	}
	distDir := filepath.Join(challengeDir, "dist")
	writeFiles(t, distDir, cfg.DistFiles)
	if len(cfg.DistFiles) == 0 {
		ensurePlaceholderFile(t, distDir)
	}

	if err := os.MkdirAll(filepath.Join(challengeDir, "src"), 0o750); err != nil {
		t.Fatalf("failed to create src directory: %v", err)
	}
	srcDir := filepath.Join(challengeDir, "src")
	writeFiles(t, srcDir, cfg.SrcFiles)
	if len(cfg.SrcFiles) == 0 {
		ensurePlaceholderFile(t, srcDir)
	}

	if cfg.IncludeSolver {
		solverDir := filepath.Join(challengeDir, "solver")
		if err := os.MkdirAll(solverDir, 0o750); err != nil {
			t.Fatalf("failed to create solver directory: %v", err)
		}

		if len(cfg.SolverFiles) == 0 {
			readme := cfg.SolverReadme
			if readme == "" {
				readme = "default solver"
			}
			cfg.SolverFiles = map[string]string{"README.md": readme}
		}
		writeFiles(t, solverDir, cfg.SolverFiles)
	}

	for name, content := range cfg.ExtraRootFiles {
		if err := os.WriteFile(filepath.Join(challengeDir, name), []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write extra root file %s: %v", name, err)
		}
	}

	for _, dir := range cfg.ExtraRootDirs {
		if err := os.MkdirAll(filepath.Join(challengeDir, dir), 0o750); err != nil {
			t.Fatalf("failed to create extra root directory %s: %v", dir, err)
		}
	}

	archivePath := filepath.Join(root, "challenge.zip")
	if err := fileutil.ZipSource(challengeDir, archivePath); err != nil {
		t.Fatalf("failed to zip challenge: %v", err)
	}

	return archivePath
}

func writeFiles(t *testing.T, base string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		target := filepath.Join(base, name)
		if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
			t.Fatalf("failed to create directory for %s: %v", name, err)
		}
		if err := os.WriteFile(target, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
}

func ensurePlaceholderFile(t *testing.T, dir string) {
	t.Helper()
	placeholder := filepath.Join(dir, ".gitkeep")
	if err := os.WriteFile(placeholder, []byte{}, 0o600); err != nil {
		t.Fatalf("failed to write placeholder file in %s: %v", dir, err)
	}
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
