# Tasks: Add TUI Keyboard Customization

## Implementation Checklist

### Phase 1: Configuration Schema
- [ ] Define `TUIConfig` struct:
  ```go
  type TUIConfig struct {
      Keybindings Keybindings `mapstructure:"keybindings"`
  }
  
  type Keybindings struct {
      Quit        []string `mapstructure:"quit"`
      Search      []string `mapstructure:"search"`
      Push        []string `mapstructure:"push"`
      Close       []string `mapstructure:"close"`
      OpenEditor  []string `mapstructure:"open_editor"`
      ToggleStale []string `mapstructure:"toggle_stale"`
      Details     []string `mapstructure:"details"`
      Confirm     []string `mapstructure:"confirm"`
      Cancel      []string `mapstructure:"cancel"`
  }
  ```
- [ ] Add `TUI TUIConfig` to main Config
- [ ] Set defaults in config loading

### Phase 2: Keybinding Defaults
- [ ] Define default keybindings as constants
- [ ] Apply defaults when config values are empty
- [ ] Validate keybinding strings are valid

### Phase 3: Model Integration
- [ ] Pass keybindings to `NewModel()`
- [ ] Store keybindings in Model struct
- [ ] Create `matchesKey(key string, bindings []string) bool` helper

### Phase 4: Update Handler Changes
- [ ] Refactor `handleListKey()` to use configurable bindings
- [ ] Refactor `handleDetailKey()` to use configurable bindings
- [ ] Refactor `handleConfirmKey()` to use configurable bindings
- [ ] Replace hardcoded key checks with `matchesKey()`

### Phase 5: View Updates
- [ ] Update footer to show configured keys
- [ ] Update help text dynamically
- [ ] Show first key from each binding in shortcuts

### Phase 6: Documentation
- [ ] Document available actions and default keys
- [ ] Add example configuration to docs
- [ ] Document key name format (ctrl+c, shift+a, etc.)

### Phase 7: Testing
- [ ] Test custom keybindings are loaded
- [ ] Test default keybindings work
- [ ] Test multiple keys per action
- [ ] Test invalid keybinding handling
