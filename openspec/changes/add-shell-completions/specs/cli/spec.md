# cli Specification Delta

## ADDED Requirements

### Requirement: Shell Completion Support
The CLI SHALL provide shell completion scripts for common shells.

#### Scenario: Generate bash completion
- **WHEN** user runs `canopy completion bash`
- **THEN** a bash completion script is output to stdout
- **AND** script can be sourced to enable completions

#### Scenario: Generate zsh completion
- **WHEN** user runs `canopy completion zsh`
- **THEN** a zsh completion script is output to stdout
- **AND** script follows zsh completion conventions

#### Scenario: Generate fish completion
- **WHEN** user runs `canopy completion fish`
- **THEN** a fish completion script is output to stdout
- **AND** script follows fish completion conventions

#### Scenario: Dynamic workspace completion
- **GIVEN** workspaces `PROJ-1` and `PROJ-2` exist
- **WHEN** user types `canopy workspace close P<TAB>`
- **THEN** shell suggests `PROJ-1` and `PROJ-2`

#### Scenario: Dynamic repo completion
- **GIVEN** canonical repos `backend` and `frontend` exist
- **WHEN** user types `canopy repo status b<TAB>`
- **THEN** shell suggests `backend`
