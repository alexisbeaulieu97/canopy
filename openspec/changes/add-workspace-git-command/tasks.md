# Implementation Tasks

## 1. Core Git Execution
- [x] 1.1 Add `RunCommand(repoPath, args ...string)` to gitx package
- [x] 1.2 Capture stdout and stderr separately
- [x] 1.3 Return structured result with exit code

## 2. Service Layer
- [x] 2.1 Add `RunGitInWorkspace(id string, args []string, opts RunOpts)` method
- [x] 2.2 Implement sequential execution (default)
- [x] 2.3 Implement parallel execution with goroutines
- [x] 2.4 Collect results per repo

## 3. CLI Command
- [x] 3.1 Create `workspaceGitCmd` cobra command
- [x] 3.2 Add `--parallel` flag (default: false)
- [x] 3.3 Add `--continue-on-error` flag (default: false)
- [x] 3.4 Parse remaining args as git command

## 4. Output Formatting
- [x] 4.1 Show repo name header before each output
- [x] 4.2 Use colors to distinguish repos
- [x] 4.3 Summarize success/failure count at end
- [x] 4.4 Show clear error messages for failed repos

## 5. Error Handling
- [x] 5.1 Stop on first error (default behavior)
- [x] 5.2 Continue and collect all errors with `--continue-on-error`
- [x] 5.3 Exit with appropriate code (0 = all success, 1 = any failure)

## 6. Testing
- [x] 6.1 Test basic command execution (git status)
- [x] 6.2 Test parallel execution
- [x] 6.3 Test error handling (one repo fails)
- [x] 6.4 Test continue-on-error behavior
