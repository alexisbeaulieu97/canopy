# Tasks: Decouple Config Validation

## Implementation Checklist

### 1. Define New Methods
- [ ] 1.1 Create `ValidateValues()` method on `Config`
- [ ] 1.2 Check `CloseDefault` is "delete" or "archive"
- [ ] 1.3 Check regex patterns compile
- [ ] 1.4 Check `StaleThresholdDays >= 0`
- [ ] 1.5 Check required fields are non-empty
- [ ] 1.6 Create `ValidateEnvironment()` method on `Config`
- [ ] 1.7 Check paths exist
- [ ] 1.8 Check paths are directories
- [ ] 1.9 Optionally check paths are writable

### 2. Refactor Existing Validate
- [ ] 2.1 Update `Validate()` to call `ValidateValues()` then `ValidateEnvironment()`
- [ ] 2.2 Ensure backward compatibility
- [ ] 2.3 Document the two-phase validation in comments

### 3. Update validateRoot Helper
- [ ] 3.1 Rename to `validateRootPath()` for clarity
- [ ] 3.2 Move to `ValidateEnvironment()` only
- [ ] 3.3 Add `validateRequiredField()` helper for value checks

### 4. Add Tests
- [ ] 4.1 Add `TestValidateValues` with various invalid configs
- [ ] 4.2 Add `TestValidateEnvironment` with temp directories
- [ ] 4.3 Ensure existing `TestValidate` still passes
- [ ] 4.4 Test that invalid values fail before environment checks

### 5. Documentation
- [ ] 5.1 Add godoc comments explaining validation phases
- [ ] 5.2 Update any docs that reference Validate()
