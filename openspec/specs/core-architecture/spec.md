# core-architecture Specification

## Purpose
Defines the core architecture patterns for Canopy including service initialization, dependency injection via App context, command registration, hexagonal architecture, and project branding conventions.
## Requirements
### Requirement: Hexagonal Architecture with Port Interfaces
The system SHALL use hexagonal architecture with interface-defined ports to decouple the service layer from infrastructure implementations.

#### Scenario: Service depends on interfaces
- **WHEN** the Service struct is initialized
- **THEN** it accepts interface types (GitOperations, WorkspaceStorage, ConfigProvider)
- **AND** concrete implementations are injected at runtime

#### Scenario: Interface definitions in ports package
- **WHEN** a developer looks for interface contracts
- **THEN** they find all port interfaces in `internal/ports/`
- **AND** each interface is documented with its contract

#### Scenario: Mock implementations for testing
- **WHEN** a test needs to isolate the Service
- **THEN** mock implementations from `internal/mocks/` can be injected
- **AND** error scenarios can be tested without filesystem or git access

#### Scenario: Compile-time interface checks
- **WHEN** an implementation is updated
- **THEN** compile-time assertions verify interface compliance
- **EXAMPLES**: `var _ ports.GitOperations = (*gitx.GitEngine)(nil)`

### Requirement: Centralized Service Initialization
The system SHALL initialize all services through a centralized App struct that manages dependencies and lifecycle. The App struct SHALL support functional options for injecting custom implementations.

**Supported Functional Options:**
- `WithGitOperations(ports.GitOperations)` — Inject custom git engine or mock for testing
- `WithWorkspaceStorage(ports.WorkspaceStorage)` — Inject custom workspace storage or mock for testing
- `WithConfigProvider(ports.ConfigProvider)` — Inject custom config provider or mock for testing
- `WithLogger(*logging.Logger)` — Inject custom logger instance

#### Scenario: App creation with defaults
- **WHEN** `app.New(debug)` is called with valid config and no options
- **THEN** an App struct is returned with default GitEngine, WorkspaceStore, and ConfigProvider
- **AND** all services are ready for use

#### Scenario: App creation with custom dependencies
- **WHEN** `app.New(debug, WithGitOperations(mockGit), WithWorkspaceStorage(mockStore))` is called
- **THEN** an App struct is returned with the provided mock implementations
- **AND** the service uses the injected dependencies

#### Scenario: App creation fails with missing config
- **WHEN** `app.New(debug)` is called and config file does not exist
- **THEN** an error is returned describing the missing config
- **AND** no App instance is created

#### Scenario: Unit test with injected mocks
- **WHEN** a test creates App with `WithGitOperations(mocks.NewGitOps())`
- **THEN** the App uses the mock implementation
- **AND** no real git operations occur

### Requirement: Command Registration Uses App Context
Commands SHALL be registered through builder functions that retrieve dependencies from the App stored in command context.

#### Scenario: Workspace commands registered
- **WHEN** the root command is initialized
- **THEN** workspace command builder functions are called
- **AND** workspace subcommands are attached to the root command
- **AND** each command can access the App via context

#### Scenario: Command execution with dependencies
- **WHEN** a user executes `canopy workspace new PROJ-123`
- **THEN** the command handler retrieves the App from context
- **AND** uses the App service to create the workspace
- **AND** no duplicate service initialization occurs

### Requirement: Testable Command Handlers
Command handlers SHALL support swapping dependencies for tests through the App struct.

#### Scenario: Unit test with mock service
- **WHEN** a test creates an App with mocked services
- **THEN** a command can execute using the mock
- **AND** the test can verify service method calls

#### Scenario: Integration test with real services
- **WHEN** a test creates an App with temporary directories
- **THEN** commands execute against the real filesystem and config
- **AND** the test can verify end-to-end behavior

### Requirement: No Global Service Variables
The system SHALL avoid global service or config variables, requiring commands to obtain dependencies from the App context.

#### Scenario: Command reads config without globals
- **WHEN** a command needs configuration or logger access
- **THEN** it retrieves the App from context
- **AND** uses App.Config and App.Logger instead of any global variables

### Requirement: Project Naming and Branding
The system SHALL be named "Canopy" with the binary named `canopy`, using forest/tree metaphors in all user-facing communication.

#### Scenario: Binary installation and invocation
- **WHEN** a user installs the tool via `go install`
- **THEN** the binary is named `canopy` (not `yard` or `yardmaster`)
- **AND** all commands are invoked as `canopy <command>`

#### Scenario: Configuration directory naming
- **WHEN** the system initializes or loads configuration
- **THEN** configuration is stored in `~/.canopy/` directory
- **AND** config file is `~/.canopy/config.yaml`

#### Scenario: Environment variables
- **WHEN** configuration is loaded from environment
- **THEN** environment variables use `CANOPY_` prefix
- **EXAMPLES**: `CANOPY_PROJECTS_ROOT`, `CANOPY_WORKSPACES_ROOT`

#### Scenario: Documentation uses consistent branding
- **WHEN** users read help text, README, or error messages
- **THEN** the project is referred to as "Canopy"
- **AND** metaphors reference canopy, forest, trees, and branches (not railroad/yard terminology)
- **AND** the metaphor explanation appears in the README introduction

### Requirement: Canopy Metaphor Documentation
The README SHALL include an explanation of the canopy metaphor in the introduction section.

#### Scenario: README metaphor explanation
- **WHEN** a user reads the README introduction
- **THEN** they see an explanation that canopy represents a bird's-eye view above the forest
- **AND** the explanation connects the metaphor to managing git workspaces and branches
- **AND** it clarifies that the TUI provides a literal canopy-level view of all workspaces

### Requirement: Interface-Based Dependencies
Core services SHALL depend on interfaces rather than concrete implementations.

#### Scenario: Git operations via interface
- **GIVEN** the Service depends on GitOperations interface
- **WHEN** tests provide a mock implementation
- **THEN** tests SHALL run without real git operations

#### Scenario: Workspace storage via interface
- **GIVEN** the Service depends on WorkspaceStorage interface
- **WHEN** tests provide a mock implementation
- **THEN** tests SHALL run without filesystem access

### Requirement: Hexagonal Architecture
The codebase SHALL follow hexagonal architecture patterns.

#### Scenario: Port definitions
- **GIVEN** interfaces are defined in `internal/ports/`
- **WHEN** adapters implement these interfaces
- **THEN** the domain layer SHALL remain decoupled from infrastructure

### Requirement: Pure go-git Implementation
The system SHALL use go-git library as the primary implementation for all git operations. An explicit, documented escape hatch (`internal/gitx/git.go:RunCommand`) MAY invoke the git CLI for operations go-git cannot support.

#### Scenario: Clone repository with go-git
- **WHEN** a repository needs to be cloned
- **THEN** the operation is performed using go-git's Clone function
- **AND** no external git process is spawned

#### Scenario: Create worktree with go-git
- **WHEN** a worktree needs to be created
- **THEN** the operation is performed using go-git's Worktree API
- **AND** no external git process is spawned

#### Scenario: Fetch updates with go-git
- **WHEN** repository updates need to be fetched
- **THEN** the operation is performed using go-git's Fetch function
- **AND** no external git process is spawned

#### Scenario: Branch operations with go-git
- **WHEN** branch creation or checkout is needed
- **THEN** the operation is performed using go-git's Branch and Checkout APIs
- **AND** no external git process is spawned

#### Scenario: Escape hatch for unsupported operations
- **WHEN** a git operation cannot be performed with go-git (e.g., worktree creation with detached HEAD)
- **THEN** the `RunCommand` escape hatch MAY be used to invoke the git CLI
- **AND** usage is documented in the code
- **AND** the CLI invocation still returns domain errors wrapped with contextual information

### Requirement: Uniform Error Handling for Git Operations
All git operations SHALL return domain errors wrapped with context, without exposing go-git internals. This requirement extends the Typed Error System to specifically cover git operation failures.

#### Scenario: Git operation returns domain error
- **WHEN** a git operation fails
- **THEN** the error is wrapped as an internal/errors type
- **AND** the original go-git error is available via errors.Unwrap()

#### Scenario: go-git error mapping
- **WHEN** a go-git operation fails
- **THEN** the error is mapped to an appropriate CanopyError code:
  - Authentication failures (SSH key issues, credentials) → `ErrAuthenticationFailed`
  - Network/timeouts (DNS, connection refused, timeouts) → `ErrNetworkFailed`
  - Repository not found (invalid URL, 404) → `ErrRepoNotFound`
  - Permission denied (403, read-only repo) → `ErrPermissionDenied`
  - Protocol errors (git protocol issues) → `ErrGitOperationFailed`
- **AND** the original error is preserved via `errors.Unwrap()`

#### Scenario: CLI escape hatch error handling
- **WHEN** a git CLI command (via RunCommand) fails
- **THEN** the error is wrapped as `ErrCommandFailed` with exit code context
- **AND** stderr output is included in the error context

### Requirement: Typed Error System
The application SHALL use typed errors with error codes for all domain errors.

#### Scenario: Create workspace not found error
- **WHEN** `NewWorkspaceNotFound("my-ws")` is called
- **THEN** returned CanopyError SHALL have Code `WORKSPACE_NOT_FOUND`
- **AND** Message SHALL contain the workspace ID "my-ws"

#### Scenario: Create repo not clean error
- **WHEN** `NewRepoNotClean("/path/to/repo")` is called
- **THEN** returned CanopyError SHALL have Code `REPO_NOT_CLEAN`
- **AND** Context SHALL contain the repo path

### Requirement: Error Wrapping
The application SHALL support wrapping errors to preserve root cause using a general-purpose `Wrap` function.

#### Scenario: Wrap error with operation context
- **WHEN** `errors.Wrap(err, errors.ErrGitOperationFailed, "clone failed")` is called
- **THEN** returned CanopyError SHALL have the specified Code
- **AND** Cause SHALL contain the original error
- **AND** `errors.Unwrap()` SHALL return the original error

**Note**: The wrapping pattern uses a single `Wrap(cause error, code ErrorCode, message string)` function rather than domain-specific wrappers. Callers specify the appropriate error code for their domain.

### Requirement: Error Matching
Errors SHALL support standard Go error matching with `errors.Is()` and `errors.As()`.

#### Scenario: Match error by code
- **WHEN** CanopyError is returned and `errors.Is()` is used
- **THEN** matching SHALL succeed for errors with same ErrorCode

#### Scenario: Extract error details
- **WHEN** `errors.As()` is used with `*CanopyError`
- **THEN** full error details SHALL be accessible including Code, Message, and Context

### Requirement: Error Context
Errors SHALL support contextual key-value pairs for debugging.

#### Scenario: Include context in error
- **WHEN** error is created with Context map
- **THEN** Context SHALL contain all provided key-value pairs
- **AND** Context SHALL be accessible for logging and debugging

### Requirement: Single Responsibility Service Components
The workspaces service layer SHALL be composed of focused sub-services, each with a single responsibility.

#### Scenario: RepoResolver handles identifier resolution
- **WHEN** a repo identifier is provided (name, alias, or URL)
- **THEN** the RepoResolver component resolves it to a canonical repo path
- **AND** the resolution logic is isolated from workspace operations

#### Scenario: DiskUsageCalculator handles size computation
- **WHEN** workspace disk usage is requested
- **THEN** the DiskUsageCalculator component computes and caches the result
- **AND** caching logic is isolated from workspace operations

#### Scenario: CanonicalRepoService handles repo management
- **WHEN** canonical repo operations are performed (list, add, remove, sync)
- **THEN** the CanonicalRepoService component handles the operation
- **AND** repo management is isolated from workspace lifecycle

### Requirement: Context Propagation in Service Layer
All public Service methods SHALL accept `context.Context` as their first parameter to enable cancellation, timeout, and observability.

#### Scenario: Service method accepts context
- **WHEN** a CLI command calls a Service method
- **THEN** the command SHALL pass `cmd.Context()` as the first argument
- **AND** the Service SHALL propagate context to downstream operations

#### Scenario: Context cancellation stops operation
- **WHEN** context is cancelled during a git operation
- **THEN** the operation SHALL return an error
- **AND** any spawned goroutines SHALL terminate

#### Scenario: Context timeout for network operations
- **WHEN** a git network operation exceeds the context deadline
- **THEN** the operation SHALL return a context deadline exceeded error
- **AND** partial operations SHALL be cleaned up where possible

### Requirement: Git Operations Context Support
Git network operations (Clone, Fetch, Push, Pull) SHALL respect context cancellation and deadlines.

#### Scenario: Clone respects context timeout
- **WHEN** `GitOperations.Clone(ctx, url, name)` is called with a deadline
- **AND** the clone exceeds the deadline
- **THEN** the clone SHALL be aborted
- **AND** a context deadline exceeded error SHALL be returned

#### Scenario: Fetch respects context cancellation
- **WHEN** `GitOperations.Fetch(ctx, name)` is called
- **AND** the context is cancelled during fetch
- **THEN** the fetch SHALL stop
- **AND** a context cancelled error SHALL be returned

#### Scenario: Parallel operations cancel on context done
- **WHEN** `RunGitInWorkspace` executes in parallel mode
- **AND** the context is cancelled
- **THEN** all pending goroutines SHALL be signalled to stop
- **AND** the function SHALL return promptly

### Requirement: Default Network Timeout
Network-bound git operations SHALL use a default timeout when no deadline is set on the context.

#### Scenario: Default timeout applied
- **GIVEN** a context with no deadline
- **WHEN** a git network operation is initiated
- **THEN** a default timeout of 5 minutes SHALL be applied
- **AND** operations exceeding this timeout SHALL fail

#### Scenario: Explicit deadline overrides default
- **GIVEN** a context with an explicit deadline
- **WHEN** a git network operation is initiated
- **THEN** the explicit deadline SHALL be used
- **AND** the default timeout SHALL NOT apply

### Requirement: Automatic Retry for Git Network Operations
Git network operations SHALL automatically retry on transient failures using exponential backoff.

#### Scenario: Transient failure triggers retry
- **GIVEN** a git clone operation
- **WHEN** the operation fails with a network timeout
- **THEN** the operation SHALL be retried
- **AND** subsequent attempts SHALL use exponential backoff

#### Scenario: Permanent failure does not retry
- **GIVEN** a git clone operation
- **WHEN** the operation fails with authentication error (401/403)
- **THEN** the operation SHALL NOT be retried
- **AND** the error SHALL be returned immediately

#### Scenario: Max attempts exceeded
- **GIVEN** retry configuration with max_attempts=3
- **WHEN** all 3 attempts fail with transient errors
- **THEN** the final error SHALL be returned
- **AND** the error message SHALL indicate retry exhaustion

### Requirement: Exponential Backoff with Jitter
Retry delays SHALL use exponential backoff with random jitter to prevent thundering herd.

#### Scenario: Backoff calculation
- **GIVEN** initial_delay=1s and multiplier=2
- **WHEN** calculating delay for attempt N
- **THEN** base delay SHALL be initial_delay * (multiplier ^ (N-1))
- **AND** jitter SHALL be applied (±25% of base delay)
- **AND** delay SHALL not exceed max_delay

#### Scenario: Jitter prevents synchronized retries
- **GIVEN** multiple concurrent operations failing
- **WHEN** retries are scheduled
- **THEN** retry times SHALL be randomized
- **AND** NOT synchronized to the same instant

### Requirement: Retry Logging
Retry attempts SHALL be logged for debugging and observability.

#### Scenario: Log retry attempt
- **WHEN** a retry is attempted
- **THEN** an Info-level log message SHALL be emitted
- **AND** the log SHALL include attempt number, operation, and delay

#### Scenario: Log final failure
- **WHEN** all retry attempts are exhausted
- **THEN** a Warning-level log message SHALL be emitted
- **AND** the log SHALL include total attempts and final error

### Requirement: Context-Aware Retry
Retry operations SHALL respect context cancellation and deadlines.

#### Scenario: Context cancelled during backoff
- **GIVEN** a retry operation waiting for backoff delay
- **WHEN** the context is cancelled
- **THEN** the retry SHALL be aborted immediately
- **AND** a context cancelled error SHALL be returned

#### Scenario: Context deadline during retry
- **GIVEN** a context with 5-second deadline
- **AND** retry backoff would exceed deadline
- **THEN** retry SHALL be skipped
- **AND** the last error SHALL be returned

### Requirement: Direct Workspace Lookup
The workspace storage SHALL support direct lookup by workspace ID without listing all workspaces. The method signature SHALL be `LoadByID(id string) (*domain.Workspace, string, error)` returning the workspace metadata, directory name, and any error.

#### Scenario: Direct lookup by ID
- **WHEN** `LoadByID(id)` is called with a valid workspace ID
- **THEN** the storage SHALL attempt direct path access
- **THEN** the method SHALL return `(workspace, dirName, nil)` where `workspace` is the metadata and `dirName` is the directory name

#### Scenario: Direct lookup fallback
- **WHEN** direct path access fails because the ID differs from the directory name
- **THEN** the storage SHALL fall back to scanning all workspaces
- **THEN** the method SHALL return `(workspace, dirName, nil)` if the workspace exists

#### Scenario: Workspace not found
- **WHEN** `LoadByID(id)` is called with a non-existent workspace ID
- **THEN** the method SHALL return `(nil, "", WorkspaceNotFound)` error

### Requirement: Workspace Metadata Caching
The service layer SHALL cache workspace metadata to reduce filesystem I/O.

#### Scenario: Cache hit
- **WHEN** looking up a workspace that was recently accessed and the cache entry has not expired
- **THEN** the cached workspace SHALL be returned
- **THEN** no filesystem I/O SHALL occur

#### Scenario: Cache miss
- **WHEN** looking up a workspace not in cache
- **THEN** the workspace SHALL be loaded from storage
- **THEN** the result SHALL be added to the cache

#### Scenario: Cache invalidation on write
- **WHEN** a workspace is created, updated, or deleted
- **THEN** the cache entry for that workspace SHALL be invalidated
- **THEN** subsequent lookups SHALL reload from storage

### Requirement: Extensible Repository Resolution
The system SHALL use a Strategy pattern for repository resolution to enable extensibility.

#### Scenario: URL resolution strategy
- **WHEN** a repository identifier starts with a URL scheme (http://, https://, git@, ssh://, git://, file://)
- **THEN** the URL strategy SHALL handle resolution
- **THEN** the repository name SHALL be derived from the URL path

#### Scenario: Registry resolution strategy
- **WHEN** a repository identifier matches a registered alias
- **THEN** the registry strategy SHALL return the registered URL
- **THEN** the alias SHALL be used as the repository name

#### Scenario: GitHub shorthand resolution strategy
- **WHEN** a repository identifier contains exactly one slash (owner/repo format) with neither segment empty
- **THEN** the GitHub shorthand strategy SHALL construct a GitHub HTTPS URL
- **THEN** the repo segment SHALL be used as the repository name

#### Scenario: Strategy chain execution
- **WHEN** resolving a repository identifier
- **THEN** strategies SHALL be tried in default order: URL → Registry → GitHub shorthand
- **THEN** the first strategy that returns a successful match SHALL be used (first-match wins)
- **THEN** if a strategy matches but encounters an error during resolution, the chain SHALL abort with that error
- **THEN** if no strategy matches the input format, an `UnknownRepository` error SHALL be returned

#### Scenario: Strategy precedence override
- **WHEN** the resolver is configured with a custom strategy order
- **THEN** the custom order SHALL override the default precedence
- **THEN** strategies not in the custom list SHALL be excluded from resolution

### Requirement: Shared Git URL Utilities
The system SHALL provide a shared package for Git URL parsing with the following operations:
- **Scheme detection**: Determine if a string is a valid Git URL
- **Repository name extraction**: Extract the repo name from a URL
- **Alias derivation**: Generate a default alias from a URL

#### Scenario: URL scheme detection
- **WHEN** checking if a string is a Git URL
- **THEN** the utility SHALL recognize: http://, https://, ssh://, git://, git@, file://
- **THEN** the utility SHALL return false for plain strings without URL schemes

#### Scenario: Repository name extraction
- **WHEN** extracting a repository name from a URL
- **THEN** the utility SHALL handle SCP-style URLs (`git@host:owner/repo.git`)
- **THEN** the utility SHALL handle standard URLs (`https://host/owner/repo.git`)
- **THEN** the utility SHALL strip `.git` suffix if present
- **THEN** the utility SHALL return the last non-empty path segment
- **THEN** the utility SHALL return empty string for invalid or empty input

#### Scenario: Alias derivation from valid URL
- **WHEN** deriving an alias from a valid Git URL
- **THEN** the utility SHALL extract the repository name
- **THEN** the utility SHALL convert to lowercase
- **THEN** the utility SHALL return a non-empty string suitable for use as a registry alias

#### Scenario: Alias derivation from invalid URL
- **WHEN** deriving an alias from an invalid or empty URL
- **THEN** the utility SHALL return an empty string

