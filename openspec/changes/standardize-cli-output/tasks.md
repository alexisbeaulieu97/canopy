## 1. Create Output Utilities
- [ ] 1.1 Create `internal/output/colors.go` with centralized styles
- [ ] 1.2 Add TTY detection for color output
- [ ] 1.3 Create formatting constants (SeparatorWidth, etc.)
- [ ] 1.4 Add helper functions for common patterns

## 2. Standardize JSON Output
- [ ] 2.1 Create `OutputFormatter` interface
- [ ] 2.2 Implement consistent JSON error format
- [ ] 2.3 Wire JSON error output through all commands
- [ ] 2.4 Remove unused `formatErrorJSON` or integrate it

## 3. Migrate Existing Code
- [ ] 3.1 Update `cmd/canopy/repo.go` to use centralized colors
- [ ] 3.2 Update `cmd/canopy/doctor.go` to use centralized colors
- [ ] 3.3 Update `cmd/canopy/presenters.go` to use new utilities
- [ ] 3.4 Remove duplicate color constants

## 4. Formatting Consistency
- [ ] 4.1 Consolidate separator width constants
- [ ] 4.2 Use consistent separator character (─ not -)
- [ ] 4.3 Apply consistent table formatting

## 5. Testing
- [ ] 5.1 Add unit tests for color utilities
- [ ] 5.2 Test TTY vs non-TTY output

## 6. Documentation
- [ ] 6.1 Document color scheme in CONTRIBUTING.md
