# CI/CD Setup Documentation

This document describes the CI/CD infrastructure for go-recur.

## Overview

The project uses GitHub Actions for continuous integration and automated releases. The setup includes:

- **Continuous Integration**: Automated testing, linting, and building on every push and pull request
- **Automated Releases**: Tag-based releases with automatic changelog generation and publishing
- **Code Quality**: Comprehensive linting and formatting checks
- **Multi-version Testing**: Tests run on Go 1.21, 1.22, and 1.23

## Workflows

### 1. CI Workflow (`.github/workflows/ci.yml`)

**Triggers:**
- Push to `main`, `master`, or `develop` branches
- Pull requests to `main`, `master`, or `develop` branches

**Jobs:**

#### Test Job
- Runs on Go 1.21, 1.22, and 1.23 (matrix)
- Downloads dependencies with caching
- Runs tests with race detection: `go test -v -race -coverprofile=coverage.out`
- Uploads coverage to Codecov (Go 1.23 only)

#### Lint Job
- Runs `go vet` to check for suspicious constructs
- Verifies code formatting with `gofmt`
- Runs `golangci-lint` with comprehensive checks (see `.golangci.yml`)

#### Build Job
- Builds the main library
- Builds both example programs (`simple` and `http`)

#### Examples Job
- Runs example programs to ensure they execute without errors
- Uses timeout to prevent hanging

**Duration:** ~2-5 minutes

### 2. Release Workflow (`.github/workflows/release.yml`)

**Triggers:**
- Push of tags matching pattern `v*.*.*` (e.g., `v0.1.0`, `v1.2.3`)

**Jobs:**

#### Release Job
- Checks out code with full history
- Runs all tests to verify stability
- Runs `go vet` for code quality
- Builds all examples
- Extracts version from tag
- Generates changelog from git commits since last tag
- Creates GitHub release with:
  - Version number
  - Changelog
  - Installation instructions
  - Documentation links
- Marks pre-releases automatically (tags with `-alpha`, `-beta`, `-rc`)

#### Publish Job
- Triggers pkg.go.dev indexing via Go proxy
- Ensures new version is discoverable

#### Notify Job
- Outputs release information
- Provides links to GitHub release and pkg.go.dev

**Duration:** ~3-7 minutes

## Configuration Files

### `.golangci.yml`

Comprehensive linting configuration with:

**Enabled Linters:**
- `errcheck` - Unchecked errors
- `gosimple` - Code simplification
- `govet` - Suspicious constructs
- `ineffassign` - Ineffectual assignments
- `staticcheck` - Advanced analysis
- `unused` - Unused code detection
- `gofmt` - Formatting
- `goimports` - Import organization
- `misspell` - Spelling
- `revive` - Fast, configurable linting
- `gocritic` - Extensible checks
- `gocyclo` - Cyclomatic complexity
- `exportloopref` - Loop variable issues
- `gosec` - Security checks
- `typecheck` - Type checking

**Exclusions:**
- Test files: Less strict on `gocyclo`, `errcheck`, `gosec`
- Examples: Less strict on `errcheck`, `gosec`

**Settings:**
- 5-minute timeout
- Unlimited issues per linter
- All issues shown (not just new ones)

## Badges

Add to your README:

```markdown
[![CI](https://github.com/amr8t/go-recur/actions/workflows/ci.yml/badge.svg)](https://github.com/amr8t/go-recur/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/amr8t/go-recur)](https://goreportcard.com/report/github.com/amr8t/go-recur)
[![codecov](https://codecov.io/gh/amr8t/go-recur/branch/main/graph/badge.svg)](https://codecov.io/gh/amr8t/go-recur)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
```

## Local Development

### Running Tests Locally

```bash
# All tests
go test ./...

# With race detection (like CI)
go test -race ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Running Linters Locally

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter (same as CI)
golangci-lint run

# Run go vet
go vet ./...

# Check formatting
gofmt -l -s .

# Fix formatting
gofmt -w -s .
```

### Building Examples

```bash
# Simple example
cd examples/simple && go build

# HTTP example
cd examples/http && go build
```

### Pre-commit Checklist

Run these before pushing:

```bash
# Format code
gofmt -s -w .

# Run tests
go test -race ./...

# Run linter (if installed)
golangci-lint run

# Run go vet
go vet ./...

# Build examples
(cd examples/simple && go build) && (cd examples/http && go build)
```

## Creating a Release

### Step 1: Prepare

1. Ensure all changes are merged to `main`
2. Verify all CI checks pass on `main`
3. Update `CHANGELOG.md` (optional - will auto-generate)

### Step 2: Tag and Push

```bash
# Checkout main and update
git checkout main
git pull origin main

# Create annotated tag
git tag -a v0.2.0 -m "Release v0.2.0: Add new features"

# Push tag (triggers release workflow)
git push origin v0.2.0
```

### Step 3: Verify

After 5-10 minutes:

1. **GitHub Release**: Check https://github.com/amr8t/go-recur/releases
2. **pkg.go.dev**: Check https://pkg.go.dev/github.com/amr8t/go-recur (may take 15+ min)
3. **Installation**: Test `go get github.com/amr8t/go-recur@v0.2.0`

### Pre-release Versions

For testing before stable release:

```bash
# Alpha
git tag -a v0.2.0-alpha.1 -m "Pre-release: Alpha 1"
git push origin v0.2.0-alpha.1

# Beta
git tag -a v0.2.0-beta.1 -m "Pre-release: Beta 1"
git push origin v0.2.0-beta.1

# Release Candidate
git tag -a v0.2.0-rc.1 -m "Pre-release: RC 1"
git push origin v0.2.0-rc.1
```

These automatically get marked as pre-releases on GitHub.

## Troubleshooting

### CI Workflow Failing

**Test failures:**
1. Check test logs in GitHub Actions
2. Reproduce locally: `go test -race ./...`
3. Fix and push again

**Lint failures:**
1. Run locally: `golangci-lint run`
2. Fix issues reported
3. Ensure `gofmt -l .` returns empty

**Build failures:**
1. Check that examples still import correctly
2. Verify `go.mod` is up to date
3. Test builds locally

### Release Workflow Failing

**Tag already exists:**
```bash
# Delete and recreate
git tag -d v0.2.0
git push --delete origin v0.2.0
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

**Tests fail during release:**
1. Delete the tag
2. Fix tests on main
3. Create tag again

**pkg.go.dev not updating:**
1. Wait 15-30 minutes (can be slow)
2. Manually trigger:
   ```bash
   VERSION="v0.2.0"
   curl "https://proxy.golang.org/github.com/amr8t/go-recur/@v/${VERSION}.info"
   ```

### Codecov Not Working

1. Sign up at https://codecov.io with your GitHub account
2. Add your repository
3. No token needed for public repos
4. For private repos, add `CODECOV_TOKEN` secret to GitHub

## Security Considerations

### Secrets

No secrets are required for the basic CI/CD setup. Optional:

- `CODECOV_TOKEN`: Only needed for private repositories
- `GITHUB_TOKEN`: Automatically provided by GitHub Actions

### Permissions

The release workflow requires `contents: write` permission to create releases.
This is declared in the workflow file.

### Dependency Updates

Keep GitHub Actions up to date:
- `actions/checkout@v4` - Latest checkout action
- `actions/setup-go@v5` - Latest Go setup
- `golangci/golangci-lint-action@v4` - Latest lint action
- `softprops/action-gh-release@v1` - Release creation

## Maintenance

### Regular Tasks

**Weekly:**
- Monitor CI status
- Review and merge dependabot PRs (if enabled)

**Monthly:**
- Update GitHub Actions versions
- Check for new golangci-lint rules
- Review and update linter configuration

**Per Release:**
- Update CHANGELOG.md
- Test installation of new version
- Verify documentation on pkg.go.dev

### Monitoring

**GitHub Actions:**
- View workflow runs: https://github.com/amr8t/go-recur/actions
- Check status badges in README
- Enable notifications for failed workflows

**Go Report Card:**
- Monitor score: https://goreportcard.com/report/github.com/amr8t/go-recur
- Address issues that lower the grade

**Codecov:**
- Track coverage trends: https://codecov.io/gh/amr8t/go-recur
- Aim for >80% coverage

## Future Enhancements

### Potential Additions

1. **Benchmarking**
   - Add benchmark CI job
   - Track performance over time
   - Compare with other libraries

2. **Advanced Testing**
   - Fuzz testing
   - Integration tests
   - Compatibility testing

3. **Documentation**
   - Auto-generate API docs
   - Deploy examples to GitHub Pages
   - Create video tutorials

4. **Release Automation**
   - Auto-update changelog from PRs
   - Generate migration guides
   - Post announcements to social media

5. **Code Quality**
   - Add mutation testing
   - Enforce coverage thresholds
   - Add security scanning

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint Documentation](https://golangci-lint.run/)
- [Go Modules Reference](https://go.dev/ref/mod)
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)

## Support

For CI/CD issues:
1. Check this documentation
2. Review GitHub Actions logs
3. Open an issue with workflow run link
4. Check GitHub Actions status page

---

Last Updated: 2025-01-31