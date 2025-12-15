## MODIFIED Requirements

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

## ADDED Requirements

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

