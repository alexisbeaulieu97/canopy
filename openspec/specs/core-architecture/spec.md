# core-architecture Specification

## Purpose
Defines the core **architectural patterns and structural organization** for Canopy including service initialization, dependency injection via App context, command registration, hexagonal architecture, interface-based dependencies, error handling patterns, retry strategies, and project branding conventions. This specification governs **how** the system is structured and organized, independent of specific business rules. For domain rules and behaviors, see `core`.
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

#### Scenario: Adapter package naming
- **GIVEN** the hexagonal architecture separates ports from adapters
- **WHEN** a developer looks for the WorkspaceStorage implementation
- **THEN** they find it in `internal/storage/` package
- **AND** the package name reflects its role as a storage adapter

### Requirement: Pure go-git Implementation
The system SHALL use go-git library as the primary implementation for all git operations. An explicit, documented escape hatch (`internal/gitx/git.go:RunCommand`) MAY invoke the git CLI for operations go-git cannot support, including worktree management.

#### Scenario: Clone repository with go-git
- **WHEN** a repository needs to be cloned
- **THEN** the operation is performed using go-git's Clone function
- **AND** no external git process is spawned

#### Scenario: Create worktree with git CLI
- **WHEN** a worktree needs to be created for a workspace
- **THEN** the operation is performed using `git worktree add` via RunCommand
- **AND** the worktree shares objects with the canonical repository
- **AND** the worktree's origin remote points to the upstream URL (not the canonical path)

#### Scenario: Remove worktree with git CLI
- **WHEN** a worktree needs to be removed during workspace close
- **THEN** the operation is performed using `git worktree remove` via RunCommand
- **AND** the worktree reference is cleaned from the canonical repository

#### Scenario: Fetch updates with go-git
- **WHEN** repository updates need to be fetched
- **THEN** the operation is performed using go-git's Fetch function
- **AND** no external git process is spawned

#### Scenario: Branch operations with go-git
- **WHEN** branch creation or checkout is needed
- **THEN** the operation is performed using go-git's Branch and Checkout APIs
- **AND** no external git process is spawned

#### Scenario: Escape hatch for unsupported operations
- **WHEN** a git operation cannot be performed with go-git (e.g., worktree management)
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
All errors returned by internal packages SHALL use the typed error system defined in `internal/errors`. Raw `fmt.Errorf` calls MUST NOT be used in production code paths.

The error system provides:
- `CanopyError` struct with `Code`, `Message`, `Cause`, and `Context` fields
- Constructor functions for each error type (e.g., `NewWorkspaceNotFound`, `NewIOFailed`)
- Sentinel errors for use with `errors.Is()` matching
- `Wrap()` for adding context while preserving the underlying error

#### Scenario: All internal packages use typed errors
- **WHEN** any error is returned from internal code, **THEN** it SHALL be a `*CanopyError` or wrap one
- **WHEN** checking error type, **THEN** it SHALL be matchable with `errors.Is(err, cerrors.SomeError)`

#### Scenario: Config validation returns typed errors
- **WHEN** `ValidateValues()` is called on a config with an invalid value, **THEN** the error SHALL be a `ConfigValidation` error
- **WHEN** a config validation error occurs, **THEN** the error context SHALL include the field name and reason

#### Scenario: Path validation returns typed errors
- **WHEN** `ValidateEnvironment()` is called on a path that doesn't exist or isn't a directory, **THEN** the error SHALL be a `PathInvalid` or `PathNotDirectory` error
- **WHEN** a path validation error occurs, **THEN** the error context SHALL include the path

#### Scenario: Workspace storage returns typed errors
- **WHEN** an I/O error occurs during workspace storage operations (read, write, delete), **THEN** the error SHALL be wrapped with `NewIOFailed` or `NewWorkspaceMetadataError`
- **WHEN** wrapping storage errors, **THEN** the underlying cause SHALL be preserved via `Unwrap()`

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
Service components SHALL follow the Single Responsibility Principle, with each service focused on a cohesive set of operations.

The workspace service layer SHALL be organized as follows:
- `Service` - Coordinator for workspace lifecycle (create, close, restore, list, status)
- `WorkspaceGitService` - Git command execution across workspace repos
- `WorkspaceOrphanService` - Orphan worktree detection and remediation
- `WorkspaceExportService` - Workspace export/import functionality
- `CanonicalRepoService` - Canonical repository management (existing)
- `RepoResolver` - Repository identifier resolution (existing)
- `DiskUsageCalculator` - Disk usage calculation (existing)
- `WorkspaceCache` - Workspace lookup caching (existing)

#### Scenario: Main service coordinates sub-services
- **GIVEN** a workspace operation that spans multiple concerns
- **WHEN** the operation is invoked on the main Service
- **THEN** the Service SHALL delegate to appropriate sub-services
- **AND** the public API SHALL remain unchanged

#### Scenario: Sub-services are independently testable
- **GIVEN** a sub-service like WorkspaceGitService
- **WHEN** unit tests are written
- **THEN** the sub-service SHALL be testable without instantiating the full Service
- **AND** dependencies SHALL be injectable via interfaces

#### Scenario: Service file size is manageable
- **GIVEN** the service layer structure
- **WHEN** any single service file is examined
- **THEN** it SHALL contain fewer than 500 lines of code
- **AND** it SHALL have a clear, single purpose

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
All git operations SHALL respect context cancellation and deadlines.

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

#### Scenario: CreateWorktree respects context cancellation
- **WHEN** `GitOperations.CreateWorktree(ctx, repoName, worktreePath, branchName)` is called
- **AND** the context is cancelled during worktree creation
- **THEN** the operation SHALL be aborted
- **AND** a context cancelled error SHALL be returned

#### Scenario: Status respects context cancellation
- **WHEN** `GitOperations.Status(ctx, path)` is called
- **AND** the context is cancelled during status check
- **THEN** the operation SHALL be aborted
- **AND** a context cancelled error SHALL be returned

#### Scenario: Checkout respects context cancellation
- **WHEN** `GitOperations.Checkout(ctx, path, branchName, create)` is called
- **AND** the context is cancelled during checkout
- **THEN** the operation SHALL be aborted
- **AND** a context cancelled error SHALL be returned

#### Scenario: List respects context cancellation
- **WHEN** `GitOperations.List(ctx)` is called
- **AND** the context is cancelled during listing
- **THEN** the operation SHALL be aborted
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

### Requirement: Default Local Operation Timeout
Local git operations (CreateWorktree, Status, Checkout, List) SHALL use a default timeout when no deadline is set on the context.

#### Scenario: Default local timeout applied
- **GIVEN** a context with no deadline
- **WHEN** a local git operation is initiated
- **THEN** a default timeout of 30 seconds SHALL be applied
- **AND** operations exceeding this timeout SHALL fail with a timeout error

#### Scenario: Explicit deadline overrides default for local ops
- **GIVEN** a context with an explicit deadline
- **WHEN** a local git operation is initiated
- **THEN** the explicit deadline SHALL be used
- **AND** the default timeout SHALL NOT apply

### Requirement: Integration Test Coverage

The project SHALL maintain integration tests covering all major user workflows.

Integration tests SHALL exercise the complete stack from CLI through service layer to filesystem operations, using real git repositories in isolated temporary directories.

#### Scenario: Workspace lifecycle coverage

- **WHEN** running the integration test suite
- **THEN** tests SHALL cover workspace create, list, view, close, restore, and rename operations

#### Scenario: Repository management coverage

- **WHEN** running the integration test suite
- **THEN** tests SHALL cover adding repos to workspaces, removing repos, and status reporting

#### Scenario: Branch operation coverage

- **WHEN** running the integration test suite
- **THEN** tests SHALL cover branch switching and creation across workspace repositories

#### Scenario: Error handling coverage

- **WHEN** running the integration test suite
- **THEN** tests SHALL verify appropriate error messages for common failure scenarios including dirty repos, missing workspaces, and invalid configuration

### Requirement: Parallel Repository Operations
Repository operations during workspace creation SHALL execute in parallel with bounded concurrency.

#### Scenario: Parallel EnsureCanonical execution
- **WHEN** creating a workspace with multiple repositories, **THEN** EnsureCanonical operations SHALL execute in parallel, the number of concurrent operations SHALL be limited by `parallel_workers` config, and worktree creation SHALL wait for the corresponding EnsureCanonical to complete

#### Scenario: Configurable worker count
- **WHEN** creating a workspace with 10 repositories and config has `parallel_workers: 6`, **THEN** at most 6 EnsureCanonical operations SHALL run concurrently

#### Scenario: Default worker count
- **WHEN** creating a workspace with multiple repositories and `parallel_workers` is not configured, **THEN** the default of 4 concurrent operations SHALL be used

#### Scenario: Worker count validation
- **WHEN** `parallel_workers` is configured with an invalid value (0, negative, or exceeding maximum), **THEN** the configuration SHALL fail validation with a clear error message

#### Scenario: Error handling with fail-fast (default)
- **WHEN** one EnsureCanonical operation fails during workspace creation and `continue_on_error` is false (default), **THEN** remaining operations SHALL be cancelled, successfully cloned repositories SHALL be cleaned up, and the error message SHALL indicate which repository failed

#### Scenario: Error handling with continue-on-error
- **WHEN** one EnsureCanonical operation fails during workspace creation and `continue_on_error: true`, **THEN** remaining operations SHALL continue, partial results SHALL be available, and errors SHALL be aggregated and reported

#### Scenario: Context cancellation propagates to workers
- **WHEN** workspace creation context is cancelled, **THEN** all parallel operations SHALL receive cancellation and the operation SHALL return promptly with a cancellation error

### Requirement: Named Timeout Constants

Timeout values SHALL be defined as named constants with documentation explaining their purpose and rationale.

Named constants for timeouts (e.g., `gitx.DefaultLocalTimeout`) SHALL include a documentation comment explaining their purpose.

#### Scenario: Cleanup operation timeout

- **WHEN** a cleanup operation requires a timeout context, **THEN** the code SHALL use a named constant (e.g., `gitx.DefaultLocalTimeout`) with a documentation comment

#### Scenario: No magic timeout numbers

- **WHEN** reviewing CLI command handlers, **THEN** no inline magic numbers for timeouts SHALL be present and all timeouts SHALL reference named constants

### Requirement: Transactional Operation Integrity
Mutating operations SHALL maintain system consistency by cleaning up partial changes on failure.

#### Scenario: Create workspace fails during clone
- **WHEN** `CreateWorkspace` successfully creates the workspace directory
- **AND** clone operation fails for one repo
- **THEN** the system SHALL remove the workspace directory
- **AND** no metadata file SHALL remain
- **AND** the operation SHALL return an error

#### Scenario: Add repo fails during metadata update
- **WHEN** `AddRepoToWorkspace` successfully creates the worktree
- **AND** metadata update fails
- **THEN** the system SHALL remove the created worktree
- **AND** the workspace metadata SHALL remain unchanged
- **AND** the operation SHALL return an error

#### Scenario: Restore workspace fails during recreation
- **WHEN** `RestoreWorkspace` begins restoration
- **AND** worktree creation fails
- **THEN** the closed workspace entry SHALL remain intact
- **AND** no partial workspace directory SHALL remain
- **AND** the user can retry the restore operation

#### Scenario: Rollback actions logged
- **WHEN** an operation fails and rollback is triggered
- **THEN** the system SHALL log each rollback action at debug level
- **AND** include the original error and cleanup status

### Requirement: Workspace Locking
The system SHALL prevent concurrent mutating operations on the same workspace using file-based locks.

#### Scenario: Lock acquired for create operation
- **WHEN** a workspace creation is initiated
- **THEN** the system SHALL acquire an exclusive lock for the workspace ID
- **AND** release the lock when the operation completes (success or failure)

#### Scenario: Concurrent operations blocked
- **GIVEN** workspace `PROJ-1` has an active lock held by another process
- **WHEN** a second process attempts to close `PROJ-1`
- **THEN** the second process SHALL wait up to `lock_timeout` for the lock
- **AND** fail with `ErrWorkspaceLocked` if timeout expires

#### Scenario: Read operations not locked
- **GIVEN** workspace `PROJ-1` has an active lock
- **WHEN** I run `canopy workspace list` or `canopy workspace status PROJ-1`
- **THEN** the operation SHALL complete without waiting for the lock

#### Scenario: Stale lock cleanup
- **GIVEN** a lock file exists that is older than `lock_stale_threshold`
- **WHEN** another operation attempts to acquire the lock
- **THEN** the system SHALL remove the stale lock
- **AND** SHALL acquire a fresh lock

#### Scenario: Lock released on failure
- **GIVEN** an operation holds a lock on workspace `PROJ-1`
- **WHEN** the operation fails with an error
- **THEN** the lock SHALL be released
- **AND** subsequent operations SHALL proceed without waiting

### Requirement: Parallel Git Operations Early Termination
When running git commands in parallel with `continueOnError=false`, the system SHALL cancel remaining operations after the first failure.

#### Scenario: First error cancels pending operations
- **GIVEN** a workspace with 5 repositories
- **AND** parallel git execution is enabled
- **AND** `continueOnError` is false
- **WHEN** the first repository operation fails
- **THEN** pending operations SHALL be cancelled
- **AND** running operations SHALL be signalled to stop
- **AND** the function SHALL return the first error

#### Scenario: All operations complete when continueOnError is true
- **GIVEN** a workspace with 5 repositories
- **AND** parallel git execution is enabled
- **AND** `continueOnError` is true
- **WHEN** some repository operations fail
- **THEN** all operations SHALL complete
- **AND** all results (success and failure) SHALL be returned

#### Scenario: No race conditions in result collection
- **GIVEN** parallel git execution
- **WHEN** multiple goroutines write results concurrently
- **THEN** all results SHALL be collected without data races
- **AND** `go test -race` SHALL pass

### Requirement: Complete Port Interface Coverage
Every injectable service dependency SHALL have a corresponding interface in `internal/ports/`.

#### Scenario: New dependency requires interface
- **GIVEN** a new dependency is added to the Service
- **WHEN** the dependency is injectable
- **THEN** an interface SHALL be created in `internal/ports/`
- **AND** a mock implementation SHALL be created in `internal/mocks/`

### Requirement: Registry Transaction Safety

Registry modifications SHALL use atomic save-with-rollback semantics to prevent partial state on failure.

The save-with-rollback pattern SHALL:
- Attempt to persist registry changes
- On failure, execute the provided rollback function
- Log rollback failures without masking the original error
- Return the original save error to the caller

#### Scenario: Successful save

- **WHEN** a registry modification is saved successfully
- **THEN** no rollback SHALL be attempted
- **AND** the function SHALL return nil

#### Scenario: Save failure with successful rollback

- **WHEN** a registry save fails
- **THEN** the rollback function SHALL be executed
- **AND** the original save error SHALL be returned

#### Scenario: Save failure with rollback failure

- **WHEN** a registry save fails and rollback also fails
- **THEN** the rollback failure SHALL be logged
- **AND** the original save error SHALL be returned (not masked)

### Requirement: Shared Test Utilities Package
The codebase SHALL provide a shared test utilities package to avoid duplication of test helper functions.

#### Scenario: Git test helpers available
- **GIVEN** a test needs to create a git repository
- **WHEN** the test imports `internal/testutil`
- **THEN** `testutil.CreateRepoWithCommit(t, path)` SHALL be available
- **AND** the helper SHALL create a valid git repo with initial commit

#### Scenario: Filesystem test helpers available
- **GIVEN** a test needs to create temporary files
- **WHEN** the test imports `internal/testutil`
- **THEN** `testutil.MustMkdir(t, path)` SHALL be available
- **AND** `testutil.MustWriteFile(t, path, content)` SHALL be available

#### Scenario: Service test setup available
- **GIVEN** a test needs a fully configured test service
- **WHEN** the test calls `testutil.NewTestService(t)`
- **THEN** a struct with initialized dependencies SHALL be returned
- **AND** temporary directories SHALL be created
- **AND** cleanup SHALL be registered with t.Cleanup()

### Requirement: Test Helper Consistency
All test helper functions SHALL follow consistent patterns for error handling and cleanup.

#### Scenario: Helper fails test on error
- **GIVEN** a test helper function with `t *testing.T` parameter
- **WHEN** an error occurs during helper execution
- **THEN** the helper SHALL call `t.Fatalf()` with descriptive message
- **AND** the test SHALL stop execution

#### Scenario: Helper registers cleanup
- **GIVEN** a test helper that creates resources
- **WHEN** the helper completes successfully
- **THEN** cleanup functions SHALL be registered via `t.Cleanup()`
- **AND** resources SHALL be cleaned up after test completion

### Requirement: Cross-Platform Path Construction
All file path construction SHALL use `filepath.Join` or equivalent standard library functions for cross-platform compatibility.

#### Scenario: Path construction uses filepath.Join
- **WHEN** constructing a file path from multiple components
- **THEN** the code SHALL use `filepath.Join` or `filepath.Clean`
- **AND** SHALL NOT use `fmt.Sprintf` with hardcoded path separators

#### Scenario: Worktree path construction
- **WHEN** constructing a worktree path from workspace root, directory name, and repo name
- **THEN** the path SHALL be constructed as `filepath.Join(workspacesRoot, dirName, repoName)`
- **AND** the result SHALL be valid on all supported platforms (Linux, macOS, Windows)

#### Scenario: Environment variable path construction
- **WHEN** constructing paths for hook environment variables (e.g., CANOPY_REPO_PATH)
- **THEN** the path SHALL use platform-appropriate separators
- **AND** SHALL be usable by shell scripts on that platform

### Requirement: Service Delegation Pattern
The main `Service` struct SHALL maintain backward compatibility by delegating to sub-services for extracted functionality.

#### Scenario: Existing method calls work unchanged
- **GIVEN** code that calls `service.PushWorkspace()`
- **WHEN** the method is invoked after refactoring
- **THEN** it SHALL delegate to `WorkspaceGitService.Push()`
- **AND** the behavior SHALL be identical to before refactoring

### Requirement: Implementation-Agnostic Storage Interface
The `WorkspaceStorage` interface SHALL be implementation-agnostic, using domain identifiers (workspace IDs) rather than implementation details (directory names, file paths).

#### Scenario: Create workspace by domain object
- **WHEN** `Create(ctx, workspace)` is called with a domain.Workspace
- **THEN** the storage SHALL persist the workspace
- **AND** the caller SHALL NOT need to specify directory names or paths

#### Scenario: Load workspace by ID
- **WHEN** `Load(ctx, id)` is called with a workspace ID
- **THEN** the storage SHALL return the workspace metadata
- **AND** the caller SHALL NOT need to know the underlying storage path

#### Scenario: Save workspace by domain object
- **WHEN** `Save(ctx, workspace)` is called with a domain.Workspace
- **THEN** the storage SHALL update the persisted workspace using the ID from the domain object
- **AND** the caller SHALL NOT need to provide directory names

#### Scenario: Close workspace by ID
- **WHEN** `Close(ctx, id, closedAt)` is called with a workspace ID
- **THEN** the storage SHALL archive the workspace
- **AND** the caller SHALL NOT need to provide directory names

#### Scenario: Rename workspace by IDs
- **WHEN** `Rename(ctx, oldID, newID)` is called
- **THEN** the storage SHALL update the workspace ID
- **AND** the implementation MAY rename underlying directories as needed

### Requirement: Context Support in Storage Interface
All `WorkspaceStorage` methods SHALL accept `context.Context` as their first parameter to enable cancellation and timeout for I/O operations.

#### Scenario: Storage method accepts context
- **WHEN** a service method calls a storage method
- **THEN** the service SHALL pass its context to the storage method
- **AND** the storage SHALL respect context cancellation

#### Scenario: Context cancellation stops I/O
- **WHEN** context is cancelled during a storage operation
- **THEN** the operation SHALL return promptly
- **AND** an appropriate error SHALL be returned

### Requirement: Configuration Validation Error Type
Config validation errors SHALL use the `ErrConfigValidation` error code for semantic validation failures.

#### Scenario: Invalid field value
- **WHEN** a config field has an invalid value, **THEN** `NewConfigValidation(field, detail)` SHALL be returned
- **WHEN** returning config validation errors, **THEN** the error message SHALL be user-friendly

### Requirement: Path Error Types
Path-related errors SHALL use specific error codes for different failure modes.

#### Scenario: Path does not exist
- **WHEN** a required path does not exist, **THEN** `NewPathInvalid(path, "does not exist")` SHALL be returned

#### Scenario: Path is not a directory
- **WHEN** a path exists but is not a directory, **THEN** `NewPathNotDirectory(path)` SHALL be returned

### Requirement: Standard Library Concurrency Patterns
Concurrent operations SHALL use standard Go concurrency patterns from `golang.org/x/sync` where applicable, rather than custom implementations.

#### Scenario: Parallel repo operations use errgroup
- **WHEN** multiple repositories need concurrent operations (e.g., EnsureCanonical)
- **THEN** the implementation SHALL use `errgroup.Group` for coordination
- **AND** bounded concurrency SHALL be configured via `SetLimit()`

#### Scenario: Fail-fast on first error
- **WHEN** a concurrent operation fails
- **THEN** remaining operations SHALL be cancelled via errgroup's context
- **AND** the first error SHALL be returned to the caller

### Requirement: Worktree Object Sharing
Workspace repositories SHALL be created as git worktrees that share objects with the canonical repository.

#### Scenario: Worktree shares objects
- **WHEN** a worktree is created for a workspace
- **THEN** the worktree's `.git` file points to the canonical repository's worktrees directory
- **AND** no duplicate git objects are stored
- **AND** disk usage scales with working tree size, not repository history

#### Scenario: Worktree remote configuration
- **WHEN** a worktree is created
- **THEN** the worktree's `origin` remote SHALL point to the upstream URL
- **AND** push operations SHALL send commits to the upstream repository
- **AND** pull operations SHALL fetch from the upstream repository

### Requirement: Upstream URL Preservation
The canonical repository SHALL store the original upstream URL for worktree configuration.

#### Scenario: Store upstream URL during clone
- **WHEN** a canonical repository is cloned via EnsureCanonical
- **THEN** the upstream URL SHALL be stored in the repository's git config
- **AND** the URL SHALL be retrievable for worktree remote configuration

#### Scenario: Retrieve upstream URL for worktree
- **WHEN** creating a worktree
- **THEN** the system SHALL retrieve the upstream URL from the canonical repository config
- **AND** configure the worktree's origin remote with this URL

### Requirement: No Deprecated Code Aliases
The codebase SHALL NOT contain the following deprecated type aliases or wrapper functions. All code SHALL use canonical implementations directly.

Deprecated items (to be removed):
- `workspace.ClosedWorkspace` (use `domain.ClosedWorkspace`)
- `hooks.HookContext` (use `domain.HookContext`)
- `resolver.isLikelyURL` (use `giturl.IsURL`)
- `resolver.repoNameFromURL` (use `giturl.ExtractRepoName`)
- `service.CalculateDiskUsage` (use `DiskUsageCalculator.Calculate`)

#### Scenario: Domain types used directly
- **WHEN** code needs to reference domain types like Workspace, ClosedWorkspace, or HookContext, **THEN** the code SHALL import and use `domain.TypeName` directly

#### Scenario: Utility functions used directly
- **WHEN** code needs URL parsing utilities, **THEN** the code SHALL import and use `giturl.FunctionName` directly

### Requirement: CLI Layer Responsibilities
The CLI layer (cmd/canopy/) SHALL focus exclusively on user interface concerns:
- Parsing command-line flags and arguments
- Validating user input format (not business rules)
- Calling service layer methods
- Formatting output for display (text, JSON, table)
- Handling user prompts and confirmations

Business logic, orchestration, and domain operations SHALL remain in the service layer (internal/workspaces/).

#### Scenario: Clear separation of concerns
- **GIVEN** a workspace command is executed
- **WHEN** the command processes
- **THEN** flag parsing SHALL occur in CLI layer
- **AND** business validation SHALL occur in service layer
- **AND** domain operations SHALL occur in service layer
- **AND** output formatting SHALL occur in CLI layer

#### Scenario: Subcommand file organization
- **GIVEN** the CLI codebase
- **WHEN** reviewing file structure
- **THEN** each subcommand SHALL have its own file
- **AND** shared output helpers SHALL be in `presenters.go`
- **AND** parent command SHALL be in `workspace.go`

