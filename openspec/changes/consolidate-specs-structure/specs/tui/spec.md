## MODIFIED Requirements

### Requirement: Specification Scope
The TUI specification SHALL serve as the single source of truth for all TUI functionality including interactive navigation, status display, keyboard shortcuts, visual indicators, and reusable components.

#### Scenario: Single source of truth
- **GIVEN** a TUI-related requirement
- **WHEN** checking specifications
- **THEN** all TUI requirements SHALL be in `tui/spec.md`
- **AND** `tui-interface` spec SHALL no longer exist

## REMOVED Requirements
### Requirement: tui-interface Spec
**Reason**: Consolidated into `tui/spec.md` to eliminate duplication.
**Migration**: All requirements from `tui-interface` have been merged into the main `tui` spec.

