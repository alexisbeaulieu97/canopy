## ADDED Requirements

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

