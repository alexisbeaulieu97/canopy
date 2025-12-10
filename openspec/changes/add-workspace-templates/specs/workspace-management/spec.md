# Workspace Management Specification Deltas

## ADDED Requirements

### Requirement: Workspace Templates
The system SHALL support user-defined workspace templates that specify default repositories and configuration for common workspace types.

#### Scenario: Define template in config
- **WHEN** user adds template to config.yaml with name, repos, and description
- **THEN** template is available for workspace creation
- **AND** template appears in `canopy template list` output

#### Scenario: Create workspace from template
- **WHEN** user runs `canopy workspace new PROJ-123 --template fullstack`
- **THEN** workspace is created with repositories defined in "fullstack" template
- **AND** default branch from template is used if no explicit branch specified
- **AND** workspace is ready for use

#### Scenario: Template with explicit repos
- **WHEN** user runs `canopy workspace new PROJ-123 --template backend --repos extra-lib`
- **THEN** workspace includes both template repos (backend, common) AND extra-lib
- **AND** all repos are cloned successfully

#### Scenario: Unknown template error
- **WHEN** user runs `canopy workspace new PROJ-123 --template nonexistent`
- **THEN** system returns error listing available templates
- **AND** no workspace is created

### Requirement: Template Configuration Format
Templates SHALL be defined in config.yaml with name, repos, optional default branch, and optional description.

#### Scenario: Parse template from config
- **WHEN** config.yaml contains templates section with valid YAML
- **THEN** templates are loaded into Config.Templates map
- **AND** each template is accessible by name

#### Scenario: Template with all fields
- **WHEN** template includes repos, default_branch, description, and setup_commands
- **THEN** all fields are parsed correctly
- **AND** template validation succeeds

#### Scenario: Invalid template configuration
- **WHEN** template in config.yaml has missing required fields (e.g., no repos)
- **THEN** config validation fails with clear error message
- **AND** indicates which template and field is problematic

### Requirement: Template Listing and Inspection
The system SHALL provide commands to list available templates and show template details.

#### Scenario: List all templates
- **WHEN** user runs `canopy template list`
- **THEN** system displays table of template names and descriptions
- **AND** shows repo count for each template
- **AND** templates are sorted alphabetically

#### Scenario: Show template details
- **WHEN** user runs `canopy template show fullstack`
- **THEN** system displays template name, description, repos list, and default branch
- **AND** indicates which repos are available in registry/canonical storage

### Requirement: Template Setup Commands
Templates SHALL support optional setup commands that execute after workspace creation to configure the environment.

#### Scenario: Execute template setup commands
- **WHEN** workspace is created from template with setup_commands defined
- **THEN** each command executes in workspace directory after repos are cloned
- **AND** commands run in order specified
- **AND** output is shown to user

#### Scenario: Setup command failure
- **WHEN** template setup command returns non-zero exit code
- **THEN** user is warned but workspace creation continues
- **AND** subsequent setup commands still execute
- **AND** workspace is marked as partially initialized

### Requirement: Template Composition
Templates SHALL be composable, allowing explicit repos to be added to template repos.

#### Scenario: Additive repository specification
- **WHEN** user specifies both --template and --repos flags
- **THEN** final repo list is union of template repos and explicit repos
- **AND** duplicates are removed (if same repo specified in both)
- **AND** all unique repos are included in workspace

#### Scenario: Template overrides branch
- **WHEN** template specifies default_branch but user provides --branch flag
- **THEN** user's explicit branch takes precedence
- **AND** template branch is ignored
