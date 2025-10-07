# Binary Size Optimization

This document describes the binary size optimizations applied to gzcli and their results.

## Optimization Results

| Build Type | Binary Size | Reduction | Notes |
|-----------|-------------|-----------|-------|
| **Original** | 27 MB | - | Baseline build |
| **Optimized** | 18 MB | 33% | With `-trimpath`, `-s`, `-w` flags |
| **UPX Compressed** | 5.4 MB | 80% | With UPX best compression + LZMA |

## Applied Optimizations

### 1. Build Flags

The following build flags are now used to reduce binary size:

- **`-trimpath`**: Removes file system paths from the compiled executable, making it more reproducible and smaller
- **`-s`**: Omits the symbol table and debug information
- **`-w`**: Omits the DWARF symbol table

These flags are applied in:
- `.goreleaser.yml` (for releases)
- `Makefile` (for local builds)

### 2. UPX Compression

UPX (Ultimate Packer for eXecutables) compression is configured in `.goreleaser.yml` for release builds:

- **Compression level**: `best` with LZMA
- **Platforms**: Linux and Windows (all architectures)
- **macOS**: Excluded due to code signing complications
- **Size reduction**: ~70% additional reduction

#### UPX Trade-offs

**Pros:**
- Significant size reduction (70%+ compression)
- Binaries decompress automatically at runtime
- No functional changes to the program

**Cons:**
- Adds ~10-20ms to startup time (decompression overhead)
- May trigger false positives in some antivirus software
- Not compatible with macOS code signing without additional steps

#### Installing UPX

For local testing or CI/CD:

```bash
# Ubuntu/Debian
sudo apt-get install upx-ucl

# macOS
brew install upx

# Or download from GitHub
wget https://github.com/upx/upx/releases/download/v4.2.4/upx-4.2.4-amd64_linux.tar.xz
tar -xf upx-4.2.4-amd64_linux.tar.xz
sudo cp upx-4.2.4-amd64_linux/upx /usr/local/bin/
```

## Building Optimized Binaries

### Local Development

```bash
# Standard optimized build (18 MB)
make build

# With UPX compression (5.4 MB)
make build
upx --best --lzma gzcli
```

### Release Builds

Release builds automatically include all optimizations:

```bash
# Create a release build with goreleaser
make release
```

The resulting binaries in `dist/` will be:
- Optimized with build flags
- UPX compressed (Linux/Windows only)
- Ready for distribution

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

## Future Optimization Opportunities

Additional optimizations that could be explored:

1. **Build Tags**: Create separate builds for optional features
   - `gzcli-lite`: Core features only
   - `gzcli-full`: All features including proxy, email, etc.

2. **Dependency Analysis**: Review and potentially replace heavy dependencies
   - `github.com/lqqyt2423/go-mitmproxy`: Consider conditional compilation
   - `github.com/mattn/go-sqlite3`: Evaluate pure-Go alternatives like `modernc.org/sqlite`
   - QUIC libraries: Check if needed for all use cases

3. **Dead Code Elimination**: Use `-tags` to exclude unused code paths

4. **Further Compression**: Experiment with alternative compression tools
   - `gzexe`: GNU compression wrapper
   - `appimage`: Self-contained application format

## CI/CD Integration

The optimizations are automatically applied in CI/CD:

- GitHub Actions workflows use the optimized `.goreleaser.yml` configuration
- UPX is installed in the release workflow
- All release artifacts are optimized

## Troubleshooting

### UPX Antivirus Issues

Some antivirus software may flag UPX-compressed binaries. To resolve:

1. Build without UPX for affected users:
   ```bash
   make build  # Creates non-UPX binary
   ```

2. Submit false positive reports to antivirus vendors

3. Use code signing (recommended for production releases)

### macOS Code Signing

UPX-compressed binaries cannot be code-signed on macOS. Options:

1. Skip UPX for macOS (current configuration)
2. Use universal binaries without UPX
3. Decompress, sign, then recompress (advanced)

### Startup Time Concerns

If the ~10-20ms startup overhead from UPX is problematic:

1. Use non-UPX builds for time-critical applications
2. Consider the trade-off: 5.4 MB vs 18 MB
3. Profile actual startup time in your environment

## References

- [Go Build Modes](https://pkg.go.dev/cmd/go)
- [UPX Documentation](https://upx.github.io/)
- [GoReleaser UPX Integration](https://goreleaser.com/customization/upx/)
