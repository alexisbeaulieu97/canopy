# Tasks: Add Hook Dry-Run Mode

## 1. Executor Changes
- [x] 1.1 Add `DryRun` option to `ExecuteHooks` method
- [x] 1.2 When dry-run enabled, collect and return commands without executing
- [x] 1.3 Return resolved command strings with all variables substituted

## 2. CLI Flags
- [x] 2.1 Add `--dry-run-hooks` flag to `workspace new` command
- [x] 2.2 Add `--dry-run-hooks` flag to `workspace close` command
- [x] 2.3 Display hook preview in human-readable format
- [x] 2.4 Support `--json` output for hook preview

## 3. Hooks Subcommand
- [x] 3.1 Create `cmd/canopy/hooks.go` with hooks command group
- [x] 3.2 Implement `canopy hooks list` to show configured hooks
- [x] 3.3 Implement `canopy hooks test <event> --workspace <id>` for targeted testing

## 4. Testing
- [x] 4.1 Add unit tests for dry-run executor behavior
- [x] 4.2 Add integration tests for CLI dry-run flags
- [x] 4.3 Test variable resolution in dry-run output

## 5. Documentation
- [x] 5.1 Update hooks.md with dry-run usage examples
- [x] 5.2 Add troubleshooting section for hook debugging
