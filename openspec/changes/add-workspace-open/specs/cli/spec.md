# CLI Specification Deltas

## ADDED Requirements

### Requirement: Open Workspace in Editor
The `workspace open` command SHALL open the workspace directory in the user's configured editor.

#### Scenario: Open in default editor
- **GIVEN** workspace `PROJ-1` exists
- **AND** `$EDITOR` is set to `vim`
- **WHEN** I run `canopy workspace open PROJ-1`
- **THEN** vim SHALL open with the workspace directory

#### Scenario: Prefer VISUAL over EDITOR
- **GIVEN** `$VISUAL` is set to `code`
- **AND** `$EDITOR` is set to `vim`
- **WHEN** I run `canopy workspace open PROJ-1`
- **THEN** VS Code SHALL open (not vim)

#### Scenario: No editor configured
- **GIVEN** neither `$VISUAL` nor `$EDITOR` is set
- **WHEN** I run `canopy workspace open PROJ-1`
- **THEN** the command SHALL fail with error
- **AND** error message SHALL explain how to set `$EDITOR`

### Requirement: Open Workspace in Browser
The `workspace open --browser` command SHALL open repo remote URLs in the default browser.

#### Scenario: Open all repos in browser
- **GIVEN** workspace `PROJ-1` has repos `repo-a` and `repo-b`
- **AND** repos have GitHub remotes
- **WHEN** I run `canopy workspace open PROJ-1 --browser`
- **THEN** browser tabs SHALL open for both repo URLs

#### Scenario: Open specific repo in browser
- **GIVEN** workspace `PROJ-1` has repos `repo-a` and `repo-b`
- **WHEN** I run `canopy workspace open PROJ-1 --browser --repo=repo-a`
- **THEN** only `repo-a` URL SHALL open in browser

#### Scenario: SSH remote URL conversion
- **GIVEN** repo has remote `git@github.com:user/repo.git`
- **WHEN** I run `canopy workspace open PROJ-1 --browser`
- **THEN** browser SHALL open `https://github.com/user/repo`
