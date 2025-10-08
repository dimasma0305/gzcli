# Testing Guide

This guide provides comprehensive information about writing, running, and maintaining tests in gzcli.

## Table of Contents

- [Overview](#overview)
- [Testing Philosophy](#testing-philosophy)
- [Running Tests](#running-tests)
- [Writing Tests](#writing-tests)
- [Test Organization](#test-organization)
- [Mocking Strategies](#mocking-strategies)
- [Integration Tests](#integration-tests)
- [Test Coverage](#test-coverage)
- [Best Practices](#best-practices)

## Overview

gzcli uses Go's built-in testing framework along with table-driven tests for comprehensive test coverage. We aim for:

- **>80% code coverage** for core packages
- **Fast unit tests:** <5 seconds for the full suite
- **Comprehensive integration tests** for critical paths
- **Clear test names** that describe what is being tested

## Testing Philosophy

### Test Pyramid

We follow the testing pyramid approach:

```
        /\
       /  \      E2E Tests (Few)
      /----\
     /      \    Integration Tests (Some)
    /--------\
   /          \  Unit Tests (Many)
  /____________\
```

- **Unit Tests:** Test individual functions and methods in isolation
- **Integration Tests:** Test interactions between components
- **E2E Tests:** Test complete workflows (manual or CI-only)

### Test Naming

Tests should clearly describe **what** is being tested and **what** the expected behavior is:

```go
func TestWatcher_HandleFileChange_RedeploysChallenge(t *testing.T)
func TestGZAPI_Login_WithInvalidCredentials_ReturnsError(t *testing.T)
func TestConfig_Load_WithMissingFile_ReturnsError(t *testing.T)
```

**Format:** `Test<Component>_<Method>_<Scenario>_<ExpectedResult>`

## Running Tests

### All Tests

```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run with race detector
make test-race
```

### Unit Tests Only

```bash
# Short mode (skips slow integration tests)
make test-unit

# Or directly
go test -short ./...
```

### Specific Tests

```bash
# Run tests in a specific package
go test -v ./internal/gzcli/watcher/...

# Run a specific test
go test -v ./internal/gzcli/watcher -run TestWatcher_Start

# Run tests matching a pattern
go test -v ./... -run "TestWatcher.*"
```

### Component-Specific Tests

```bash
# Watcher tests
make test-watcher

# Challenge tests
make test-challenge

# API tests
make test-api
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# View in browser
make coverage-browse

# Check coverage for specific package
go test -cover ./internal/gzcli/watcher/...
```

### Watch Mode

```bash
# Auto-run tests on file changes
make test-watch
```

## Writing Tests

### Table-Driven Tests

Use table-driven tests for testing multiple scenarios:

```go
func TestParseChallenge(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Challenge
        wantErr bool
    }{
        {
            name: "valid challenge",
            input: `name: Test
category: Web`,
            want: &Challenge{
                Name:     "Test",
                Category: "Web",
            },
            wantErr: false,
        },
        {
            name:    "invalid YAML",
            input:   "invalid: [",
            want:    nil,
            wantErr: true,
        },
        {
            name:    "empty input",
            input:   "",
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseChallenge(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("ParseChallenge() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseChallenge() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Testing with Fixtures

Use test fixtures for complex test data:

```go
func TestLoadConfig(t *testing.T) {
    // Load fixture
    data, err := os.ReadFile("testdata/valid_config.yaml")
    if err != nil {
        t.Fatalf("Failed to load fixture: %v", err)
    }

    cfg, err := ParseConfig(data)
    if err != nil {
        t.Errorf("ParseConfig() failed: %v", err)
    }

    // Assertions
    if cfg.URL != "https://example.com" {
        t.Errorf("URL = %v, want https://example.com", cfg.URL)
    }
}
```

### Using Subtests

Subtests provide better organization and allow running specific test cases:

```go
func TestAPI(t *testing.T) {
    t.Run("Login", func(t *testing.T) {
        t.Run("ValidCredentials", func(t *testing.T) {
            // Test valid login
        })

        t.Run("InvalidCredentials", func(t *testing.T) {
            // Test invalid login
        })
    })

    t.Run("GetGames", func(t *testing.T) {
        // Test getting games
    })
}
```

### Error Testing

Always test both success and error cases:

```go
func TestDivide(t *testing.T) {
    t.Run("ValidDivision", func(t *testing.T) {
        result, err := Divide(10, 2)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        if result != 5 {
            t.Errorf("got %v, want 5", result)
        }
    })

    t.Run("DivisionByZero", func(t *testing.T) {
        _, err := Divide(10, 0)
        if err == nil {
            t.Error("expected error, got nil")
        }
    })
}
```

## Test Organization

### File Structure

```
package/
├── handler.go
├── handler_test.go                  # Unit tests
├── handler_integration_test.go      # Integration tests
├── testdata/                        # Test fixtures
│   ├── valid_input.json
│   └── invalid_input.json
└── testutil/                        # Test helpers
    └── mocks.go
```

### Test Tags

Use build tags to separate test types:

```go
//go:build integration
// +build integration

package gzapi_test
```

Run with: `go test -tags=integration ./...`

## Mocking Strategies

### Interface Mocking

Define interfaces for dependencies:

```go
// Interface for HTTP client
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

// Mock implementation
type MockHTTPClient struct {
    DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    return m.DoFunc(req)
}

// Test using mock
func TestAPICall(t *testing.T) {
    mockClient := &MockHTTPClient{
        DoFunc: func(req *http.Request) (*http.Response, error) {
            return &http.Response{
                StatusCode: 200,
                Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
            }, nil
        },
    }

    api := &API{client: mockClient}
    // Test with mock client
}
```

### Using testutil Package

See [testutil README](internal/gzcli/testutil/README.md) for available test utilities.

## Integration Tests

Integration tests verify interactions between components:

```go
func TestSync_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup test environment
    tmpDir := t.TempDir()

    // Create test files
    // ...

    // Run sync
    err := Sync(tmpDir)
    if err != nil {
        t.Fatalf("Sync failed: %v", err)
    }

    // Verify results
    // ...
}
```

Skip integration tests with: `go test -short ./...`

## Test Coverage

### Measuring Coverage

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Show coverage by function
go tool cover -func=coverage.out
```

### Coverage Goals

- **Critical packages** (watcher, challenge, API): >85%
- **Command handlers:** >70%
- **Utilities:** >80%
- **Overall project:** >80%

### Improving Coverage

1. Identify uncovered code:
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```

2. Add tests for uncovered paths
3. Focus on error paths and edge cases
4. Use table-driven tests to cover multiple scenarios

## Best Practices

### DO

✅ Write tests before fixing bugs (TDD for bug fixes)
✅ Test public APIs and exported functions
✅ Use table-driven tests for multiple scenarios
✅ Test error conditions and edge cases
✅ Keep tests fast and focused
✅ Use meaningful test names
✅ Clean up resources in tests (use `t.Cleanup()`)
✅ Use subtests for better organization
✅ Mock external dependencies
✅ Test concurrent code with race detector

### DON'T

❌ Test implementation details
❌ Write flaky tests (time-dependent, order-dependent)
❌ Ignore test failures
❌ Skip testing error paths
❌ Write tests that depend on external services (without mocks)
❌ Commit code with failing tests
❌ Use sleep() to wait for async operations
❌ Hard-code paths or configurations

### Example: Clean Test Structure

```go
func TestFeature(t *testing.T) {
    // Setup
    tmpDir := t.TempDir()
    t.Cleanup(func() {
        // Cleanup if needed (tmpDir auto-cleaned)
    })

    // Execute
    result, err := Feature(tmpDir)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

## Benchmarking

### Writing Benchmarks

```go
func BenchmarkParseChallenge(b *testing.B) {
    input := loadTestData()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ParseChallenge(input)
    }
}
```

### Running Benchmarks

```bash
# Run benchmarks
go test -bench=. ./...

# With memory stats
go test -bench=. -benchmem ./...

# Compare benchmarks
make bench > old.txt
# Make changes
make bench > new.txt
benchcmp old.txt new.txt
```

## Continuous Integration

Tests run automatically in CI on:
- Every push to `main` or `develop`
- Every pull request
- Multiple Go versions (1.23+)
- Multiple operating systems (Linux, macOS, Windows)

See `.github/workflows/ci.yml` for CI configuration.

## Troubleshooting

### Tests Failing Locally

```bash
# Clean build cache
go clean -testcache

# Update dependencies
go mod tidy
go mod download

# Run with verbose output
go test -v ./...
```

### Flaky Tests

If tests are flaky:
1. Remove timing dependencies
2. Use proper synchronization (channels, WaitGroups)
3. Increase timeouts if necessary
4. Ensure proper test isolation

### Race Conditions

```bash
# Run with race detector
go test -race ./...

# Fix race conditions by using:
# - Mutexes
# - Proper channel usage
# - Avoiding shared state
```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Code Review Comments: Tests](https://github.com/golang/go/wiki/CodeReviewComments#tests)
- [Advanced Testing with Go](https://www.youtube.com/watch?v=8hQG7QlcLBk)

## Getting Help

- Check existing tests for examples
- Review [testutil README](internal/gzcli/testutil/README.md) for available helpers
- Ask in [GitHub Discussions](https://github.com/dimasma0305/gzcli/discussions)
- Open an [issue](https://github.com/dimasma0305/gzcli/issues) for testing-related questions
