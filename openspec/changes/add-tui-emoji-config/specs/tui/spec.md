## ADDED Requirements

### Requirement: Configurable TUI Symbols

The TUI SHALL support configurable symbol rendering to accommodate terminals with limited Unicode support.

Users SHALL be able to disable emoji rendering via the `tui.use_emoji` configuration option.

When emoji is disabled, ASCII fallback characters SHALL be used for all decorative symbols.

#### Scenario: Emoji enabled (default)

- **GIVEN** `tui.use_emoji` is true or not specified
- **WHEN** rendering the TUI
- **THEN** emoji characters SHALL be displayed (🌲, 💾, 📂, etc.)

#### Scenario: Emoji disabled

- **GIVEN** `tui.use_emoji` is false
- **WHEN** rendering the TUI
- **THEN** ASCII fallback characters SHALL be displayed ([W], [D], [>], etc.)

#### Scenario: Backward compatibility

- **GIVEN** no `tui.use_emoji` configuration is specified
- **WHEN** rendering the TUI
- **THEN** emoji characters SHALL be displayed (preserving current behavior)
