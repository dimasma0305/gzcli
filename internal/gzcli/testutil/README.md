# Test Utilities

This package provides common testing utilities and helpers for gzcli tests.

## Overview

The `testutil` package contains shared testing functionality to reduce code duplication and make tests more maintainable. It includes mock implementations, test data generators, and assertion helpers.

## Available Utilities

### Test Data Helpers

Functions for creating test data and fixtures:

```go
// Create a temporary directory with test data
tmpDir := testutil.TempDir(t)

// Load test fixture
data := testutil.LoadFixture(t, "testdata/challenge.yaml")

// Create mock challenge configuration
challenge := testutil.MockChallenge("Web", "XSS Challenge")
```

### Mock Implementations

Mock implementations of key interfaces for testing:

#### Mock API Client

```go
// Create a mock API client that returns predefined responses
mockAPI := testutil.NewMockAPI()
mockAPI.LoginFunc = func(username, password string) error {
    if username == "admin" && password == "password" {
        return nil
    }
    return errors.New("invalid credentials")
}

// Use in tests
err := mockAPI.Login("admin", "password")
```

#### Mock File System

```go
// Create a mock file system for testing file operations
mockFS := testutil.NewMockFS()
mockFS.WriteFile("test.txt", []byte("content"))

// Verify operations
if !mockFS.FileExists("test.txt") {
    t.Error("file should exist")
}
```

### Assertion Helpers

Common assertion helpers to make tests more readable:

```go
// Assert no error
testutil.AssertNoError(t, err)

// Assert error contains message
testutil.AssertErrorContains(t, err, "expected message")

// Assert equality
testutil.AssertEqual(t, got, want)

// Assert file exists
testutil.AssertFileExists(t, "path/to/file")

// Assert directory structure
testutil.AssertDirStructure(t, tmpDir, []string{
    "challenge1/",
    "challenge1/challenge.yaml",
    "challenge2/",
})
```

### HTTP Test Helpers

Utilities for testing HTTP operations:

```go
// Create a test HTTP server
server := testutil.NewTestServer(t)
defer server.Close()

// Configure responses
server.ResponseFor("/api/login", 200, `{"token": "test"}`)

// Get server URL
apiURL := server.URL()
```

## Usage Examples

### Testing API Client

```go
func TestGZAPI_Login(t *testing.T) {
    mockAPI := testutil.NewMockAPI()
    mockAPI.LoginFunc = func(username, password string) error {
        return nil
    }

    err := mockAPI.Login("admin", "password")
    testutil.AssertNoError(t, err)
}
```

### Testing File Operations

```go
func TestLoadConfig(t *testing.T) {
    // Create temporary directory
    tmpDir := testutil.TempDir(t)

    // Create test config file
    configPath := filepath.Join(tmpDir, "config.yaml")
    testutil.WriteFile(t, configPath, []byte(`url: https://example.com`))

    // Test loading
    cfg, err := LoadConfig(configPath)
    testutil.AssertNoError(t, err)
    testutil.AssertEqual(t, cfg.URL, "https://example.com")
}
```

### Testing with Fixtures

```go
func TestParseChallenge(t *testing.T) {
    // Load fixture
    data := testutil.LoadFixture(t, "valid_challenge.yaml")

    // Parse
    challenge, err := ParseChallenge(data)
    testutil.AssertNoError(t, err)

    // Assertions
    testutil.AssertEqual(t, challenge.Name, "Test Challenge")
}
```

### Integration Testing

```go
func TestSync_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Create test environment
    env := testutil.NewTestEnv(t)
    defer env.Cleanup()

    // Setup challenges
    env.AddChallenge("web", "xss")
    env.AddChallenge("pwn", "buffer-overflow")

    // Run sync
    err := Sync(env.Dir())
    testutil.AssertNoError(t, err)

    // Verify results
    testutil.AssertFileExists(t, env.Path("web/xss/challenge.yaml"))
}
```

## Creating Custom Test Helpers

When creating new test helpers, follow these guidelines:

### 1. Helper Should Accept *testing.T

```go
func AssertSomething(t *testing.T, value string) {
    t.Helper() // Mark as helper for better error reporting

    if value == "" {
        t.Error("value should not be empty")
    }
}
```

### 2. Use t.Helper()

Always call `t.Helper()` at the beginning of helper functions:

```go
func CreateTestData(t *testing.T) string {
    t.Helper()
    // Helper implementation
}
```

### 3. Fail Fast on Setup Errors

Use `t.Fatal()` for setup errors, `t.Error()` for assertion failures:

```go
func LoadTestFile(t *testing.T, path string) []byte {
    t.Helper()

    data, err := os.ReadFile(path)
    if err != nil {
        t.Fatalf("failed to load test file: %v", err)
    }
    return data
}
```

### 4. Clean Up Resources

Use `t.Cleanup()` for resource cleanup:

```go
func CreateTempFile(t *testing.T) string {
    t.Helper()

    f, err := os.CreateTemp("", "test-*.txt")
    if err != nil {
        t.Fatalf("failed to create temp file: %v", err)
    }

    t.Cleanup(func() {
        os.Remove(f.Name())
    })

    return f.Name()
}
```

## Test Data Organization

### Fixtures Location

Test fixtures should be organized in `testdata/` directories:

```
package/
├── handler.go
├── handler_test.go
└── testdata/
    ├── valid_input.yaml
    ├── invalid_input.yaml
    └── README.md  # Document fixtures
```

### Fixture Naming

Use descriptive names that indicate the test scenario:

- `valid_*.yaml` - Valid test data
- `invalid_*.yaml` - Invalid/malformed data
- `empty_*.yaml` - Empty or minimal data
- `complex_*.yaml` - Complex scenarios

## Mock Best Practices

### Interface-Based Mocking

Define interfaces for dependencies to enable easy mocking:

```go
// Define interface
type APIClient interface {
    Login(username, password string) error
    GetGames() ([]Game, error)
}

// Implementation
type GZAPI struct {
    // ...
}

// Mock
type MockAPI struct {
    LoginFunc    func(string, string) error
    GetGamesFunc func() ([]Game, error)
}

func (m *MockAPI) Login(u, p string) error {
    if m.LoginFunc != nil {
        return m.LoginFunc(u, p)
    }
    return nil
}
```

### Recording Mocks

For verification, create mocks that record calls:

```go
type RecordingMock struct {
    calls []string
}

func (m *RecordingMock) Method(arg string) {
    m.calls = append(m.calls, arg)
}

func (m *RecordingMock) AssertCalled(t *testing.T, arg string) {
    t.Helper()
    for _, call := range m.calls {
        if call == arg {
            return
        }
    }
    t.Errorf("Method was not called with %q", arg)
}
```

## Testing Async Operations

### Using Channels

```go
func TestAsyncOperation(t *testing.T) {
    done := make(chan bool)

    go func() {
        // Async operation
        done <- true
    }()

    select {
    case <-done:
        // Success
    case <-time.After(5 * time.Second):
        t.Fatal("operation timed out")
    }
}
```

### Using sync.WaitGroup

```go
func TestConcurrentOperations(t *testing.T) {
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            // Concurrent operation
        }(i)
    }

    wg.Wait()
}
```

## Performance Testing

### Benchmarks

```go
func BenchmarkOperation(b *testing.B) {
    // Setup
    setup := prepareTestData()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        Operation(setup)
    }
}
```

### Memory Profiling

```go
func BenchmarkMemory(b *testing.B) {
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        // Operation
    }
}
```

## Common Patterns

### Table-Driven Tests with testutil

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Feature(tt.input)

            testutil.AssertError(t, err, tt.wantErr)
            if !tt.wantErr {
                testutil.AssertEqual(t, got, tt.want)
            }
        })
    }
}
```

## Contributing

When adding new test utilities:

1. Add documentation in this README
2. Include usage examples
3. Write tests for the utility itself
4. Follow existing patterns and naming conventions
5. Update TESTING.md if adding major new functionality

## Resources

- [TESTING.md](../../../TESTING.md) - Overall testing guide
- [Go Testing Package](https://golang.org/pkg/testing/)
- [Testify Package](https://github.com/stretchr/testify) - External assertion library (if we add it)
