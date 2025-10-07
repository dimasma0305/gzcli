//nolint:revive // Package utils is used for utility functions
package utils

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 4096))
	},
}

// ParseYamlFromBytes parses YAML data from bytes into a data structure
func ParseYamlFromBytes(b []byte, data any) error {
	if err := yaml.Unmarshal(b, data); err != nil {
		return fmt.Errorf("error unmarshal yaml: %w", err)
	}
	return nil
}

// ParseYamlFromFile parses YAML data from a file into a data structure
func ParseYamlFromFile(confPath string, data any) error {
	buf := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buf)
	defer buf.Reset()

	//nolint:gosec // G304: Config path is constructed by application
	f, err := os.Open(confPath)
	if err != nil {
		return fmt.Errorf("file open error: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := buf.ReadFrom(f); err != nil {
		return fmt.Errorf("file read error: %w", err)
	}

	return ParseYamlFromBytes(buf.Bytes(), data)
}
