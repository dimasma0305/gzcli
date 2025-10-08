# Automatic Semantic Versioning

This document describes the automated semantic versioning system implemented in gzcli.

## Overview

gzcli uses **automated semantic versioning** based on conventional commit messages. When you push commits to the `main` branch, the system automatically:
- Analyzes commit messages
- Determines the appropriate version bump
- Creates a git tag
- Generates a changelog
- Publishes a GitHub release with built binaries

## How It Works

### 1. Commit Message Format

All commits must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### 2. Version Bumping Rules

The commit type determines the version bump:

| Commit Type | Version Bump | Example |
|-------------|--------------|---------|
| `feat:` | **Minor** (1.x.0) | `feat: add new command` |
| `fix:` | **Patch** (1.0.x) | `fix: resolve crash on startup` |
| `perf:` | **Patch** (1.0.x) | `perf: optimize file watcher` |
| `refactor:` | **Patch** (1.0.x) | `refactor: simplify API client` |
| `feat!:` or `BREAKING CHANGE:` | **Major** (x.0.0) | `feat!: remove deprecated API` |
| `docs:`, `test:`, `chore:`, `ci:` | **No bump** | `docs: update README` |

### 3. Workflow

```
┌───────────────────────────────────────────────────────────────┐
│  Developer                                                     │
│  ├─ Write code                                                 │
│  ├─ Commit with conventional format: "feat: add feature"      │
│  └─ Push to main branch                                       │
└─────────────────────────┬─────────────────────────────────────┘
                          │
                          ▼
┌───────────────────────────────────────────────────────────────┐
│  GitHub Actions (Release Workflow)                            │
│  ├─ Runs tests                                                │
│  ├─ Analyzes commits since last release                       │
│  ├─ Determines version bump (major/minor/patch)               │
│  ├─ Creates git tag (e.g., v1.2.3)                            │
│  ├─ Generates CHANGELOG.md                                    │
│  ├─ Builds binaries for all platforms with GoReleaser        │
│  │   - Linux (amd64, arm64, arm)                             │
│  │   - macOS (Universal Binary)                              │
│  │   - Windows (amd64)                                       │
│  ├─ Injects version metadata into binaries                   │
│  │   - Version: v1.2.3                                       │
│  │   - Commit: abc1234                                       │
│  │   - BuildTime: 2025-10-07_12:34:56                       │
│  └─ Publishes GitHub release with changelog and binaries     │
└───────────────────────────────────────────────────────────────┘
```

## Version Metadata

Each built binary contains embedded version information accessible via `--version`:

```bash
$ gzcli --version
gzcli version v1.2.3
Commit: abc1234
Built: 2025-10-07_12:34:56
```

### How Version Information is Injected

#### During Local Build (via Makefile)

```makefile
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')

LDFLAGS=-ldflags "\
  -X github.com/dimasma0305/gzcli/cmd.Version=${VERSION} \
  -X github.com/dimasma0305/gzcli/cmd.BuildTime=${BUILD_TIME} \
  -X github.com/dimasma0305/gzcli/cmd.GitCommit=${GIT_COMMIT}"
```

#### During Release Build (via GoReleaser)

```yaml
ldflags:
  - -s -w
  - -X github.com/dimasma0305/gzcli/cmd.Version={{.Version}}
  - -X github.com/dimasma0305/gzcli/cmd.BuildTime={{.Date}}
  - -X github.com/dimasma0305/gzcli/cmd.GitCommit={{.Commit}}
```

#### In Code (cmd/root.go)

```go
var (
    Version   = "dev"
    GitCommit = "unknown"
    BuildTime = "unknown"
)

var rootCmd = &cobra.Command{
    Version: func() string {
        if GitCommit != "unknown" && BuildTime != "unknown" {
            return Version + "\nCommit: " + GitCommit + "\nBuilt: " + BuildTime
        }
        return Version
    }(),
}
```

## Examples

### Feature Addition (Minor Version Bump)

```bash
git commit -m "feat(watch): add support for custom ignore patterns

Add ability to specify custom file patterns to ignore in the watcher.
This allows users to exclude temporary files and build artifacts.

Closes #123"
```

**Result:** Version bumps from `1.2.3` → `1.3.0`

### Bug Fix (Patch Version Bump)

```bash
git commit -m "fix(sync): handle empty challenge directories

Previously, the sync command would fail when encountering empty
challenge directories. Now it skips them with a warning.

Fixes #456"
```

**Result:** Version bumps from `1.3.0` → `1.3.1`

### Breaking Change (Major Version Bump)

```bash
git commit -m "feat(api)!: change authentication to use OAuth2

BREAKING CHANGE: Basic authentication is no longer supported.
All users must migrate to OAuth2 authentication.

Refs #789"
```

**Result:** Version bumps from `1.3.1` → `2.0.0`

### Documentation Update (No Version Bump)

```bash
git commit -m "docs: update installation instructions

Add clarification about Go version requirements."
```

**Result:** No version bump, no release

## Manual Release (Alternative)

If you need to create a release manually without semantic-release:

```bash
# Create and push a tag manually
git tag v1.2.3
git push origin v1.2.3

# This triggers the release workflow directly
```

## Initial Setup

If starting from scratch (first release):

1. Create an initial tag:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

2. From that point forward, semantic-release will handle all versioning automatically

## Troubleshooting

### No Release Created After Pushing to Main

**Possible causes:**
1. No conventional commits since last release
2. Only commits that don't trigger releases (docs, chore, test, etc.)
3. Workflow disabled or failing

**Solution:**
Check the GitHub Actions tab for workflow runs and logs.

### Version Shows "dev" Locally

**Cause:** Building without git tags or in a dirty working directory.

**Solution:**
```bash
# Ensure you have git tags
git fetch --tags

# Or build with explicit version
make build VERSION=v1.2.3
```

### Multiple Versions Released at Once

**Cause:** Multiple feature/fix commits pushed together.

**Solution:** This is expected behavior. Semantic-release analyzes all commits and determines the highest applicable version bump.

## Best Practices

1. **Commit often with proper types:** Each logical change should be a separate commit
2. **Use descriptive scopes:** Helps organize changelog (e.g., `feat(watch):`, `fix(api):`)
3. **Reference issues:** Link commits to issues using `Fixes #123`, `Closes #456`
4. **Squash feature branches:** When merging PRs, ensure the squashed commit message follows conventions
5. **Review before merging to main:** Remember that every merge to main can trigger a release

## Configuration Files

- `.github/workflows/semantic-release.yml` - Combined semantic release automation and GoReleaser build/publish
- `.goreleaser.yml` - GoReleaser configuration
- `CONTRIBUTING.md` - Commit message guidelines for contributors
- `cmd/root.go` - Version variables definition
- `Makefile` - Local build with version injection

## Further Reading

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [Semantic Release](https://semantic-release.gitbook.io/semantic-release/)
- [GoReleaser Documentation](https://goreleaser.com/)
