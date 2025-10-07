package utils

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
	"runtime"
	"strings"
	"sync"
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
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }()

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

	// Pre-allocate buffer pool
	bufPool := sync.Pool{
		New: func() interface{} { return make([]byte, 32<<10) }, // 32KB buffers
	}

	// Collect files first to enable parallel processing
	var filePaths []string
	_ = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		filePaths = append(filePaths, path)
		return nil
	})

	// Process files in parallel but write sequentially
	type result struct {
		path string
		data []byte
		err  error
	}
	resultChan := make(chan result, len(filePaths))

	// Worker pool for parallel reading
	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup

	// Use a fixed timestamp for reproducible builds
	fixedTime := time.Date(2025, 3, 18, 0, 0, 0, 0, time.UTC)

	for _, path := range filePaths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Read file content
			data, err := os.ReadFile(p)
			resultChan <- result{p, data, err}
		}(path)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Write results in original order while maintaining directory structure
	writtenFiles := make(map[string]struct{})
	for res := range resultChan {
		if res.err != nil {
			return res.err
		}

		relPath, err := filepath.Rel(source, res.path)
		if err != nil {
			return err
		}

		// Ensure directory entries exist
		dirPath := filepath.Dir(relPath)
		if dirPath != "." {
			if _, exists := writtenFiles[dirPath]; !exists {
				header := &zip.FileHeader{
					Name:     dirPath + "/",
					Method:   zip.Deflate,
					Modified: fixedTime,
				}
				if _, err := writer.CreateHeader(header); err != nil {
					return err
				}
				writtenFiles[dirPath] = struct{}{}
			}
		}

		// Create file header
		header := &zip.FileHeader{
			Name:     relPath,
			Method:   zip.Deflate,
			Modified: fixedTime,
		}
		header.SetMode(0644)

		// Use buffer from pool
		buf := bufPool.Get().([]byte)
		defer bufPool.Put(&buf)

		// Write to zip
		w, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if _, err := io.CopyBuffer(w, bytes.NewReader(res.data), buf); err != nil {
			return err
		}
	}

	return nil
}
