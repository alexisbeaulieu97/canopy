## 1. Audit

- [x] 1.1 Search codebase for all timeout values (time.Second, time.Minute patterns)
- [x] 1.2 Document current usage locations and purposes

## 2. Implementation

- [x] 2.1 Create or update constants file with named timeout constants
- [x] 2.2 Add `DefaultCleanupTimeout` constant (30s for repo cleanup operations)
- [x] 2.3 Add documentation comments explaining each timeout's purpose
- [x] 2.4 Update `cmd/canopy/repo.go:106` to use the named constant
- [x] 2.5 Update any other locations found in audit

## 3. Verification

- [x] 3.1 Verify no magic timeout numbers remain in cmd/ package
- [x] 3.2 Run tests to ensure behavior unchanged
