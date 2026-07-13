# Contributing to groovy-check

Thanks for your interest in contributing! This guide will help you get started.

## Development Setup

### Prerequisites

- Go 1.25+
- [golangci-lint](https://golangci-lint.run/) (`brew install golangci-lint`)

### Getting Started

```bash
# Fork and clone the repository
git clone https://github.com/<your-username>/groovy-check.git
cd groovy-check

# Build
make build

# Run tests
make test

# Run linter
make lint
```

## Making Changes

1. **Create a branch** from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```

2. **Write your code** following existing patterns:
   - CLI commands go in `cmd/`
   - Keep each command's flags and logic in its own file (see `cmd/lint.go`, `cmd/validate.go`)

3. **Add tests** for any new functionality — tests live alongside the code as `*_test.go`.

4. **Ensure all checks pass**:
   ```bash
   make lint
   make test
   go build ./...
   ```

## Pull Request Guidelines

- Keep PRs focused on a single change
- Write a clear description of what and why
- Reference any related issues
- All CI checks must pass (build, vet, lint, test, gitleaks)
- A maintainer review is required before merge — this repo restricts merging
  to `main` to the maintainer, so don't expect self-service merge even after
  approval

## Code Style

- Formatter: `gofmt` / `go fmt ./...`
- Linter: golangci-lint (`.golangci.yaml` in repo root)
- Wrap errors with `fmt.Errorf("...: %w", err)` for CLI-shaped error context

## Reporting Issues

- **Bugs**: Use the [bug report template](https://github.com/aviadshiber/groovy-check/issues/new?template=bug_report.md)
- **Features**: Use the [feature request template](https://github.com/aviadshiber/groovy-check/issues/new?template=feature_request.md)
- **Security**: See [SECURITY.md](.github/SECURITY.md)

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
