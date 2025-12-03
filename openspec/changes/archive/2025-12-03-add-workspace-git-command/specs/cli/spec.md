# CLI Specification Deltas

## ADDED Requirements

### Requirement: Workspace Git Command
The `workspace git` command SHALL execute arbitrary git commands across all repos in a workspace.

#### Scenario: Run git fetch in all repos
- **GIVEN** workspace `PROJ-1` exists with repos `repo-a` and `repo-b`
- **WHEN** I run `canopy workspace git PROJ-1 fetch --all`
- **THEN** `git fetch --all` SHALL be executed in `repo-a`
- **AND** `git fetch --all` SHALL be executed in `repo-b`
- **AND** output from each repo SHALL be displayed with repo name header

#### Scenario: Run git status in all repos
- **GIVEN** workspace `PROJ-1` exists with multiple repos
- **WHEN** I run `canopy workspace git PROJ-1 status`
- **THEN** `git status` output SHALL be shown for each repo
- **AND** repos SHALL be clearly separated in output

#### Scenario: Sequential execution (default)
- **GIVEN** workspace with repos `repo-a`, `repo-b`, `repo-c`
- **WHEN** I run `canopy workspace git PROJ-1 pull`
- **THEN** git pull SHALL execute in `repo-a` first
- **AND** after `repo-a` completes, `repo-b` SHALL execute
- **AND** after `repo-b` completes, `repo-c` SHALL execute

#### Scenario: Parallel execution
- **GIVEN** workspace with multiple repos
- **WHEN** I run `canopy workspace git PROJ-1 fetch --parallel`
- **THEN** git fetch SHALL execute concurrently in all repos
- **AND** output SHALL be collected and displayed per repo

#### Scenario: Stop on first error
- **GIVEN** workspace with repos where one has merge conflicts
- **WHEN** I run `canopy workspace git PROJ-1 pull`
- **AND** the first repo fails
- **THEN** execution SHALL stop immediately
- **AND** exit code SHALL be non-zero

#### Scenario: Continue on error
- **GIVEN** workspace with repos where one will fail
- **WHEN** I run `canopy workspace git PROJ-1 pull --continue-on-error`
- **AND** one repo fails
- **THEN** execution SHALL continue to remaining repos
- **AND** summary SHALL show which repos failed
- **AND** exit code SHALL be non-zero
