## 1. Design

- [ ] 1.1 Determine best location for helper (cmd/canopy/repo.go local or internal/config/)
- [ ] 1.2 Define function signature that handles all three use cases

## 2. Implementation

- [ ] 2.1 Create `saveRegistryWithRollback` helper function
- [ ] 2.2 Update `repoAddCmd` to use the helper (lines 103-114)
- [ ] 2.3 Update `repoRegisterCmd` to use the helper (lines 215-225)
- [ ] 2.4 Update `repoUnregisterCmd` to use the helper (lines 250-261)

## 3. Testing

- [ ] 3.1 Add unit test for the helper function with successful save
- [ ] 3.2 Add unit test for the helper function with save failure and successful rollback
- [ ] 3.3 Add unit test for the helper function with save failure and rollback failure
- [ ] 3.4 Run integration tests to verify existing behavior unchanged
