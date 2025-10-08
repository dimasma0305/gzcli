//nolint:revive // utils is a common and acceptable package name
package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// testStruct is a simple struct for YAML parsing tests
type testStruct struct {
	Name    string `yaml:"name"`
	Age     int    `yaml:"age"`
	Active  bool   `yaml:"active"`
	Tags    []string `yaml:"tags"`
	Nested  nestedStruct `yaml:"nested"`
}

type nestedStruct struct {
	Field1 string `yaml:"field1"`
	Field2 int    `yaml:"field2"`
}

// TestParseYamlFromBytes_Success tests successful YAML parsing from bytes
func TestParseYamlFromBytes_Success(t *testing.T) {
	yamlData := []byte(`
name: John Doe
age: 30
active: true
tags:
  - tag1
  - tag2
nested:
  field1: value1
  field2: 42
`)

	var result testStruct
	err := ParseYamlFromBytes(yamlData, &result)
	if err != nil {
		t.Fatalf("ParseYamlFromBytes() failed: %v", err)
	}

	// Verify parsed data
	if result.Name != "John Doe" {
		t.Errorf("Name = %q, want %q", result.Name, "John Doe")
	}

	if result.Age != 30 {
		t.Errorf("Age = %d, want %d", result.Age, 30)
	}

	if !result.Active {
		t.Error("Active = false, want true")
	}

	if len(result.Tags) != 2 {
		t.Errorf("len(Tags) = %d, want 2", len(result.Tags))
	}

	if result.Nested.Field1 != "value1" {
		t.Errorf("Nested.Field1 = %q, want %q", result.Nested.Field1, "value1")
	}

	if result.Nested.Field2 != 42 {
		t.Errorf("Nested.Field2 = %d, want %d", result.Nested.Field2, 42)
	}
}

// TestParseYamlFromBytes_EmptyData tests parsing empty YAML
func TestParseYamlFromBytes_EmptyData(t *testing.T) {
	yamlData := []byte{}

	var result testStruct
	err := ParseYamlFromBytes(yamlData, &result)
	if err != nil {
		t.Errorf("ParseYamlFromBytes() with empty data failed: %v", err)
	}

	// Should have default values
	if result.Name != "" {
		t.Errorf("Name = %q, want empty string", result.Name)
	}

	if result.Age != 0 {
		t.Errorf("Age = %d, want 0", result.Age)
	}
}

// TestParseYamlFromBytes_InvalidYAML tests error handling for invalid YAML
func TestParseYamlFromBytes_InvalidYAML(t *testing.T) {
	yamlData := []byte(`
name: John Doe
age: invalid_number
`)

	var result testStruct
	err := ParseYamlFromBytes(yamlData, &result)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

// TestParseYamlFromBytes_MalformedYAML tests malformed YAML
func TestParseYamlFromBytes_MalformedYAML(t *testing.T) {
	yamlData := []byte(`
name: John Doe
  age: 30
    active: true
`)

	var result testStruct
	err := ParseYamlFromBytes(yamlData, &result)
	if err == nil {
		t.Error("Expected error for malformed YAML, got nil")
	}
}

// TestParseYamlFromBytes_PartialData tests parsing with partial data
func TestParseYamlFromBytes_PartialData(t *testing.T) {
	yamlData := []byte(`
name: Jane Doe
`)

	var result testStruct
	err := ParseYamlFromBytes(yamlData, &result)
	if err != nil {
		t.Fatalf("ParseYamlFromBytes() with partial data failed: %v", err)
	}

	if result.Name != "Jane Doe" {
		t.Errorf("Name = %q, want %q", result.Name, "Jane Doe")
	}

	// Other fields should have default values
	if result.Age != 0 {
		t.Errorf("Age = %d, want 0", result.Age)
	}

	if result.Active {
		t.Error("Active = true, want false")
	}
}

// TestParseYamlFromBytes_SpecialCharacters tests special characters
func TestParseYamlFromBytes_SpecialCharacters(t *testing.T) {
	yamlData := []byte(`
name: "John \"Doe\""
age: 30
`)

	var result testStruct
	err := ParseYamlFromBytes(yamlData, &result)
	if err != nil {
		t.Fatalf("ParseYamlFromBytes() with special characters failed: %v", err)
	}

	expected := `John "Doe"`
	if result.Name != expected {
		t.Errorf("Name = %q, want %q", result.Name, expected)
	}
}

// TestParseYamlFromBytes_UnicodeCharacters tests unicode
func TestParseYamlFromBytes_UnicodeCharacters(t *testing.T) {
	yamlData := []byte(`
name: "日本語"
age: 25
`)

	var result testStruct
	err := ParseYamlFromBytes(yamlData, &result)
	if err != nil {
		t.Fatalf("ParseYamlFromBytes() with unicode failed: %v", err)
	}

	if result.Name != "日本語" {
		t.Errorf("Name = %q, want %q", result.Name, "日本語")
	}
}

// TestParseYamlFromFile_Success tests successful file parsing
func TestParseYamlFromFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	yamlData := []byte(`
name: Alice
age: 25
active: true
tags:
  - golang
  - testing
`)

	if err := os.WriteFile(testFile, yamlData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var result testStruct
	err := ParseYamlFromFile(testFile, &result)
	if err != nil {
		t.Fatalf("ParseYamlFromFile() failed: %v", err)
	}

	if result.Name != "Alice" {
		t.Errorf("Name = %q, want %q", result.Name, "Alice")
	}

	if result.Age != 25 {
		t.Errorf("Age = %d, want %d", result.Age, 25)
	}

	if !result.Active {
		t.Error("Active = false, want true")
	}

	if len(result.Tags) != 2 {
		t.Errorf("len(Tags) = %d, want 2", len(result.Tags))
	}
}

// TestParseYamlFromFile_NonExistentFile tests error handling
func TestParseYamlFromFile_NonExistentFile(t *testing.T) {
	var result testStruct
	err := ParseYamlFromFile("/nonexistent/file.yaml", &result)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestParseYamlFromFile_EmptyFile tests parsing empty file
func TestParseYamlFromFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.yaml")

	if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var result testStruct
	err := ParseYamlFromFile(testFile, &result)
	if err != nil {
		t.Errorf("ParseYamlFromFile() with empty file failed: %v", err)
	}

	// Should have default values
	if result.Name != "" {
		t.Errorf("Name = %q, want empty string", result.Name)
	}
}

// TestParseYamlFromFile_InvalidYAML tests error handling for invalid YAML
func TestParseYamlFromFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.yaml")

	invalidYaml := []byte(`
name: Bob
age: not_a_number
`)

	if err := os.WriteFile(testFile, invalidYaml, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var result testStruct
	err := ParseYamlFromFile(testFile, &result)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

// TestParseYamlFromFile_LargeFile tests parsing large file
func TestParseYamlFromFile_LargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.yaml")

	// Create a large YAML file with many tags
	yamlData := "name: Large Test\nage: 30\ntags:\n"
	for i := 0; i < 1000; i++ {
		yamlData += "  - tag" + string(rune('0'+i%10)) + "\n"
	}

	if err := os.WriteFile(testFile, []byte(yamlData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var result testStruct
	err := ParseYamlFromFile(testFile, &result)
	if err != nil {
		t.Fatalf("ParseYamlFromFile() with large file failed: %v", err)
	}

	if result.Name != "Large Test" {
		t.Errorf("Name = %q, want %q", result.Name, "Large Test")
	}

	if len(result.Tags) != 1000 {
		t.Errorf("len(Tags) = %d, want 1000", len(result.Tags))
	}
}

// TestParseYamlFromFile_PermissionDenied tests permission error handling
func TestParseYamlFromFile_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "noperm.yaml")

	if err := os.WriteFile(testFile, []byte("name: test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Remove read permissions
	if err := os.Chmod(testFile, 0000); err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}
	defer os.Chmod(testFile, 0644) // Cleanup

	var result testStruct
	err := ParseYamlFromFile(testFile, &result)
	if err == nil {
		t.Error("Expected error for permission denied, got nil")
	}
}

// TestParseYamlFromFile_ComplexNestedStructure tests complex nested YAML
func TestParseYamlFromFile_ComplexNestedStructure(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "complex.yaml")

	yamlData := []byte(`
name: Complex Test
age: 40
active: true
tags:
  - tag1
  - tag2
  - tag3
nested:
  field1: nested_value
  field2: 100
`)

	if err := os.WriteFile(testFile, yamlData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var result testStruct
	err := ParseYamlFromFile(testFile, &result)
	if err != nil {
		t.Fatalf("ParseYamlFromFile() with complex structure failed: %v", err)
	}

	if result.Name != "Complex Test" {
		t.Errorf("Name = %q, want %q", result.Name, "Complex Test")
	}

	if result.Nested.Field1 != "nested_value" {
		t.Errorf("Nested.Field1 = %q, want %q", result.Nested.Field1, "nested_value")
	}

	if result.Nested.Field2 != 100 {
		t.Errorf("Nested.Field2 = %d, want %d", result.Nested.Field2, 100)
	}
}

// TestBufferPool_Reuse tests that buffer pool is working
func TestBufferPool_Reuse(t *testing.T) {
	// Get buffer from pool
	buf1 := bufferPool.Get()
	if buf1 == nil {
		t.Fatal("Expected non-nil buffer from pool")
	}

	// Put it back
	bufferPool.Put(buf1)

	// Get it again - should be the same buffer
	buf2 := bufferPool.Get()
	if buf2 == nil {
		t.Fatal("Expected non-nil buffer from pool")
	}

	// Note: We can't guarantee buf1 == buf2 due to pool implementation,
	// but at least verify it works
	bufferPool.Put(buf2)
}

// TestParseYamlFromBytes_Map tests parsing into map
func TestParseYamlFromBytes_Map(t *testing.T) {
	yamlData := []byte(`
key1: value1
key2: value2
nested:
  subkey: subvalue
`)

	var result map[string]interface{}
	err := ParseYamlFromBytes(yamlData, &result)
	if err != nil {
		t.Fatalf("ParseYamlFromBytes() into map failed: %v", err)
	}

	if result["key1"] != "value1" {
		t.Errorf("key1 = %v, want %q", result["key1"], "value1")
	}

	if result["key2"] != "value2" {
		t.Errorf("key2 = %v, want %q", result["key2"], "value2")
	}

	nested, ok := result["nested"].(map[interface{}]interface{})
	if !ok {
		t.Fatal("nested is not a map")
	}

	if nested["subkey"] != "subvalue" {
		t.Errorf("nested.subkey = %v, want %q", nested["subkey"], "subvalue")
	}
}

// TestParseYamlFromBytes_Array tests parsing YAML array
func TestParseYamlFromBytes_Array(t *testing.T) {
	yamlData := []byte(`
- item1
- item2
- item3
`)

	var result []string
	err := ParseYamlFromBytes(yamlData, &result)
	if err != nil {
		t.Fatalf("ParseYamlFromBytes() into array failed: %v", err)
	}

	expected := []string{"item1", "item2", "item3"}
	if len(result) != len(expected) {
		t.Fatalf("len(result) = %d, want %d", len(result), len(expected))
	}

	for i, v := range expected {
		if result[i] != v {
			t.Errorf("result[%d] = %q, want %q", i, result[i], v)
		}
	}
}
