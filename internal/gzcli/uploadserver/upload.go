package uploadserver

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/challenge"
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/fileutil"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	challengeFileRegex            = regexp.MustCompile(`(?i)^challenge\.(ya?ml)$`)
	errInvalidCategory            = errors.New("invalid category")
	errNoChallengeYML             = errors.New("challenge.yml not found in archive")
	errMultiChallenge             = errors.New("multiple challenge.yml files found in archive")
	errMissingSolver              = errors.New("solver directory missing from challenge package")
	errInvalidRootContents        = errors.New("challenge root contains unexpected entries")
	errMissingDist                = errors.New("dist directory missing from challenge package")
	errMissingSrc                 = errors.New("src directory missing from challenge package")
	errEmptyDistProvided          = errors.New("dist directory is empty while challenge.yml provides it")
	errChallengeTemplateUnchanged = errors.New("challenge.yml matches a default template")

	errArchiveTooManyEntries = errors.New("archive contains too many entries")
	errArchiveEntryTooLarge  = errors.New("archive entry exceeds maximum size")
	errArchiveTooLarge       = errors.New("archive uncompressed size exceeds limit")
)

const (
	maxExtractedEntries = 4096
	maxEntryBytes       = 40 << 20  // 40 MiB per file
	maxExtractedBytes   = 100 << 20 // 100 MiB total
)

// processUpload handles parsing, validating, and installing the uploaded challenge archive.
func (s *server) processUpload(ctx context.Context, event, category string, file multipart.File, originalName string) error {
	event = strings.TrimSpace(event)
	category = strings.TrimSpace(category)

	if event == "" {
		return errors.New("event selection is required")
	}
	if category == "" {
		return errors.New("category selection is required")
	}
	if !isValidCategory(category) {
		return fmt.Errorf("%w: %s", errInvalidCategory, category)
	}

	eventPath, err := config.GetEventPath(event)
	if err != nil {
		return fmt.Errorf("invalid event %q: %w", event, err)
	}

	tempRoot, err := os.MkdirTemp("", "gzcli-upload-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempRoot)
	}()

	archivePath := filepath.Join(tempRoot, sanitizeFileName(originalName))
	if err := writeTempArchive(file, archivePath); err != nil {
		return err
	}

	extractDir := filepath.Join(tempRoot, "extracted")
	if err := extractArchive(ctx, archivePath, extractDir); err != nil {
		return err
	}

	challengeYMLPath, err := locateChallengeYML(extractDir)
	if err != nil {
		return err
	}

	challengeRoot := filepath.Dir(challengeYMLPath)
	if err := validateChallengeRoot(challengeRoot, challengeYMLPath); err != nil {
		return err
	}

	var chall config.ChallengeYaml
	if err := fileutil.ParseYamlFromFile(challengeYMLPath, &chall); err != nil {
		return fmt.Errorf("failed to parse challenge.yml: %w", err)
	}

	if err := ensureChallengeCustomized(chall); err != nil {
		return err
	}

	if err := challenge.IsGoodChallenge(chall); err != nil {
		return err
	}

	if err := ensureProvideDistConsistency(challengeRoot, chall); err != nil {
		return err
	}

	destCategoryDir := filepath.Join(eventPath, category)
	if err := os.MkdirAll(destCategoryDir, 0750); err != nil {
		return fmt.Errorf("failed to ensure category directory: %w", err)
	}

	finalName := sanitizeChallengeDirName(chall.Name)
	if finalName == "" {
		finalName = filepath.Base(challengeRoot)
	}

	destination := filepath.Join(destCategoryDir, finalName)
	if err := os.RemoveAll(destination); err != nil {
		return fmt.Errorf("failed to replace existing challenge: %w", err)
	}

	if err := copyDir(challengeRoot, destination); err != nil {
		return fmt.Errorf("failed to install challenge: %w", err)
	}

	log.Info("Installed challenge %q into %s/%s", chall.Name, event, category)
	return nil
}

func writeTempArchive(src multipart.File, dst string) error {
	if err := srcToFile(src, dst); err != nil {
		return fmt.Errorf("failed to persist uploaded archive: %w", err)
	}
	return nil
}

func srcToFile(src multipart.File, dst string) error {
	//nolint:gosec // destination path lives in temp directory managed by the server
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, src); err != nil {
		return err
	}

	return out.Sync()
}

func extractArchive(ctx context.Context, src, dst string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer func() { _ = reader.Close() }()

	if err := os.MkdirAll(dst, 0750); err != nil {
		return fmt.Errorf("failed to create extraction directory: %w", err)
	}

	limiter := newExtractionLimiter(maxExtractedEntries, maxEntryBytes, maxExtractedBytes)

	for _, file := range reader.File {
		if err := ctx.Err(); err != nil {
			return err
		}

		if strings.HasPrefix(file.Name, "__MACOSX/") {
			continue
		}

		cleanName := filepath.Clean(file.Name)
		if strings.Contains(cleanName, "..") {
			return fmt.Errorf("archive contains invalid path %q", file.Name)
		}

		targetPath := filepath.Join(dst, cleanName)
		rel, err := filepath.Rel(dst, targetPath)
		if err != nil {
			return fmt.Errorf("failed to resolve archive path %q: %w", file.Name, err)
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("archive entry escapes extraction directory: %q", file.Name)
		}

		if err := limiter.registerEntry(); err != nil {
			return fmt.Errorf("archive entry %q rejected: %w", file.Name, err)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, ensureDirWritable(fileModeOrDefault(file.FileInfo(), 0750))); err != nil {
				return fmt.Errorf("failed to create directory %q: %w", targetPath, err)
			}
			continue
		}

		if file.UncompressedSize64 > limiter.maxEntryBytes {
			return fmt.Errorf("archive entry %q too large: %w", file.Name, errArchiveEntryTooLarge)
		}

		if err := writeZipEntry(file, targetPath, limiter); err != nil {
			return err
		}
	}

	return nil
}

func writeZipEntry(file *zip.File, target string, limiter *extractionLimiter) error {
	if err := os.MkdirAll(filepath.Dir(target), 0750); err != nil {
		return fmt.Errorf("failed to create parent directory for %q: %w", target, err)
	}

	//nolint:gosec // zip entries are extracted into dedicated temp dir
	dstFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileModeOrDefault(file.FileInfo(), 0644))
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", target, err)
	}
	defer func() { _ = dstFile.Close() }()

	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open archive entry %q: %w", file.Name, err)
	}
	defer func() { _ = rc.Close() }()

	written, err := copyWithArchiveLimits(dstFile, rc, limiter)
	if err != nil {
		_ = dstFile.Close()
		_ = os.Remove(target)
		switch {
		case errors.Is(err, errArchiveEntryTooLarge):
			return fmt.Errorf("archive entry %q too large: %w", file.Name, err)
		case errors.Is(err, errArchiveTooLarge):
			return fmt.Errorf("archive exceeds allowed size while extracting %q: %w", file.Name, err)
		default:
			return fmt.Errorf("failed to write archive entry %q: %w", file.Name, err)
		}
	}

	if err := limiter.commitBytes(written); err != nil {
		_ = os.Remove(target)
		return fmt.Errorf("archive exceeds allowed size after extracting %q: %w", file.Name, err)
	}

	return nil
}

type extractionLimiter struct {
	maxEntries     int
	maxEntryBytes  uint64
	maxTotalBytes  uint64
	entries        int
	pendingTotal   uint64
	committedTotal uint64
	entryBytes     uint64
}

func newExtractionLimiter(maxEntries int, maxEntryBytes, maxTotalBytes uint64) *extractionLimiter {
	return &extractionLimiter{
		maxEntries:    maxEntries,
		maxEntryBytes: maxEntryBytes,
		maxTotalBytes: maxTotalBytes,
	}
}

func (l *extractionLimiter) registerEntry() error {
	if l.maxEntries <= 0 {
		return nil
	}
	if l.entries+1 > l.maxEntries {
		return errArchiveTooManyEntries
	}
	l.entries++
	l.entryBytes = 0
	return nil
}

func (l *extractionLimiter) recordBytes(n uint64) error {
	if n == 0 {
		return nil
	}
	if l.entryBytes+n > l.maxEntryBytes {
		return errArchiveEntryTooLarge
	}
	if l.pendingTotal+n > l.maxTotalBytes {
		return errArchiveTooLarge
	}
	l.entryBytes += n
	l.pendingTotal += n
	return nil
}

func (l *extractionLimiter) commitBytes(n uint64) error {
	if n == 0 {
		return nil
	}
	if l.pendingTotal != l.committedTotal+n {
		// pendingTotal already includes this entry; ensure it aligns.
		l.committedTotal += n
	} else {
		l.committedTotal = l.pendingTotal
	}
	if l.committedTotal > l.maxTotalBytes {
		return errArchiveTooLarge
	}
	l.pendingTotal = l.committedTotal
	return nil
}

func copyWithArchiveLimits(dst io.Writer, src io.Reader, limiter *extractionLimiter) (uint64, error) {
	var written uint64
	buf := make([]byte, 32*1024)

	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			if n > len(buf) {
				return written, fmt.Errorf("invalid read length %d exceeds buffer capacity", n)
			}
			chunkLen := uint64(len(buf[:n]))
			if err := limiter.recordBytes(chunkLen); err != nil {
				return written, err
			}

			writtenChunk, err := dst.Write(buf[:n])
			if err != nil {
				return written, err
			}
			if writtenChunk != n {
				return written, io.ErrShortWrite
			}

			written += chunkLen
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return written, readErr
		}
	}

	return written, nil
}

func fileModeOrDefault(info fs.FileInfo, fallback fs.FileMode) fs.FileMode {
	if info == nil {
		return fallback
	}
	mode := info.Mode()
	if mode == 0 {
		return fallback
	}
	return mode
}

func ensureDirWritable(mode fs.FileMode) fs.FileMode {
	if mode == 0 {
		mode = 0750
	}
	return (mode | 0700) & fs.ModePerm
}

func validateChallengeRoot(root, challengeYMLPath string) error {
	if filepath.Base(challengeYMLPath) != "challenge.yml" {
		return fmt.Errorf("challenge definition file must be named challenge.yml")
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("failed to inspect challenge root: %w", err)
	}

	var (
		hasChallenge bool
		hasDist      bool
		hasSolver    bool
		hasSrc       bool
	)

	for _, entry := range entries {
		name := entry.Name()
		switch name {
		case "challenge.yml":
			if entry.IsDir() {
				return fmt.Errorf("challenge.yml must be a file, not a directory")
			}
			hasChallenge = true
		case "dist":
			if !entry.IsDir() {
				return fmt.Errorf("dist exists but is not a directory")
			}
			hasDist = true
		case "solver":
			if !entry.IsDir() {
				return fmt.Errorf("solver exists but is not a directory")
			}
			hasSolver = true
		case "src":
			if !entry.IsDir() {
				return fmt.Errorf("src exists but is not a directory")
			}
			hasSrc = true
		default:
			return fmt.Errorf("%w: %s", errInvalidRootContents, name)
		}
	}

	if !hasChallenge {
		return errNoChallengeYML
	}
	if !hasDist {
		return errMissingDist
	}
	if !hasSrc {
		return errMissingSrc
	}
	if !hasSolver {
		return errMissingSolver
	}

	if err := ensureSolverDir(root); err != nil {
		return err
	}

	return nil
}

func ensureChallengeCustomized(chall config.ChallengeYaml) error {
	for _, tpl := range challengeTemplates {
		templatePath := path.Join(tpl.SourcePath, "challenge.yml")
		content, err := fs.ReadFile(templateFS, templatePath)
		if err != nil {
			return fmt.Errorf("failed to load challenge template %s: %w", tpl.Slug, err)
		}

		var templateChall config.ChallengeYaml
		if err := fileutil.ParseYamlFromBytes(content, &templateChall); err != nil {
			return fmt.Errorf("failed to parse embedded template %s: %w", tpl.Slug, err)
		}

		if reflect.DeepEqual(templateChall, chall) {
			return errChallengeTemplateUnchanged
		}
	}

	return nil
}

func ensureProvideDistConsistency(root string, chall config.ChallengeYaml) error {
	if chall.Provide == nil {
		return nil
	}

	if !referencesDist(*chall.Provide) {
		return nil
	}

	distPath := filepath.Join(root, "dist")
	defaultState, err := isDefaultDist(distPath)
	if err != nil {
		return err
	}

	if defaultState {
		return errEmptyDistProvided
	}

	return nil
}

func referencesDist(provide string) bool {
	p := strings.TrimSpace(provide)
	if p == "" {
		return false
	}

	clean := filepath.Clean(p)
	clean = filepath.ToSlash(clean)
	clean = strings.TrimPrefix(clean, "./")
	clean = strings.Trim(clean, "/")

	return clean == "dist"
}

func isDefaultDist(distPath string) (bool, error) {
	entries, err := os.ReadDir(distPath)
	if err != nil {
		return false, fmt.Errorf("failed to inspect dist directory: %w", err)
	}

	if len(entries) == 0 {
		return true, nil
	}

	if len(entries) == 1 {
		entry := entries[0]
		if entry.Name() == ".gitkeep" && !entry.IsDir() {
			return true, nil
		}
	}

	return false, nil
}

func locateChallengeYML(root string) (string, error) {
	var matches []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if challengeFileRegex.MatchString(d.Name()) {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to inspect extracted archive: %w", err)
	}

	switch len(matches) {
	case 0:
		return "", errNoChallengeYML
	case 1:
		return matches[0], nil
	default:
		return "", errMultiChallenge
	}
}

func ensureSolverDir(root string) error {
	solverDir := filepath.Join(root, "solver")
	info, err := os.Stat(solverDir)
	if err != nil {
		if os.IsNotExist(err) {
			return errMissingSolver
		}
		return fmt.Errorf("failed to inspect solver directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("solver exists but is not a directory")
	}

	return nil
}

func isValidCategory(category string) bool {
	for _, cat := range config.CHALLENGE_CATEGORY {
		if cat == category {
			return true
		}
	}
	return false
}

func sanitizeFileName(name string) string {
	if name == "" {
		return "challenge.zip"
	}
	base := fileutil.NormalizeFileName(strings.TrimSuffix(name, filepath.Ext(name)))
	if base == "" {
		base = "challenge"
	}
	return base + ".zip"
}

func sanitizeChallengeDirName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	return fileutil.NormalizeFileName(name)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, rel)

		if info.IsDir() {
			if err := os.MkdirAll(target, ensureDirWritable(info.Mode())); err != nil {
				return fmt.Errorf("failed to create directory %q: %w", target, err)
			}
			return nil
		}

		if err := copyFile(path, target, info.Mode()); err != nil {
			return err
		}

		return nil
	})
}

func copyFile(src, dst string, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return fmt.Errorf("failed to create parent directory for %q: %w", dst, err)
	}

	//nolint:gosec // paths come from validated challenge directory
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %q: %w", src, err)
	}
	defer func() { _ = in.Close() }()

	//nolint:gosec // destination resides inside workspace event directory
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create destination file %q: %w", dst, err)
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("failed to copy file %q: %w", src, err)
	}

	return out.Sync()
}

// writeTemplateArchive packages the embedded challenge template into a ZIP archive.
func writeTemplateArchive(w io.Writer, tpl challengeTemplate) error {
	zw := zip.NewWriter(w)
	timestamp := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	hasSolver := false

	err := fs.WalkDir(templateFS, tpl.SourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == tpl.SourcePath {
			return nil
		}

		rel, err := filepath.Rel(tpl.SourcePath, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		zipPath := filepath.ToSlash(filepath.Join(tpl.Slug, rel))

		if d.IsDir() {
			if rel == "solver" {
				hasSolver = true
			}

			header := &zip.FileHeader{
				Name:     zipPath + "/",
				Method:   zip.Deflate,
				Modified: timestamp,
			}
			header.SetMode(0755)
			if _, err := zw.CreateHeader(header); err != nil {
				return err
			}
			return nil
		} else if strings.HasPrefix(rel, "solver/") {
			hasSolver = true
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		header := &zip.FileHeader{
			Name:     zipPath,
			Method:   zip.Deflate,
			Modified: timestamp,
		}
		header.SetMode(info.Mode())

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		content, err := templateFS.Open(path)
		if err != nil {
			return err
		}

		if _, err := io.Copy(writer, content); err != nil {
			_ = content.Close()
			return err
		}
		if err := content.Close(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		_ = zw.Close()
		return fmt.Errorf("failed to package template %s: %w", tpl.Slug, err)
	}

	if !hasSolver {
		header := &zip.FileHeader{
			Name:     filepath.ToSlash(filepath.Join(tpl.Slug, "solver")) + "/",
			Method:   zip.Deflate,
			Modified: timestamp,
		}
		header.SetMode(0755)
		if _, err := zw.CreateHeader(header); err != nil {
			_ = zw.Close()
			return fmt.Errorf("failed to add solver directory: %w", err)
		}
	}

	return zw.Close()
}
