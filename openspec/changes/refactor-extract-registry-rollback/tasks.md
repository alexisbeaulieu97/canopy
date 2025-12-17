## 1. Design

- [x] 1.1 Determine best location for helper (cmd/canopy/repo.go local or internal/config/)
- [x] 1.2 Define function signature that handles all three use cases

## 2. Implementation

- [x] 2.1 Create `saveRegistryWithRollback` helper function
- [x] 2.2 Update `repoAddCmd` to use the helper (lines 103-114)
- [x] 2.3 Update `repoRegisterCmd` to use the helper (lines 215-225)
- [x] 2.4 Update `repoUnregisterCmd` to use the helper (lines 250-261)

## 3. Testing

- [x] 3.1 Add unit test for the helper function with successful save
- [x] 3.2 Add unit test for the helper function with save failure and successful rollback
- [x] 3.3 Add unit test for the helper function with save failure and rollback failure
- [x] 3.4 Run integration tests to verify existing behavior unchanged
