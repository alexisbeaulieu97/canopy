# Change: Add Shell Integration Documentation and Helper Functions

## Why
One of Yardmaster's key value propositions is fast navigation to workspace directories. However, `yard workspace path` requires users to wrap it with `cd $(...)` which is cumbersome. Providing documented shell functions makes the tool dramatically more ergonomic for daily use, enabling instant workspace switching with simple commands like `yw PROJ-123`.

## What Changes
- Add `docs/shell-integration.md` with copy-paste shell functions for bash and zsh
- Add `yard shell-init` command that outputs shell function code
- Provide examples for: `yw <ID>` (cd to workspace), `yr <name>` (cd to canonical repo)
- Include instructions for adding to .bashrc/.zshrc
- Add shell integration section to README.md with quick start
- Include Fish shell support alongside bash/zsh

## Impact
- Affected specs: `specs/user-documentation/spec.md` (new capability)
- Affected code:
  - `cmd/yard/main.go` - Add `shell-init` command
  - `docs/shell-integration.md` (new) - Shell function documentation
  - `README.md` - Add shell integration quick start section
