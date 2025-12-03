```markdown
## MODIFIED Requirements

### Requirement: JSON Output
All CLI commands that produce output SHALL support `--json` flag for machine-readable output.

#### Scenario: Workspace status JSON
- **WHEN** I run `canopy workspace status PROJ-123 --json`
- **THEN** the output SHALL be valid JSON
- **AND** include workspace ID, branch, repos, and status

#### Scenario: Workspace path JSON
- **WHEN** I run `canopy workspace path PROJ-123 --json`
- **THEN** the output SHALL be `{"path": "/full/path/to/workspace"}`

#### Scenario: Repo list JSON
- **WHEN** I run `canopy repo list --json`
- **THEN** the output SHALL be a JSON array of repo objects

#### Scenario: Check JSON
- **WHEN** I run `canopy check --json`
- **THEN** the output SHALL be valid JSON with configuration status
```
