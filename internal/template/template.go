// Package template provides utilities for processing and applying templates
package template

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/dimasma0305/gzcli/internal/utils"
)

var (
	//go:embed all:templates
	// File is the embedded filesystem containing all templates
	File embed.FS
)

// ==================================================
// Filesystem Abstraction
// ==================================================

type fileSystem interface {
	ReadFile(string) ([]byte, error)
	ReadDir(string) ([]fs.DirEntry, error)
	Open(string) (fs.File, error)
	Stat(string) (fs.FileInfo, error)
}

type embeddedFS struct{ fs embed.FS }

func (e embeddedFS) ReadFile(name string) ([]byte, error)       { return e.fs.ReadFile(name) }
func (e embeddedFS) ReadDir(name string) ([]fs.DirEntry, error) { return e.fs.ReadDir(name) }
func (e embeddedFS) Open(name string) (fs.File, error)          { return e.fs.Open(name) }
func (e embeddedFS) Stat(name string) (fs.FileInfo, error) {
	f, err := e.fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	return f.Stat()
}

type osFS struct{}

//nolint:gosec // G304: Template paths come from embedded FS or user templates directory
func (osFS) ReadFile(name string) ([]byte, error)       { return os.ReadFile(name) }
func (osFS) ReadDir(name string) ([]fs.DirEntry, error) { return os.ReadDir(name) }

//nolint:gosec // G304: Template paths come from embedded FS or user templates directory
func (osFS) Open(name string) (fs.File, error)     { return os.Open(name) }
func (osFS) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }

// ==================================================
// Main Template Processing
// ==================================================

// TemplateFSToDestination processes a template from the embedded filesystem
//
//nolint:revive // Function name kept for API consistency
func TemplateFSToDestination(file string, info interface{}, destination string) []error {
	return processWithFS(embeddedFS{File}, file, info, destination)
}

// TemplateToDestination processes a template from the operating system filesystem
//
//nolint:revive // Function name kept for API consistency
func TemplateToDestination(src string, info interface{}, destination string) []error {
	return processWithFS(osFS{}, src, info, destination)
}

func processWithFS(fsys fileSystem, src string, info interface{}, dest string) []error {
	fi, err := fsys.Stat(src)
	if err != nil {
		return []error{fmt.Errorf("failed to stat source %q: %w", src, err)}
	}

	if fi.IsDir() {
		return processDir(fsys, src, info, dest)
	}
	return processFile(fsys, src, info, dest)
}

func processDir(fsys fileSystem, dir string, info interface{}, destination string) []error {
	var errs []error

	if err := os.MkdirAll(destination, 0750); err != nil {
		errs = append(errs, fmt.Errorf("failed to create directory %q: %w", destination, err))
		return errs
	}

	entries, err := fsys.ReadDir(dir)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to read directory %q: %w", dir, err))
		return errs
	}

	for _, entry := range entries {
		srcPath := filepath.Join(dir, entry.Name())
		destPath := filepath.Join(destination, entry.Name())
		if subErrs := processWithFS(fsys, srcPath, info, destPath); len(subErrs) > 0 {
			errs = append(errs, subErrs...)
		}
	}
	return errs
}

func processFile(fsys fileSystem, file string, info interface{}, destination string) []error {
	var errs []error

	file = utils.NormalizePath(file)
	destination = strings.ReplaceAll(destination, "{{replaceit}}", "")

	content, err := processTemplate(fsys, file, info)
	if err != nil {
		// Template processing failed, fall back to raw copy
		errs = append(errs, fmt.Errorf("template processing error for %q: %w", file, err))

		rawFile, openErr := fsys.Open(file)
		if openErr != nil {
			errs = append(errs, fmt.Errorf("failed to open raw file %q: %w", file, openErr))
			return errs
		}
		defer func() {
			_ = rawFile.Close()
		}()
		content = rawFile
	}

	if err := writeContent(destination, content); err != nil {
		errs = append(errs, fmt.Errorf("write error for %q: %w", destination, err))
	}

	return errs
}

func processTemplate(fsys fileSystem, file string, info interface{}) (io.Reader, error) {
	data, err := fsys.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %q: %w", file, err)
	}

	tmpl, err := template.New(filepath.Base(file)).Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("template parse error: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, info); err != nil {
		return nil, fmt.Errorf("template execute error: %w", err)
	}

	return strings.NewReader(buf.String()), nil
}

// ==================================================
// Shared Function
// ==================================================

//nolint:gosec // G304: Destination path is constructed from validated template config
func writeContent(destination string, content io.Reader) error {
	destFile, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("file already exists")
		}
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = destFile.Close()
	}()

	if _, err := io.Copy(destFile, content); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	return nil
}

// WriteFile writes data to a file, creating parent directories if needed
func WriteFile(path string, data []byte) error {
	// Create parent directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}

	return nil
}
