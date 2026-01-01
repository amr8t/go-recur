# Release Guide

This document describes how to create and publish a new release of go-recur.

## Prerequisites

- Write access to the repository
- All changes merged to main branch
- All tests passing on main
- CHANGELOG.md updated (optional, will auto-generate)

## Release Process

### 1. Determine Version Number

Follow [Semantic Versioning](https://semver.org/):

- **Patch release** (e.g., v0.1.0 → v0.1.1): Bug fixes, no API changes
- **Minor release** (e.g., v0.1.0 → v0.2.0): New features, backward compatible
- **Major release** (e.g., v0.1.0 → v1.0.0): Breaking API changes

### 2. Update Documentation (Optional)

If you want a custom changelog, update it before tagging:

```bash
# Edit CHANGELOG.md
vim CHANGELOG.md

# Commit changes
git add CHANGELOG.md
git commit -m "docs: update changelog for vX.Y.Z"
git push origin main
```

### 3. Create and Push Tag

```bash
# Ensure you're on main and up to date
git checkout main
git pull origin main

# Create annotated tag
git tag -a v0.2.0 -m "Release v0.2.0: Add jitter support and improve error matching"

# Push tag to GitHub (this triggers the release workflow)
git push origin v0.2.0
```

### 4. Automated Release

Once the tag is pushed, GitHub Actions automatically:

1. ✅ Runs all tests
2. ✅ Runs linters and code quality checks
3. ✅ Builds all examples
4. ✅ Generates changelog from git commits
5. ✅ Creates GitHub release
6. ✅ Triggers pkg.go.dev indexing
7. ✅ Publishes release notes

### 5. Verify Release

After ~2-5 minutes, verify:

1. **GitHub Release**: https://github.com/amr8t/go-recur/releases
   - Check that release was created
   - Verify changelog is correct
   - Ensure it's not marked as pre-release (unless intended)

2. **pkg.go.dev**: https://pkg.go.dev/github.com/amr8t/go-recur
   - Wait 10-15 minutes for indexing
   - Verify new version appears in version dropdown
   - Check documentation rendered correctly

3. **Test Installation**:
   ```bash
   # In a test directory
   go get github.com/amr8t/go-recur@v0.2.0
   ```

## Pre-releases

For alpha, beta, or release candidate versions:

```bash
# Alpha
git tag -a v0.2.0-alpha.1 -m "Pre-release: v0.2.0-alpha.1"
git push origin v0.2.0-alpha.1

# Beta
git tag -a v0.2.0-beta.1 -m "Pre-release: v0.2.0-beta.1"
git push origin v0.2.0-beta.1

# Release Candidate
git tag -a v0.2.0-rc.1 -m "Pre-release: v0.2.0-rc.1"
git push origin v0.2.0-rc.1
```

These will automatically be marked as pre-releases on GitHub.

## Rollback a Release

If you need to remove a release:

1. **Delete GitHub Release**: Go to releases page and delete the release
2. **Delete Tag Locally**:
   ```bash
   git tag -d v0.2.0
   ```
3. **Delete Tag Remotely**:
   ```bash
   git push --delete origin v0.2.0
   ```

**Note**: You cannot unpublish from pkg.go.dev once indexed. Instead, publish a new patch version.

## Release Checklist

Before creating a release, ensure:

- [ ] All tests pass locally: `go test -race ./...`
- [ ] Code is formatted: `gofmt -s -w .`
- [ ] No lint errors: `golangci-lint run` (if installed)
- [ ] Examples build and run: `cd examples/simple && go run main.go`
- [ ] Documentation is up to date
- [ ] Breaking changes are documented (if any)
- [ ] Migration guide included (for major versions)

After creating a release:

- [ ] GitHub release created successfully
- [ ] Release notes are accurate
- [ ] pkg.go.dev shows new version (wait 10-15 min)
- [ ] Installation works: `go get github.com/amr8t/go-recur@vX.Y.Z`

## Troubleshooting

### Release Workflow Failed

1. Check GitHub Actions: https://github.com/amr8t/go-recur/actions
2. Review error logs
3. Fix issues and re-run workflow or create new tag

### Tag Already Exists

```bash
# Delete and recreate
git tag -d v0.2.0
git push --delete origin v0.2.0
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

### pkg.go.dev Not Updating

1. Wait 15-30 minutes (indexing can be slow)
2. Manually request indexing:
   ```bash
   curl https://proxy.golang.org/github.com/amr8t/go-recur/@v/v0.2.0.info
   ```
3. Check for errors: https://pkg.go.dev/github.com/amr8t/go-recur?tab=versions

## Version History

| Version | Date | Highlights |
|---------|------|------------|
| v0.1.0  | TBD  | Initial release with fluent API, composable policies, error matching |

## Questions?

Open an issue or discussion on GitHub: https://github.com/amr8t/go-recur/issues