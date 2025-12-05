# Tasks: Decouple Config Validation

## Implementation Checklist

### Phase 1: Define New Methods
- [ ] Create `ValidateValues()` method on `Config`
- [ ] Move pure validation logic:
  - [ ] Check `CloseDefault` is "delete" or "archive"
  - [ ] Check regex patterns compile
  - [ ] Check `StaleThresholdDays >= 0`
  - [ ] Check required fields are non-empty
- [ ] Create `ValidateEnvironment()` method on `Config`
- [ ] Move filesystem validation logic:
  - [ ] Check paths exist
  - [ ] Check paths are directories
  - [ ] Optionally check paths are writable

### Phase 2: Refactor Existing Validate
- [ ] Update `Validate()` to call `ValidateValues()` then `ValidateEnvironment()`
- [ ] Ensure backward compatibility
- [ ] Document the two-phase validation in comments

### Phase 3: Update validateRoot Helper
- [ ] Rename to `validateRootPath()` for clarity
- [ ] Move to `ValidateEnvironment()` only
- [ ] Add `validateRequiredField()` helper for value checks

### Phase 4: Add Tests
- [ ] Add `TestValidateValues` with various invalid configs
- [ ] Add `TestValidateEnvironment` with temp directories
- [ ] Ensure existing `TestValidate` still passes
- [ ] Test that invalid values fail before environment checks

### Phase 5: Documentation
- [ ] Add godoc comments explaining validation phases
- [ ] Update any docs that reference Validate()
