## Description

<!-- Provide a clear and concise description of your changes -->

## Type of Change

<!-- Mark the relevant option with an 'x' -->

- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Code refactoring
- [ ] CI/CD changes
- [ ] Other (please describe):

## Motivation and Context

<!-- Why is this change required? What problem does it solve? -->
<!-- If it fixes an open issue, please link to the issue here -->

Fixes #(issue)

## How Has This Been Tested?

<!-- Describe the tests you ran to verify your changes -->
<!-- Provide instructions so others can reproduce -->

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed
- [ ] Examples tested

**Test Configuration**:
- Go version:
- OS:

## Screenshots (if applicable)

<!-- Add screenshots to help explain your changes -->

## Checklist

<!-- Mark completed items with an 'x' -->

### Code Quality

- [ ] My code follows the style guidelines of this project
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] My changes generate no new warnings
- [ ] I have run `gofmt -s -w .` to format the code
- [ ] I have run `go vet ./...` with no errors

### Testing

- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] I have run `go test -race ./...` successfully
- [ ] Any dependent changes have been merged and published

### Documentation

- [ ] I have updated the documentation accordingly
- [ ] I have added/updated godoc comments for exported symbols
- [ ] I have updated the README.md (if needed)
- [ ] I have updated CHANGELOG.md

### Examples

- [ ] I have updated/added examples if this is a new feature
- [ ] All examples build successfully
- [ ] All examples run without errors

## Additional Notes

<!-- Add any other context about the pull request here -->

## Breaking Changes

<!-- If this is a breaking change, describe the impact and migration path -->

**Before:**
```go
// Old API usage
```

**After:**
```go
// New API usage
```

## Reviewer Notes

<!-- Anything specific you want reviewers to focus on? -->