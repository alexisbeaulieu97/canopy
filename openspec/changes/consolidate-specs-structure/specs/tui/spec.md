## MODIFIED Requirements

### Requirement: Specification Scope
The TUI specification SHALL serve as the single source of truth for all TUI functionality including interactive navigation, status display, keyboard shortcuts, visual indicators, and reusable components.

#### Scenario: Single source of truth
- **WHEN** checking specifications for a TUI-related requirement, **THEN** all TUI requirements SHALL be in `tui/spec.md`
- **WHEN** the consolidation is complete, **THEN** `tui-interface` spec SHALL no longer exist

## REMOVED Requirements
### Requirement: tui-interface Spec
**Reason**: Consolidated into `tui/spec.md` to eliminate duplication.
**Migration**: All requirements from `tui-interface` have been merged into the main `tui` spec.

#### Scenario: Requirement removal validation
- **WHEN** this requirement was removed, **THEN** it was because all functionality is now covered by `tui/spec.md`
- **WHEN** referencing TUI interface behavior, **THEN** use `tui/spec.md` as the authoritative source

