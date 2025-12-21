# Tasks: Add Doctor Command

## 1. Core Implementation
- [x] 1.1 Create `cmd/canopy/doctor.go` with command scaffold
- [x] 1.2 Implement git version check (minimum version, installed check)
- [x] 1.3 Implement config file validation check
- [x] 1.4 Implement directory permission checks
- [x] 1.5 Implement canonical repo health checks

## 2. Output and Flags
- [x] 2.1 Add severity levels (error, warning, info) to check results
- [x] 2.2 Implement human-readable output with status indicators
- [x] 2.3 Implement `--json` flag for JSON output
- [x] 2.4 Implement `--fix` flag for auto-remediation
- [x] 2.5 Add exit code mapping (0 = healthy, 1 = warnings, 2 = errors)

## 3. Testing
- [x] 3.1 Add unit tests for individual check functions
- [x] 3.2 Add integration tests for doctor command
- [x] 3.3 Test `--fix` behavior with missing directories

## 4. Documentation
- [x] 4.1 Add doctor command to usage.md
- [x] 4.2 Add troubleshooting section referencing doctor command
