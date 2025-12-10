# tui-interface Specification Delta

## ADDED Requirements

### Requirement: Customizable Keyboard Bindings
The TUI SHALL support user-configurable keyboard bindings.

#### Scenario: Default keybindings work
- **GIVEN** no keybinding configuration exists
- **WHEN** user launches TUI
- **THEN** default keybindings are active (q=quit, j/k=navigate, etc.)

#### Scenario: Custom keybinding from config
- **GIVEN** config contains `tui.keybindings.delete: "d"`
- **WHEN** user presses "d" in TUI
- **THEN** delete action is triggered

#### Scenario: Override default keybinding
- **GIVEN** config contains `tui.keybindings.quit: "x"`
- **WHEN** user presses "x" in TUI
- **THEN** quit action is triggered
- **AND** "q" no longer triggers quit

#### Scenario: Keybinding conflict detection
- **GIVEN** config assigns same key to two actions
- **WHEN** config is validated
- **THEN** error is reported: "keybinding conflict: 'd' assigned to both 'delete' and 'details'"

#### Scenario: Invalid keybinding rejected
- **GIVEN** config contains invalid keybinding value
- **WHEN** config is validated
- **THEN** error is reported with the invalid value
