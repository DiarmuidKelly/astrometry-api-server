# Contributing to Astrometry API Server

Thank you for your interest in contributing to Astrometry API Server! This document provides guidelines and workflows for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Pull Request Process](#pull-request-process)
- [Commit Message Convention](#commit-message-convention)
- [Testing Guidelines](#testing-guidelines)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a code of conduct that promotes a welcoming and inclusive environment. Please be respectful and considerate in all interactions.

## Getting Started

### Prerequisites

- **Go 1.21 or higher**
- **Docker** with `dm90/astrometry` image
- **git** for version control
- **make** for build automation

### Setup Development Environment

```bash
# Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/Astrometry-API-Server.git
cd Astrometry-API-Server

# Download dependencies
go mod download

# Run tests to verify setup
make test

# Build the server
make build
```

## Development Workflow

1. **Create a feature branch**
   ```bash
   git checkout -b feat/my-feature
   # or
   git checkout -b fix/my-bugfix
   ```

2. **Make your changes**
   - Write clean, idiomatic Go code
   - Follow existing code style and patterns
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**
   ```bash
   # Run tests
   make test

   # Run linter
   make lint

   # Check formatting
   go fmt ./...
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

5. **Push to your fork**
   ```bash
   git push origin feat/my-feature
   ```

## Pull Request Process

### PR Title Format

**IMPORTANT:** PR titles must follow one of these formats:

#### Version Bump (Triggers Release)

```
[MAJOR] Breaking change description
[MINOR] New feature description
[PATCH] Bug fix description
```

**Examples:**
- `[MAJOR] Redesign API endpoints with breaking changes`
- `[MINOR] Add batch solving endpoint`
- `[PATCH] Fix CORS header configuration`

#### Conventional Commits (No Release)

```
<type>: <description>
<type>(<scope>): <description>
```

**Types:**
- `feat:` - New feature (no release)
- `fix:` - Bug fix (no release)
- `docs:` - Documentation updates
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `perf:` - Performance improvements
- `test:` - Test additions or updates
- `build:` - Build system changes
- `ci:` - CI/CD pipeline changes
- `chore:` - Other changes (dependencies, etc.)

**Examples:**
- `docs: Update API examples in README`
- `test: Add integration tests for health endpoint`
- `refactor: Simplify error handling in solve handler`

#### Skip Release

```
[skip-release] Description
```

Use this for changes that should not trigger a release (similar to conventional commits).

### PR Description

Include in your PR description:

- **Summary**: Brief description of changes
- **Motivation**: Why is this change needed?
- **Changes**: List of specific changes made
- **Testing**: How have you tested these changes?
- **Related Issues**: Link any related issues

**Template:**

```markdown
## Summary
Brief description of the change.

## Motivation
Why is this change necessary? What problem does it solve?

## Changes
- Change 1
- Change 2
- Change 3

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed
- [ ] Linter passes

## Related Issues
Fixes #123
```

### PR Checklist

Before submitting, ensure:

- [ ] Code follows Go style guidelines
- [ ] Tests added/updated and passing
- [ ] Documentation updated (README, code comments)
- [ ] PR title follows required format
- [ ] Commits are clean and atomic
- [ ] No merge conflicts with main branch
- [ ] Linter passes without errors

## Commit Message Convention

While PR titles drive the release process, individual commits should still follow best practices:

### Good Commit Messages

```bash
# Good - Clear and descriptive
git commit -m "feat: add timeout parameter to solve endpoint"
git commit -m "fix: handle nil pointer in health handler"
git commit -m "docs: add curl examples to README"

# Bad - Vague or too brief
git commit -m "update"
git commit -m "fix bug"
git commit -m "changes"
```

### Commit Best Practices

- **Atomic commits**: Each commit should represent one logical change
- **Descriptive messages**: Explain what and why, not how
- **Present tense**: "Add feature" not "Added feature"
- **Keep commits focused**: Don't mix refactoring with features

## Testing Guidelines

### Writing Tests

- Place tests in `*_test.go` files
- Use table-driven tests for multiple scenarios
- Test both success and error cases
- Mock external dependencies (Docker, filesystem)

**Example:**

```go
func TestSolveHandler(t *testing.T) {
    tests := []struct {
        name       string
        imageFile  string
        wantSolved bool
        wantErr    bool
    }{
        {
            name:       "valid image",
            imageFile:  "test.jpg",
            wantSolved: true,
            wantErr:    false,
        },
        {
            name:       "invalid format",
            imageFile:  "test.txt",
            wantSolved: false,
            wantErr:    true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Specific package
go test ./internal/handlers/...

# Verbose output
go test -v ./...
```

## Release Process

The project uses **automated releases** via GitHub Actions. The release process is triggered when a PR is merged to `main`.

### How Releases Work

1. **PR Title** determines the version bump:
   - `[MAJOR]` → Major version (e.g., 1.0.0 → 2.0.0)
   - `[MINOR]` → Minor version (e.g., 1.0.0 → 1.1.0)
   - `[PATCH]` → Patch version (e.g., 1.0.0 → 1.0.1)

2. **On PR merge**, the workflow:
   - Runs `scripts/auto-release.sh` to bump version
   - Updates `VERSION` and `CHANGELOG.md`
   - Creates a git commit and tag
   - Builds multi-platform binaries
   - Creates a GitHub release with artifacts
   - Pushes Docker images to GitHub Container Registry

3. **Skip release** by using:
   - `[skip-release]` prefix
   - Conventional commit types: `docs:`, `test:`, `chore:`, etc.

### Manual Release (Maintainers Only)

```bash
# Bump version
./scripts/auto-release.sh [major|minor|patch]

# Push changes
git push origin main
git push origin vX.Y.Z
```

## Code Style

### Go Style Guidelines

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` and `goimports` for formatting
- Keep functions focused and concise
- Avoid deep nesting (max 3-4 levels)
- Use meaningful variable names
- Document exported functions and types

### Code Organization

```
internal/
├── handlers/       # HTTP request handlers
├── middleware/     # HTTP middleware (logging, CORS)
└── ...
```

- `internal/` - Private application code
- `cmd/` - Application entry points
- `pkg/` - Public libraries (if any)

## Questions?

If you have questions or need help:

1. Check existing [Issues](https://github.com/DiarmuidKelly/Astrometry-API-Server/issues)
2. Create a new issue with the `question` label
3. Reach out to maintainers

## License

By contributing, you agree that your contributions will be licensed under the GNU General Public License v3.0.
