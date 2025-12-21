## ADDED Requirements

### Requirement: Specification Scope
The TUI specification SHALL serve as the single source of truth for all TUI functionality including interactive navigation, status display, keyboard shortcuts, visual indicators, and reusable components.

#### Scenario: Single source of truth
- **WHEN** checking specifications for a TUI-related requirement, **THEN** all TUI requirements SHALL be in `tui/spec.md`
- **WHEN** the consolidation is complete, **THEN** `tui-interface` spec SHALL no longer exist
