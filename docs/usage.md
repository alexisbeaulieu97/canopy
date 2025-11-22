# Usage Guide

## Workflow

1.  **Create a Workspace**:
    When you pick up work (e.g., `PROJ-123`), run:
    ```bash
    canopy workspace new PROJ-123 --repos backend,frontend
    ```
    This creates a workspace at `~/workspaces/PROJ-123` (using the default config).

2.  **Work**:
    ```bash
    cd ~/workspaces/PROJ-123
    # Edit files in backend/ and frontend/
    ```
    You are automatically on branch `PROJ-123` in both repos.

3.  **Check Status**:
    ```bash
    canopy status
    ```

4.  **Finish**:
    Push your changes using standard git commands inside the worktrees.
    ```bash
    cd backend && git push origin PROJ-123
    ```

5.  **Cleanup**:
    ```bash
    canopy workspace archive PROJ-123
    ```
    This removes worktrees and keeps metadata in `~/.canopy/archives`. Use `canopy workspace restore PROJ-123` to recreate worktrees later, or `canopy workspace close PROJ-123` to delete without archiving.

    Use `--archive` / `--no-archive` on `workspace close` to control behavior without prompts (non-TTY runs never prompt).

## Configuration Notes

Key paths are set in `~/.canopy/config.yaml`:

```yaml
projects_root: ~/projects
workspaces_root: ~/workspaces
archives_root: ~/.canopy/archives
workspace_close_default: delete # set to archive to make archiving the default when no flags are provided
```

- `workspace_close_default` controls what `workspace close` does when you omit flags. Use `--archive` or `--no-archive` to override per command.
