# Binary Size Optimization

This document describes the binary size optimizations applied to gzcli and their results.

## Optimization Results

| Build Type | Binary Size | Reduction | Notes |
|-----------|-------------|-----------|-------|
| **Original** | 27 MB | - | Baseline build without optimization |
| **Optimized** | 18 MB | 33% | With `-trimpath`, `-s`, `-w` flags |

## Applied Optimizations

### Build Flags

The following build flags are used to reduce binary size:

- **`-trimpath`:** Removes file system paths from the compiled executable, making it more reproducible and smaller
- **`-s`:** Omits the symbol table and debug information
- **`-w`:** Omits the DWARF symbol table

These flags are applied in:
- `.goreleaser.yml` (for releases)
- `Makefile` (for local builds)

### How It Works

The optimization flags:
1. **Remove debug symbols** (`-s`, `-w`): Strips debugging information that's useful for development but not needed in production
2. **Trim file paths** (`-trimpath`): Removes absolute file system paths, making builds reproducible and smaller
3. **No runtime impact**: These are compile-time optimizations with zero performance overhead

## Building Optimized Binaries

### Local Development

```bash
# Standard optimized build (~18 MB)
make build

# Or manually with Go
go build -trimpath -ldflags="-s -w" -o gzcli .
```

### Release Builds

Release builds automatically include all optimizations:

```bash
# Create a release build with goreleaser
make release
```

The resulting binaries in `dist/` will be:
- Optimized with build flags
- Ready for distribution across all platforms

## Configuration

### GoReleaser Configuration

From `.goreleaser.yml`:

```yaml
builds:
  - id: gzcli
    binary: gzcli
    env:
      - CGO_ENABLED=0
    flags:
      - -trimpath
    ldflags:
      - -s -w
      - -X github.com/dimasma0305/gzcli/cmd.Version={{.Version}}
      - -X github.com/dimasma0305/gzcli/cmd.BuildTime={{.Date}}
      - -X github.com/dimasma0305/gzcli/cmd.GitCommit={{.Commit}}
```

### Makefile Configuration

For local builds:

```makefile
LDFLAGS=-ldflags "\
  -s -w \
  -X github.com/dimasma0305/gzcli/cmd.Version=${VERSION} \
  -X github.com/dimasma0305/gzcli/cmd.BuildTime=${BUILD_TIME} \
  -X github.com/dimasma0305/gzcli/cmd.GitCommit=${GIT_COMMIT}"

build:
	go build -trimpath $(LDFLAGS) -o gzcli .
```

## Verification

To verify the optimizations don't affect functionality:

```bash
# Check version
./gzcli --version

# Run help
./gzcli --help

# Run tests
make test
```

## Size Comparison

### By Platform

| Platform | Architecture | Size (approx) |
|----------|-------------|---------------|
| Linux | amd64 | ~18 MB |
| Linux | arm64 | ~18 MB |
| Linux | arm v6/v7 | ~17 MB |
| macOS | Universal | ~19 MB |
| Windows | amd64 | ~18 MB |
| Windows | 386 | ~17 MB |

### Compression

Archive formats further reduce download size:

- **tar.gz (Linux/macOS):** ~5-6 MB compressed
- **zip (Windows):** ~5-6 MB compressed

## Future Optimization Opportunities

Additional optimizations that could be explored:

1. **Build Tags**: Create separate builds for optional features
   - `gzcli-lite`: Core features only
   - `gzcli-full`: All features including optional components

2. **Dependency Analysis:** Review and potentially replace heavy dependencies
   - Evaluate pure-Go alternatives for dependencies
   - Remove unused dependencies

3. **Dead Code Elimination:** Use build tags to exclude unused code paths
   - Profile which features are actually used
   - Create minimal builds for specific use cases

4. **Module Trimming:** Analyze and remove unused code paths
   - Use tools like `goweight` to identify large dependencies
   - Consider alternatives for heavy dependencies

## Why No Additional Compression?

We deliberately chose not to use additional binary packers (like UPX or garble) because:

1. **Simplicity:** Standard Go builds are predictable and widely supported
2. **Compatibility:** No issues with code signing, antivirus, or platform-specific quirks
3. **Transparency:** Users can inspect binaries with standard tools
4. **Debugging:** Easier to debug issues in production if needed
5. **Build Speed:** Faster builds without additional compression steps

The 33% size reduction from build flags alone provides a good balance between size and simplicity.

## CI/CD Integration

The optimizations are automatically applied in CI/CD:

- GitHub Actions workflows use the optimized `.goreleaser.yml` configuration
- All release artifacts are consistently optimized
- Build process is reproducible and transparent

## Troubleshooting

### Large Binary Size

Current binary sizes:
- All builds: ~18 MB (optimized)
- Compressed archives: ~5-6 MB

To further reduce size:
1. Use build tags to exclude features
2. Review dependencies with `go mod graph`
3. Profile and remove unused code

### Version Information Missing

If `gzcli --version` shows "dev", ensure you're building with version flags:

```bash
# Local build with version
make build VERSION=v1.0.0

# Or let Makefile auto-detect from git
make build
```

## References

- [Go Build Modes](https://pkg.go.dev/cmd/go)
- [Go Linker Flags](https://pkg.go.dev/cmd/link)
- [GoReleaser Documentation](https://goreleaser.com/)
- [Binary Size Optimization Guide](https://github.com/golang/go/wiki/CompilerOptimizations)

## Contributing

When optimizing binary size:

1. **Measure first:** Check current binary size with `ls -lh gzcli`
2. **Document:** Explain the optimization and expected impact
3. **Test:** Ensure functionality is maintained
4. **Verify:** Confirm size reduction across platforms
5. **Review:** Consider maintainability and complexity trade-offs

For questions or suggestions, open an issue or discussion on GitHub.
