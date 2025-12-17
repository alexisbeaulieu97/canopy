# Change: Add Comprehensive Integration Test Coverage

## Why

The current integration test suite covers only basic workspace lifecycle operations. Critical flows like workspace restore, rename, branch operations, and orphan detection lack end-to-end test coverage. This increases regression risk during refactoring and makes it harder to validate the complete user journey.

## What Changes

- Add integration tests for workspace restore workflow (close -> restore cycle)
- Add integration tests for workspace rename operations
- Add integration tests for branch switching across repos
- Add integration tests for orphan detection and handling
- Add integration tests for parallel git operations
- Add test helper utilities to reduce boilerplate
- Improve test isolation and cleanup

## Impact

- Affected specs: core-architecture (testing requirements)
- Affected code: `test/integration/`
- Risk: Low - additive changes only, no production code modifications
