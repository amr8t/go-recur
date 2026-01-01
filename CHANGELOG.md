# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Fluent builder API for retry configuration
- Composable policies with `CombinePolicies`
- Rich error matching with And/Or/Not combinators
- Multiple backoff strategies: Constant, Exponential, Fibonacci, Linear, NoDelay
- Context and timeout support
- Retry hooks for monitoring and observability
- Comprehensive test suite
- Complete documentation and examples
- CI/CD pipeline with GitHub Actions
- Automated release workflow

### Changed
- N/A

### Deprecated
- N/A

### Removed
- N/A

### Fixed
- N/A

### Security
- N/A

## [0.1.0] - TBD

### Added
- Initial release
- Core retry engine with error handling
- Fluent API with method chaining
- Policy composition system
- Error matchers (MatchAny, MatchErrors, MatchFunc)
- Error matcher combinators (And, Or, Not)
- Five backoff strategies
- Context support with cancellation
- Overall timeout configuration
- OnRetry hooks for monitoring
- Comprehensive examples (simple and HTTP)
- Full API documentation
- MIT License

[Unreleased]: https://github.com/amr8t/go-recur/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/amr8t/go-recur/releases/tag/v0.1.0