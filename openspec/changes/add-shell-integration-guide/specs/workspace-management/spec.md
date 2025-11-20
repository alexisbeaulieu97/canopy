# User Documentation Specification Deltas

## ADDED Requirements

### Requirement: Shell Integration Command
The system SHALL provide a `shell-init` command that outputs shell-specific function code for easy installation.

#### Scenario: Generate bash shell functions
- **WHEN** user runs `yard shell-init --shell bash`
- **THEN** system outputs bash function definitions for yw and yr
- **AND** output includes installation instructions
- **AND** functions use yard CLI internally

#### Scenario: Auto-detect shell
- **WHEN** user runs `yard shell-init` without --shell flag
- **THEN** system detects shell from $SHELL environment variable
- **AND** outputs appropriate shell-specific functions
- **AND** returns error if shell cannot be detected

#### Scenario: Install shell functions
- **WHEN** user runs `eval "$(yard shell-init)" ` in their shell
- **THEN** yw and yr functions become available immediately
- **AND** functions work correctly for current session

### Requirement: Workspace Navigation Function
Shell functions SHALL enable one-command navigation to workspaces using short commands.

#### Scenario: Navigate to workspace with yw
- **WHEN** user runs `yw PROJ-123` in shell with functions loaded
- **THEN** current directory changes to workspace root
- **AND** prompt reflects new location
- **AND** workspace repos are accessible

#### Scenario: Error on nonexistent workspace
- **WHEN** user runs `yw INVALID-ID` for workspace that doesn't exist
- **THEN** error message is displayed
- **AND** current directory remains unchanged

### Requirement: Repository Navigation Function
Shell functions SHALL enable one-command navigation to canonical repositories.

#### Scenario: Navigate to canonical repo with yr
- **WHEN** user runs `yr backend` in shell with functions loaded
- **THEN** current directory changes to canonical repo path
- **AND** repo is accessible for manual operations

#### Scenario: Error on nonexistent repo
- **WHEN** user runs `yr nonexistent` for repo that doesn't exist
- **THEN** error message is displayed
- **AND** current directory remains unchanged

### Requirement: Shell Integration Documentation
The system SHALL provide comprehensive documentation for setting up and using shell integration.

#### Scenario: Documentation includes all supported shells
- **WHEN** user reads docs/shell-integration.md
- **THEN** instructions are provided for bash, zsh, and fish
- **AND** each shell has copy-paste installation steps
- **AND** examples demonstrate common workflows

#### Scenario: README links to shell integration
- **WHEN** user reads README.md
- **THEN** shell integration section is prominently featured
- **AND** one-liner installation command is provided
- **AND** link to full documentation is included
