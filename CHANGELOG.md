# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - TBD

### Added
- Iterator-native retry library for Go 1.23+
- Native `for...range` pattern with `Iter().Seq()` using Go 1.23 iterators
- Fluent builder API for retry configuration
- Automatic retry control with `Attempt.Result(err)` method
- Built-in metrics collection with `MetricsCollector` for observability
- Five backoff strategies:
  - Constant - Fixed delay between retries
  - Exponential - Exponentially increasing delays
  - Fibonacci - Delays following Fibonacci sequence
  - Linear - Linearly increasing delays
  - NoDelay - Immediate retry with no delay
- Rich error matching system:
  - `MatchAny` - Retry all errors
  - `MatchErrors` - Match specific error values
  - `MatchFunc` - Custom error matching logic
  - Combinators: `And`, `Or`, `Not` for complex conditions
- Context support with cancellation and timeout
- Overall timeout configuration with `WithTimeout`
- Manual retry control with `ShouldRetry()` method (optional)
- Comprehensive test suite with >90% coverage
- Complete documentation and examples:
  - Basic retry patterns
  - Backoff strategies
  - Error-specific handling
  - HTTP client examples
  - Database with fallback
  - Circuit breaker integration
- CI/CD pipeline with GitHub Actions:
  - Multi-version Go testing (1.23, 1.24)
  - golangci-lint integration
  - Code coverage reporting with codecov
  - Example builds
- MIT License

### Notes
- Requires Go 1.23 or later for iterator support
- Iterator-first design for idiomatic Go 1.23+ usage
- Automatic metrics tracking with minimal boilerplate
- Explicit error handling over implicit callbacks

[0.1.0]: https://github.com/amr8t/go-recur/releases/tag/v0.1.0