## 1. Configuration

- [ ] 1.1 Add `UseEmoji bool` field to `TUIConfig` struct in `internal/config/config.go`
- [ ] 1.2 Set default value to `true` for backward compatibility
- [ ] 1.3 Add `GetUseEmoji()` method to config

## 2. Symbol Definitions

- [ ] 2.1 Create symbol mapping in `internal/tui/symbols.go`
- [ ] 2.2 Define emoji symbols: 🌲 💾 📂 ⚠ ✓ 🔍 ⏳ 📁
- [ ] 2.3 Define ASCII fallbacks: [W] [D] [>] [!] [*] [?] [...] [-]
- [ ] 2.4 Create `Symbols` type with method to get appropriate symbol based on config

## 3. TUI Updates

- [ ] 3.1 Pass emoji config to TUI Model
- [ ] 3.2 Update `renderHeader()` to use symbol mapping (lines 87, 96, 108, 120, 125)
- [ ] 3.3 Update `renderDetailView()` to use symbol mapping (line 186)
- [ ] 3.4 Update `renderRepoLine()` to use symbol mapping (lines 281, 287)
- [ ] 3.5 Update `renderDetailOrphans()` to use symbol mapping (lines 297, 303)
- [ ] 3.6 Update `renderFooter()` to use symbol mapping (line 134)

## 4. Documentation

- [ ] 4.1 Add `tui.use_emoji` option to `docs/configuration.md`
- [ ] 4.2 Document ASCII fallback characters

## 5. Testing

- [ ] 5.1 Add unit test verifying emoji output when enabled
- [ ] 5.2 Add unit test verifying ASCII output when disabled
