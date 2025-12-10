# Tasks: Decouple Config Validation

## Implementation Checklist

### 1. Define New Methods
- [x] 1.1 Create `ValidateValues()` method on `Config`
- [x] 1.2 Check `CloseDefault` is "delete" or "archive"
- [x] 1.3 Check regex patterns compile
- [x] 1.4 Check `StaleThresholdDays >= 0`
- [x] 1.5 Check required fields are non-empty
- [x] 1.6 Create `ValidateEnvironment()` method on `Config`
- [x] 1.7 Check paths exist
- [x] 1.8 Check paths are directories
- [x] 1.9 Optionally check paths are writable

### 2. Refactor Existing Validate
- [x] 2.1 Update `Validate()` to call `ValidateValues()` then `ValidateEnvironment()`
- [x] 2.2 Ensure backward compatibility
- [x] 2.3 Document the two-phase validation in comments

### 3. Update validateRoot Helper
- [x] 3.1 Rename to `validateRootPath()` for clarity
- [x] 3.2 Move to `ValidateEnvironment()` only
- [x] 3.3 Add `validateRequiredField()` helper for value checks

### 4. Add Tests
- [x] 4.1 Add `TestValidateValues` with various invalid configs
- [x] 4.2 Add `TestValidateEnvironment` with temp directories
- [x] 4.3 Ensure existing `TestValidate` still passes
- [x] 4.4 Test that invalid values fail before environment checks

### 5. Documentation
- [x] 5.1 Add godoc comments explaining validation phases
- [x] 5.2 Update any docs that reference Validate()
