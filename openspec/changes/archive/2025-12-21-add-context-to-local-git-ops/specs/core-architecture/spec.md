## MODIFIED Requirements

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

## ADDED Requirements

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

