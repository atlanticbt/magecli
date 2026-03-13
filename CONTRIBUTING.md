# Contributing to magecli

Thanks for your interest in contributing! This guide will help you get started.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/magecli.git`
3. Create a branch: `git checkout -b my-feature`
4. Make your changes
5. Submit a pull request

## Development Setup

**Requirements:** Go 1.25+

```bash
make build        # Build binary to bin/magecli
make test         # Run tests
make fmt          # Format code
make lint         # Run linter (requires golangci-lint)
make tidy         # Tidy modules
make vulncheck    # Check for vulnerabilities
```

## Pull Request Guidelines

- Keep PRs focused on a single change
- Add tests for new functionality
- Run `make test` and `make lint` before submitting
- Follow existing code patterns and conventions
- Update documentation if you change commands, flags, or behavior (see Documentation Checklist in CLAUDE.md)

## Reporting Bugs

Open a [GitHub Issue](https://github.com/atlanticbt/magecli/issues/new?template=bug_report.md) with:

- magecli version (`magecli --version`)
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior

## Suggesting Features

Open a [GitHub Issue](https://github.com/atlanticbt/magecli/issues/new?template=feature_request.md) describing:

- The problem you're trying to solve
- Your proposed solution
- Any alternatives you've considered

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use table-driven tests with `t.Run()` subtests
- No mocking frameworks — use `t.TempDir()` for config isolation
- Keep commands read-only (GET only) unless there's a strong reason otherwise

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
