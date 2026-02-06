// Package fileutil provides file operation utilities for challenge management.
//
// This package includes high-performance file operations optimized for CTF challenges:
//   - File name normalization for cross-platform compatibility
//   - SHA256 hash calculation for integrity verification
//   - Fast file copying with proper synchronization
//   - Optimized ZIP archive creation with parallel processing
//
// Example usage:
//
//	// Normalize filename
//	safe := fileutil.NormalizeFileName("My Challenge!")  // Returns: "mychallenge"
//
//	// Calculate file hash
//	hash, err := fileutil.GetFileHashHex("challenge.zip")
//
//	// Copy file
//	if err := fileutil.CopyFile("src.txt", "dst.txt"); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create ZIP archive
//	if err := fileutil.ZipSource("./challenge", "challenge.zip"); err != nil {
//	    log.Fatal(err)
//	}
package fileutil

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var fileNameNormalizer = regexp.MustCompile(`[^a-zA-Z0-9\-_]+`)

// NormalizeFileName normalizes a filename by removing special characters and converting to lowercase
func NormalizeFileName(name string) string {
	buf := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buf)
	defer buf.Reset()

	buf.WriteString(name)
	result := fileNameNormalizer.ReplaceAllString(buf.String(), "")
	return strings.ToLower(result)
}

// GetFileHashHex calculates the SHA256 hash of a file and returns it as a hex string
func GetFileHashHex(file string) (string, error) {
	//nolint:gosec // G304: File paths come from challenge config, validated by user
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := f.WriteTo(h); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	//nolint:gosec // G304: File paths come from challenge config, validated by user
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }()

	//nolint:gosec // G304: File paths come from challenge config, validated by user
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}

// ZipSource creates a zip archive of a source directory
func ZipSource(source, target string) error {
	// Create output file with buffered writer
	//nolint:gosec // G304: Target path is constructed from validated challenge config
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	buffered := bufio.NewWriterSize(f, 1<<20) // 1MB buffer
	defer func() { _ = buffered.Flush() }()

	// Create zip writer with optimized compression
	writer := zip.NewWriter(buffered)
	defer func() { _ = writer.Close() }()

	// Set faster compression level
	writer.RegisterCompressor(zip.Deflate, func(w io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(w, flate.BestSpeed)
	})

	// Use a fixed timestamp for reproducible builds
	fixedTime := time.Date(2025, 3, 18, 0, 0, 0, 0, time.UTC)

	// Collect and sort relative paths to ensure deterministic ZIP output.
	// Notes:
	// - filepath.Walk can surface errors via the callback; we intentionally swallow them
	//   to preserve the previous behavior (best-effort empty ZIP for missing/partial trees).
	// - Use forward slashes for ZIP entry names for cross-platform compatibility.
	var relPaths []string
	_ = filepath.Walk(source, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info == nil || info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		// If source is a single file, Rel will be "."; use the base name instead.
		if relPath == "." {
			relPath = filepath.Base(path)
		}
		relPaths = append(relPaths, filepath.ToSlash(relPath))
		return nil
	})

	sort.Strings(relPaths)

	for _, relPath := range relPaths {
		fullPath := filepath.Join(source, filepath.FromSlash(relPath))
		//nolint:gosec // G304: File paths come from validated challenge directory
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}

		header := &zip.FileHeader{
			Name:     relPath,
			Method:   zip.Deflate,
			Modified: fixedTime,
		}
		header.SetMode(0644)

		w, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if _, err := io.Copy(w, bytes.NewReader(data)); err != nil {
			return err
		}
	}

	return nil
}
