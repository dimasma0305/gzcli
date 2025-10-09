package repository

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/dimasma0305/gzcli/internal/gzcli/errors"
	"github.com/dimasma0305/gzcli/internal/log"
)

// FileRepository implements FileRepository for filesystem operations
type FileRepository struct {
	basePath string
}

// NewFileRepository creates a new file repository
func NewFileRepository(basePath string) FileRepository {
	return FileRepository{
		basePath: basePath,
	}
}

// ReadFile reads a file from the filesystem
func (r *FileRepository) ReadFile(ctx context.Context, path string) ([]byte, error) {
	fullPath := r.resolvePath(path)
	
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrapf(errors.ErrFileNotFound, "file not found: %s", path)
		}
		return nil, errors.Wrapf(err, "failed to read file: %s", path)
	}

	log.Debug("Successfully read file: %s", path)
	return data, nil
}

// WriteFile writes data to a file
func (r *FileRepository) WriteFile(ctx context.Context, path string, data []byte) error {
	fullPath := r.resolvePath(path)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrapf(err, "failed to create directory: %s", dir)
	}

	err := os.WriteFile(fullPath, data, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to write file: %s", path)
	}

	log.Debug("Successfully wrote file: %s", path)
	return nil
}

// Exists checks if a file exists
func (r *FileRepository) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := r.resolvePath(path)
	
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to check file existence: %s", path)
	}

	return true, nil
}

// DeleteFile deletes a file
func (r *FileRepository) DeleteFile(ctx context.Context, path string) error {
	fullPath := r.resolvePath(path)
	
	err := os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Wrapf(errors.ErrFileNotFound, "file not found: %s", path)
		}
		return errors.Wrapf(err, "failed to delete file: %s", path)
	}

	log.Debug("Successfully deleted file: %s", path)
	return nil
}

// ListFiles lists files in a directory
func (r *FileRepository) ListFiles(ctx context.Context, dir string) ([]string, error) {
	fullPath := r.resolvePath(dir)
	
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrapf(errors.ErrFileNotFound, "directory not found: %s", dir)
		}
		return nil, errors.Wrapf(err, "failed to list files in directory: %s", dir)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	log.Debug("Successfully listed %d files in directory: %s", len(files), dir)
	return files, nil
}

// resolvePath resolves a relative path to an absolute path
func (r *FileRepository) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(r.basePath, path)
}

// WalkFiles walks through files in a directory recursively
func (r *FileRepository) WalkFiles(ctx context.Context, dir string, fn func(string, fs.FileInfo) error) error {
	fullPath := r.resolvePath(dir)
	
	return filepath.Walk(fullPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Get relative path
		relPath, err := filepath.Rel(fullPath, path)
		if err != nil {
			return errors.Wrapf(err, "failed to get relative path: %s", path)
		}
		
		return fn(relPath, info)
	})
}