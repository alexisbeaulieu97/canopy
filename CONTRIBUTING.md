# Contributing to Canopy

Thank you for your interest in contributing to Canopy! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Development Environment Setup](#development-environment-setup)
- [Build Instructions](#build-instructions)
- [Testing Guidelines](#testing-guidelines)
- [Code Style](#code-style)
- [Pull Request Process](#pull-request-process)
- [Project Architecture](#project-architecture)

## Development Environment Setup

### Prerequisites

- **Go 1.21+** - [Install Go](https://go.dev/doc/install)
- **Git** - For version control
- **Make** - For running build commands (optional but recommended)

### Getting Started

1. **Fork the repository** on GitHub

2. **Clone your fork**:
   ```bash
   git clone https://github.com/<your-username>/canopy.git
   cd canopy
   ```

3. **Install development tools**:
   ```bash
   make tools
   ```

4. **Verify setup**:
   ```bash
   make test
   make lint
   ```

### IDE Setup

For VS Code, recommended extensions:
- Go (official extension)
- EditorConfig

For GoLand/IntelliJ:
- Enable "Go Modules integration"
- Configure the project GOPATH

## Build Instructions

### Using Make (Recommended)

```bash
# Build the binary with version info
make build

# Install to $GOPATH/bin
make install

# Clean build artifacts
make clean
```

### Manual Build

```bash
# Simple build
go build -o canopy ./cmd/canopy

# Build with version info
go build -ldflags "-X main.version=$(git describe --tags)" -o canopy ./cmd/canopy
```

### Cross-Compilation

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 make build

# Build for macOS ARM
GOOS=darwin GOARCH=arm64 make build
```

## Testing Guidelines

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/workspaces/...

# Run with verbose output
go test -v ./...
```

### Writing Tests

1. **Test file naming**: Use `*_test.go` suffix
2. **Table-driven tests**: Preferred for testing multiple cases
3. **Mocks**: Use interfaces and dependency injection
4. **Test isolation**: Each test should be independent

Example test structure:

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "valid input",
            input: "test",
            want:  "expected",
        },
        {
            name:    "invalid input",
            input:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

Integration tests are located in `test/integration/`. They test the full application flow and require a git environment.

```bash
# Run integration tests
make test-integration
```

### Test Helpers

The `internal/testutil/` package provides shared test utilities to avoid code duplication across test files:

**Git Helpers** (`testutil/git.go`):
- `CreateRepoWithCommit(t, path)` - Initialize a git repo with an initial commit
- `RunGit(t, dir, args...)` - Execute a git command, fail on error
- `RunGitOutput(t, dir, args...)` - Execute git and return output
- `CloneToBare(t, source, dest)` - Clone a repo as bare

**Filesystem Helpers** (`testutil/fs.go`):
- `MustMkdir(t, path)` - Create directory, fail on error
- `MustWriteFile(t, path, content)` - Write file, fail on error
- `MustReadFile(t, path)` - Read file, fail on error
- `MustTempDir(t, pattern)` - Create temp dir with cleanup

Example usage:

```go
import "github.com/alexisbeaulieu97/canopy/internal/testutil"

func TestSomething(t *testing.T) {
    dir := testutil.MustTempDir(t, "test-")
    testutil.CreateRepoWithCommit(t, dir)
    testutil.RunGit(t, dir, "status")
}
```

## Code Style

### Go Conventions

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` or `goimports` for formatting
- Run `golangci-lint` before committing

### Linting

```bash
# Run linter
make lint

# Auto-fix issues
golangci-lint run --fix
```

### Naming Conventions

- **Packages**: Short, lowercase, no underscores (e.g., `workspaces`, `gitx`)
- **Exported functions**: PascalCase with clear names
- **Private functions**: camelCase
- **Interfaces**: Verb phrases for behavior (e.g., `ConfigProvider`, `GitOperations`)
- **Errors**: Prefix with `Err` (e.g., `ErrWorkspaceNotFound`)

### Documentation

- Add godoc comments to all exported types and functions
- Include examples for complex APIs
- Keep comments concise but informative

Example:

```go
// CreateWorkspace creates a new workspace with the given ID and repositories.
// It creates the workspace directory, clones all repositories, and runs
// post_create hooks if configured.
//
// Returns the created directory name and any error encountered.
func (s *Service) CreateWorkspace(ctx context.Context, id string, repos []domain.Repo) (string, error) {
    // implementation
}
```

### Error Handling

- Use typed errors from `internal/errors`
- Wrap errors with context using `cerrors.Wrap()`
- Check errors explicitly, don't ignore them

```go
// Good
if err != nil {
    return cerrors.WrapGitError(err, "clone repository")
}

// Bad
_ = SomeFunction()
```

## Pull Request Process

### Before Submitting

1. **Create an issue** describing the change (for non-trivial changes)
2. **Branch from main**: `git checkout -b feature/your-feature`
3. **Make focused commits** with clear messages
4. **Run tests**: `make test`
5. **Run linter**: `make lint`
6. **Update documentation** if needed

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Formatting, no code change
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding or updating tests
- `chore`: Build process or auxiliary tool changes

Examples:
```
feat(workspace): add rename command
fix(git): handle timeout in clone operation
docs: update README with new commands
```

### Review Process

1. Create a pull request with a clear description
2. Link related issues
3. Respond to review feedback
4. Squash commits if requested
5. Ensure CI passes

## Project Architecture

Canopy follows hexagonal architecture (ports and adapters):

```
internal/
├── app/          # Application container
├── config/       # Configuration loading
├── domain/       # Core domain models
├── errors/       # Typed error definitions
├── gitx/         # Git operations (adapter)
├── hooks/        # Hook execution (adapter)
├── ports/        # Interface definitions
├── storage/      # Workspace storage (adapter)
├── workspaces/   # Business logic (core)
└── validation/   # Input validation
```

See [Architecture Documentation](docs/architecture.md) for details.

## Questions?

- Open an issue for bugs or feature requests
- Start a discussion for questions or ideas

Thank you for contributing!
