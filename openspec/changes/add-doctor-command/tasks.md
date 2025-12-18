# Tasks: Add Doctor Command

## 1. Core Implementation
- [ ] 1.1 Create `cmd/canopy/doctor.go` with command scaffold
- [ ] 1.2 Implement git version check (minimum version, installed check)
- [ ] 1.3 Implement config file validation check
- [ ] 1.4 Implement directory permission checks
- [ ] 1.5 Implement canonical repo health checks

## 2. Output and Flags
- [ ] 2.1 Add severity levels (error, warning, info) to check results
- [ ] 2.2 Implement human-readable output with status indicators
- [ ] 2.3 Implement `--json` flag for JSON output
- [ ] 2.4 Implement `--fix` flag for auto-remediation
- [ ] 2.5 Add exit code mapping (0 = healthy, 1 = warnings, 2 = errors)

## 3. Testing
- [ ] 3.1 Add unit tests for individual check functions
- [ ] 3.2 Add integration tests for doctor command
- [ ] 3.3 Test `--fix` behavior with missing directories

## 4. Documentation
- [ ] 4.1 Add doctor command to usage.md
- [ ] 4.2 Add troubleshooting section referencing doctor command
