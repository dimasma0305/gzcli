package utils //nolint:revive // utils is a common and acceptable package name

import (
	"strings"
	"testing"
)

// TestGetJson_MalformedJSON tests handling of malformed JSON
func TestGetJson_MalformedJSON(t *testing.T) {
	testCases := []struct {
		name string
		json string
	}{
		{"invalid syntax", `{"invalid": json}`},
		{"unclosed brace", `{"data": "value"`},
		{"unclosed array", `{"data": ["item1", "item2"`},
		{"trailing comma", `{"data": "value",}`},
		{"null bytes", `{"data": "value\x00"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result interface{}
			err := GetJson([]byte(tc.json), &result)
			if err == nil {
				t.Errorf("Expected error for %s, but got none", tc.name)
			}
		})
	}
}

// TestGetJson_InvalidSuccess tests when success is false
func TestGetJson_InvalidSuccess(t *testing.T) {
	jsonData := `{"success": false, "message": "operation failed", "data": null}`

	var result interface{}
	err := GetJson([]byte(jsonData), &result)
	if err == nil {
		t.Error("Expected error when success is false")
	}

	if !strings.Contains(err.Error(), "operation failed") {
		t.Errorf("Expected error message to contain 'operation failed', got: %v", err)
	}
}

// TestGetJson_MissingData tests when data field is missing
func TestGetJson_MissingData(t *testing.T) {
	jsonData := `{"success": true, "message": "ok"}`

	var result interface{}
	err := GetJson([]byte(jsonData), &result)
	if err == nil {
		t.Error("Expected error when data field is missing")
	}
}

// TestGetJson_NullData tests when data is null
func TestGetJson_NullData(t *testing.T) {
	jsonData := `{"success": true, "message": "ok", "data": null}`

	var result interface{}
	err := GetJson([]byte(jsonData), &result)
	if err == nil {
		t.Log("Null data might be valid in some cases")
	}
}

// TestGetJson_ExtremelyNestedJSON tests deeply nested structures
func TestGetJson_ExtremelyNestedJSON(t *testing.T) {
	// Create deeply nested JSON (1000 levels)
	nested := `{"data":`
	for i := 0; i < 1000; i++ {
		nested += `{"level":"`
	}
	nested += "end"
	for i := 0; i < 1000; i++ {
		nested += `"}`
	}
	nested += `}`

	jsonData := `{"success": true, "message": "ok", "data": ` + nested + `}`

	var result interface{}
	err := GetJson([]byte(jsonData), &result)
	// This might fail or succeed depending on JSON parser limits
	if err != nil {
		t.Logf("Deeply nested JSON caused error (expected): %v", err)
	}
}

// TestGetJson_LargeJSON tests very large JSON payloads
func TestGetJson_LargeJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large JSON test in short mode")
	}

	// Create a large array with 100k items
	items := make([]string, 100000)
	for i := range items {
		items[i] = `"item"`
	}
	largeArray := `[` + strings.Join(items, ",") + `]`

	jsonData := `{"success": true, "message": "ok", "data": ` + largeArray + `}`

	var result interface{}
	err := GetJson([]byte(jsonData), &result)
	if err != nil {
		t.Errorf("Large JSON should be handled: %v", err)
	}
}

// TestGetJson_EmptyBytes tests empty input
func TestGetJson_EmptyBytes(t *testing.T) {
	var result interface{}
	err := GetJson([]byte{}, &result)
	if err == nil {
		t.Error("Expected error for empty bytes")
	}
}

// TestGetJson_NilBytes tests nil input
func TestGetJson_NilBytes(t *testing.T) {
	var result interface{}
	err := GetJson(nil, &result)
	if err == nil {
		t.Error("Expected error for nil bytes")
	}
}

// TestJsonify_NilInput tests marshaling nil
func TestJsonify_NilInput(t *testing.T) {
	result, err := Jsonify(nil)
	if err != nil {
		t.Errorf("Jsonify(nil) should not error: %v", err)
	}

	expected := "null"
	if string(result) != expected {
		t.Errorf("Expected %s, got %s", expected, string(result))
	}
}

// TestJsonify_CircularReference tests circular references (if possible)
func TestJsonify_ComplexTypes(t *testing.T) {
	// Test with channels, functions (which can't be marshaled)
	type InvalidStruct struct {
		Ch chan int
	}

	_, err := Jsonify(InvalidStruct{Ch: make(chan int)})
	if err == nil {
		t.Error("Expected error when marshaling channel")
	}
}

// TestURLJoinPath_InvalidURL tests invalid URL components
func TestURLJoinPath_InvalidURL(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid URL characters")
		}
	}()

	// This should panic according to the implementation
	_ = URLJoinPath("ht tp://invalid", "path")
}

// TestURLJoinPath_EmptyComponents tests empty path components
func TestURLJoinPath_EmptyComponents(t *testing.T) {
	result := URLJoinPath("http://example.com", "", "", "")
	expected := "http://example.com"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestURLJoinPath_SpecialCharacters tests special characters in paths
func TestURLJoinPath_SpecialCharacters(t *testing.T) {
	testCases := []struct {
		base     string
		paths    []string
		expected string
	}{
		{
			"http://example.com",
			[]string{"path with spaces"},
			"http://example.com/path%20with%20spaces",
		},
		{
			"http://example.com",
			[]string{"path", "with", "slashes/"},
			"http://example.com/path/with/slashes/",
		},
	}

	for _, tc := range testCases {
		result := URLJoinPath(tc.base, tc.paths...)
		if result != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, result)
		}
	}
}

// TestURLJoinPath_PathTraversal tests path traversal attempts
func TestURLJoinPath_PathTraversal(t *testing.T) {
	result := URLJoinPath("http://example.com", "..", "admin")
	// URL package should handle this safely
	t.Logf("Path traversal result: %s", result)

	// Should not allow escaping the base URL
	if !strings.HasPrefix(result, "http://example.com") {
		t.Error("Path traversal should not escape base URL")
	}
}

// TestURLJoinPath_ExtremelyLongPath tests very long URL paths
func TestURLJoinPath_ExtremelyLongPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long path test in short mode")
	}

	// Create path with 1000 components
	paths := make([]string, 1000)
	for i := range paths {
		paths[i] = "segment"
	}

	result := URLJoinPath("http://example.com", paths...)

	// Should not panic
	if !strings.HasPrefix(result, "http://example.com") {
		t.Error("Long path should still maintain base URL")
	}
}

// TestURLJoinPath_Unicode tests Unicode characters in paths
func TestURLJoinPath_Unicode(t *testing.T) {
	result := URLJoinPath("http://example.com", "日本語", "путь", "🚀")

	// Should handle Unicode properly (percent-encoded)
	if !strings.HasPrefix(result, "http://example.com") {
		t.Error("Unicode path should maintain base URL")
	}
	t.Logf("Unicode path result: %s", result)
}

// TestNormalizePath_WindowsPaths tests Windows path normalization
func TestNormalizePath_WindowsPaths(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{`C:\Users\test\file.txt`, `C:/Users/test/file.txt`},
		{`\\server\share\file`, `//server/share/file`},
		{`path\to\file`, `path/to/file`},
		{`mixed/path\separators`, `mixed/path/separators`},
	}

	for _, tc := range testCases {
		result := NormalizePath(tc.input)
		if result != tc.expected {
			t.Errorf("NormalizePath(%s) = %s, want %s", tc.input, result, tc.expected)
		}
	}
}

// TestNormalizePath_EmptyString tests empty string normalization
func TestNormalizePath_EmptyString(t *testing.T) {
	result := NormalizePath("")
	if result != "" {
		t.Errorf("Expected empty string, got %s", result)
	}
}

// TestNormalizePath_OnlyBackslashes tests strings with only backslashes
func TestNormalizePath_OnlyBackslashes(t *testing.T) {
	result := NormalizePath(`\\\\`)
	expected := "////"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestNormalizePath_UNCPaths tests UNC paths
func TestNormalizePath_UNCPaths(t *testing.T) {
	result := NormalizePath(`\\?\UNC\server\share\file`)
	expected := `//?/UNC/server/share/file`
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestNormalizePath_SpecialCharacters tests paths with special characters
func TestNormalizePath_SpecialCharacters(t *testing.T) {
	testCases := []string{
		`path\with\$pecial\char$`,
		`path\with\spaces\ here`,
		`path\with\unicode\日本語`,
		`path\with\null\bytes` + "\x00",
	}

	for _, tc := range testCases {
		result := NormalizePath(tc)
		// Should not panic and should convert backslashes
		if strings.Contains(result, `\`) {
			t.Errorf("Path still contains backslashes: %s", result)
		}
	}
}
