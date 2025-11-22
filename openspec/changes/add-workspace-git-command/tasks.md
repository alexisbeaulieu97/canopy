# Implementation Tasks

## 1. Core Git Execution
- [ ] 1.1 Add `RunCommand(repoPath, args ...string)` to gitx package
- [ ] 1.2 Capture stdout and stderr separately
- [ ] 1.3 Return structured result with exit code

## 2. Service Layer
- [ ] 2.1 Add `RunGitInWorkspace(id string, args []string, opts RunOpts)` method
- [ ] 2.2 Implement sequential execution (default)
- [ ] 2.3 Implement parallel execution with goroutines
- [ ] 2.4 Collect results per repo

## 3. CLI Command
- [ ] 3.1 Create `workspaceGitCmd` cobra command
- [ ] 3.2 Add `--parallel` flag (default: false)
- [ ] 3.3 Add `--continue-on-error` flag (default: false)
- [ ] 3.4 Parse remaining args as git command

## 4. Output Formatting
- [ ] 4.1 Show repo name header before each output
- [ ] 4.2 Use colors to distinguish repos
- [ ] 4.3 Summarize success/failure count at end
- [ ] 4.4 Show clear error messages for failed repos

## 5. Error Handling
- [ ] 5.1 Stop on first error (default behavior)
- [ ] 5.2 Continue and collect all errors with `--continue-on-error`
- [ ] 5.3 Exit with appropriate code (0 = all success, 1 = any failure)

## 6. Testing
- [ ] 6.1 Test basic command execution (git status)
- [ ] 6.2 Test parallel execution
- [ ] 6.3 Test error handling (one repo fails)
- [ ] 6.4 Test continue-on-error behavior
