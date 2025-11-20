# Implementation Tasks

## 1. Create Shell Function Templates
- [ ] 1.1 Design bash/zsh function for `yw <workspace-id>`
- [ ] 1.2 Design bash/zsh function for `yr <repo-name>`
- [ ] 1.3 Create Fish shell equivalents
- [ ] 1.4 Add error handling for nonexistent workspaces/repos
- [ ] 1.5 Add shell completion hints

## 2. Implement shell-init Command
- [ ] 2.1 Create `yard shell-init` command in cmd/yard/
- [ ] 2.2 Detect shell from $SHELL environment variable
- [ ] 2.3 Output appropriate shell-specific functions
- [ ] 2.4 Add --shell flag for explicit shell selection (bash|zsh|fish)
- [ ] 2.5 Include installation instructions in output

## 3. Documentation
- [ ] 3.1 Create docs/shell-integration.md with full guide
- [ ] 3.2 Include installation instructions for each shell
- [ ] 3.3 Add troubleshooting section
- [ ] 3.4 Document function behavior and error cases
- [ ] 3.5 Add animated GIF demos to docs (optional)

## 4. README Updates
- [ ] 4.1 Add "Shell Integration" section to README
- [ ] 4.2 Include one-liner installation command
- [ ] 4.3 Show usage examples with yw/yr functions
- [ ] 4.4 Link to full shell-integration.md docs

## 5. Testing
- [ ] 5.1 Manually test bash functions on Linux/macOS
- [ ] 5.2 Manually test zsh functions on macOS
- [ ] 5.3 Test Fish shell functions
- [ ] 5.4 Verify shell-init output for each shell type
