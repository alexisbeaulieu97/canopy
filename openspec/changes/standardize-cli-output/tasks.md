## 1. Create Output Utilities
- [x] 1.1 Create `internal/output/colors.go` with centralized styles
- [x] 1.2 Add TTY detection for color output
- [x] 1.3 Create formatting constants (SeparatorWidth, etc.)
- [x] 1.4 Add helper functions for common patterns

## 2. Standardize JSON Output
- [x] 2.1 Create `OutputFormatter` interface
- [x] 2.2 Implement consistent JSON error format
- [x] 2.3 Wire JSON error output through all commands
- [x] 2.4 Remove unused `formatErrorJSON` or integrate it

## 3. Migrate Existing Code
- [x] 3.1 Update `cmd/canopy/repo.go` to use centralized colors
- [x] 3.2 Update `cmd/canopy/doctor.go` to use centralized colors
- [x] 3.3 Update `cmd/canopy/presenters.go` to use new utilities
- [x] 3.4 Remove duplicate color constants

## 4. Formatting Consistency
- [x] 4.1 Consolidate separator width constants
- [x] 4.2 Use consistent separator character (─ not -)
- [x] 4.3 Apply consistent table formatting

## 5. Testing
- [x] 5.1 Add unit tests for color utilities
- [x] 5.2 Test TTY vs non-TTY output

## 6. Documentation
- [x] 6.1 Document color scheme in CONTRIBUTING.md
