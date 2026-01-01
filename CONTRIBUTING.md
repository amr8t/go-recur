# Contributing to go-recur

Thank you for your interest in contributing to go-recur! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Code Quality](#code-quality)
- [Submitting Changes](#submitting-changes)
- [CI/CD Pipeline](#cicd-pipeline)
- [Release Process](#release-process)

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-recur.git
   cd go-recur
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/amr8t/go-recur.git
   ```

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- (Optional) golangci-lint for local linting

### Install Dependencies

```bash
go mod download
```

### Install Development Tools

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Making Changes

1. Create a new branch for your feature or bug fix:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. Make your changes following our coding standards (see below)

3. Write tests for your changes

4. Ensure all tests pass:
   ```bash
   go test ./...
   ```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run tests with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Writing Tests

- Place tests in `*_test.go` files
- Use table-driven tests where appropriate
- Aim for high test coverage (>80%)
- Test edge cases and error conditions
- Use descriptive test names

Example:
```go
func TestRetrier_BasicRetry(t *testing.T) {
    tests := []struct {
        name        string
        maxAttempts int
        shouldFail  bool
        wantErr     bool
    }{
        {"success on first try", 3, false, false},
        {"success after retries", 3, true, false},
        {"failure after max attempts", 2, true, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Code Quality

### Code Style

- Follow standard Go conventions and idioms
- Run `gofmt` before committing:
  ```bash
  gofmt -s -w .
  ```
- Run `goimports` to organize imports:
  ```bash
  goimports -w .
  ```

### Linting

Run golangci-lint locally before pushing:

```bash
golangci-lint run
```

Fix any issues reported by the linter.

### Documentation

- Add godoc comments for all exported functions, types, and constants
- Use complete sentences in comments
- Include examples where appropriate
- Update README.md if adding new features

Example:
```go
// MaxAttempts creates a policy that sets the maximum number of retry attempts.
// The operation will be attempted at most 'max' times (including the initial attempt).
//
// Example:
//   err := recur.Do(operation).
//       WithMaxAttempts(5).
//       Run()
func MaxAttempts(max int) Policy {
    // Implementation
}
```

## Submitting Changes

### Before Submitting

Ensure your changes pass all checks:

```bash
# Format code
gofmt -s -w .

# Run tests
go test -race ./...

# Run linter
golangci-lint run

# Run go vet
go vet ./...

# Build examples
cd examples/simple && go build
cd ../http && go build
```

### Pull Request Process

1. Update your branch with the latest upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. Push your changes to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

3. Open a Pull Request on GitHub with:
   - Clear title describing the change
   - Detailed description of what changed and why
   - Reference to any related issues
   - Screenshots/examples if applicable

4. Address review feedback:
   - Make requested changes
   - Push updates to the same branch
   - The PR will automatically update

### Pull Request Template

```markdown
## Description
[Describe your changes here]

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] All tests pass
- [ ] Added new tests for changes
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No new warnings generated
```

## CI/CD Pipeline

All pull requests and commits to main branches trigger automated CI checks:

### CI Jobs

1. **Test Job** (runs on Go 1.21, 1.22, 1.23)
   - Downloads dependencies
   - Runs tests with race detection
   - Generates coverage reports
   - Uploads coverage to Codecov

2. **Lint Job**
   - Runs `go vet`
   - Checks code formatting with `gofmt`
   - Runs `golangci-lint` with comprehensive checks

3. **Build Job**
   - Builds the library
   - Builds all examples

4. **Examples Job**
   - Runs example programs to ensure they work

### CI Badges

You can add these badges to see CI status:

```markdown
[![CI](https://github.com/amr8t/go-recur/actions/workflows/ci.yml/badge.svg)](https://github.com/amr8t/go-recur/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/amr8t/go-recur)](https://goreportcard.com/report/github.com/amr8t/go-recur)
[![codecov](https://codecov.io/gh/amr8t/go-recur/branch/main/graph/badge.svg)](https://codecov.io/gh/amr8t/go-recur)
```

## Release Process

Releases are automated through GitHub Actions when version tags are pushed.

### Creating a Release

1. **Update Version Documentation**
   - Update CHANGELOG.md with new version changes
   - Update version references in documentation if needed

2. **Create and Push Tag**
   ```bash
   # Create annotated tag
   git tag -a v0.2.0 -m "Release v0.2.0"
   
   # Push tag to trigger release
   git push upstream v0.2.0
   ```

3. **Automated Release Workflow**
   - Runs all tests
   - Builds examples
   - Generates changelog from git commits
   - Creates GitHub release
   - Triggers pkg.go.dev update
   - Notifies about successful release

### Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version (v1.0.0 → v2.0.0): Incompatible API changes
- **MINOR** version (v0.1.0 → v0.2.0): New functionality, backward compatible
- **PATCH** version (v0.1.0 → v0.1.1): Bug fixes, backward compatible

### Pre-release Versions

For pre-releases, append a suffix:
- Alpha: `v0.2.0-alpha.1`
- Beta: `v0.2.0-beta.1`
- Release Candidate: `v0.2.0-rc.1`

These will be marked as pre-releases on GitHub automatically.

## Questions?

- Open an issue for bugs or feature requests
- Start a discussion for questions or ideas
- Reach out to maintainers for guidance

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to:

- Be respectful and inclusive
- Accept constructive criticism gracefully
- Focus on what's best for the community
- Show empathy towards other community members

Thank you for contributing to go-recur! 🎉