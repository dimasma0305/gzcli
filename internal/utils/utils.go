// Package utils provides common utility functions
//
//nolint:revive // utils is a common and acceptable package name for utility functions
package utils

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// NormalizePath normalizes a file path for Windows compatibility
func NormalizePath(str string) string {
	str = strings.ReplaceAll(str, "\\", "/")
	return str
}

// GetJSON unmarshals JSON data from bytes into the provided data structure
func GetJSON(b []byte, data any) error {
	var tmp struct {
		Message string
		Success bool
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	if !tmp.Success {
		return fmt.Errorf("request end with %s status", tmp.Message)
	}
	if err := json.Unmarshal(tmp.Data, data); err != nil {
		return err
	}
	return nil
}

// GetJson is deprecated, use GetJSON instead
//
//nolint:revive // Deprecated function kept for backward compatibility
func GetJson(b []byte, data any) error {
	return GetJSON(b, data)
}

// Jsonify marshals an object to JSON bytes
func Jsonify(data any) ([]byte, error) {
	return json.Marshal(data)
}

// URLJoinPath joins URL paths safely
func URLJoinPath(base string, path ...string) string {
	res, err := url.JoinPath(base, path...)
	if err != nil {
		panic(err)
	}
	return res
}
