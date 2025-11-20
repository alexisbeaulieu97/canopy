# Git Operations Specification Deltas

## ADDED Requirements

### Requirement: Automated Git Hook Installation
The system SHALL automatically install configured git hooks to all worktrees when creating workspaces.

#### Scenario: Install hooks on workspace creation
- **WHEN** user creates workspace with repos and hooks are configured
- **THEN** each configured hook is installed to each repo's .git/hooks/ directory
- **AND** hook files are made executable
- **AND** hooks execute when triggered by git operations

#### Scenario: Hook configuration from file
- **WHEN** config defines hook with file path reference
- **THEN** system reads hook script from specified file
- **AND** installs content to worktree .git/hooks/
- **AND** returns error if file doesn't exist

#### Scenario: Inline hook script
- **WHEN** config defines hook with inline script
- **THEN** system writes script directly to .git/hooks/
- **AND** hook executes as defined

### Requirement: Hook Synchronization
The system SHALL provide command to update hooks in existing workspaces when configuration changes.

#### Scenario: Sync hooks to existing workspace
- **WHEN** user runs `yard workspace sync-hooks PROJ-123`
- **THEN** all configured hooks are reinstalled to workspace repos
- **AND** existing hooks are overwritten
- **AND** summary shows which hooks were updated per repo

#### Scenario: Dry-run hook sync
- **WHEN** user runs `yard workspace sync-hooks PROJ-123 --dry-run`
- **THEN** system shows what hooks would be installed
- **AND** no actual changes are made
- **AND** user can review before applying

### Requirement: Hook Configuration Format
Git hooks SHALL be defined in config.yaml with hook name and script content or file path.

#### Scenario: Multiple hooks configured
- **WHEN** config.yaml defines pre-commit and pre-push hooks
- **THEN** both hooks are parsed successfully
- **AND** both are installed to worktrees

#### Scenario: Invalid hook name
- **WHEN** config defines hook with non-standard name
- **THEN** validation error is returned during config load
- **AND** lists valid hook names
