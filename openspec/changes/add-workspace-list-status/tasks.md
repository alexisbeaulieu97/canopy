## 1. CLI Command
- [ ] 1.1 Add `--status` flag to `workspaceListCmd` in `cmd/canopy/workspace.go`
- [ ] 1.2 Add `--timeout` flag for status check timeout (default: 5s per repo)
- [ ] 1.3 Fetch git status for each workspace when `--status` is set
- [ ] 1.4 Format status indicators (dirty count, ahead/behind counts)
- [ ] 1.5 Support `--json` output with status data
- [ ] 1.6 Add integration test for list with status

## 2. Documentation
- [ ] 2.1 Update README.md command reference
- [ ] 2.2 Update docs/usage.md with --status examples
