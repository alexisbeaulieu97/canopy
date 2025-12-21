## 1. CLI Command
- [x] 1.1 Add `--status` flag to `workspaceListCmd` in `cmd/canopy/workspace.go`
- [x] 1.2 Add `--timeout` flag for status check timeout (default: 5s per repo)
- [x] 1.3 Fetch git status for each workspace when `--status` is set
- [x] 1.4 Format status indicators (dirty count, ahead/behind counts)
- [x] 1.5 Support `--json` output with status data
- [x] 1.6 Add integration test for list with status

## 2. Documentation
- [x] 2.1 Update README.md command reference
- [x] 2.2 Update docs/usage.md with --status examples
