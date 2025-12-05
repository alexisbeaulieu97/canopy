# Tasks: Add TUI Keyboard Customization

## Implementation Checklist

### 1. Configuration Schema
- [ ] 1.1 Define `TUIConfig` struct:
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
- [ ] 1.2 Add `TUI TUIConfig` to main Config
- [ ] 1.3 Set defaults in config loading

### 2. Keybinding Validation
- [ ] 2.1 Define default keybindings as constants
- [ ] 2.2 Apply defaults when config values are empty
- [ ] 2.3 Validate keybinding strings are valid key names
- [ ] 2.4 Detect conflicting keybindings (same key assigned to multiple actions)
- [ ] 2.5 Return config validation error listing all conflicts
- [ ] 2.6 Add unit tests for conflict detection

### 3. Model Integration
- [ ] 3.1 Pass keybindings to `NewModel()`
- [ ] 3.2 Store keybindings in Model struct
- [ ] 3.3 Create `matchesKey(key string, bindings []string) bool` helper

### 4. Update Handler Changes
- [ ] 4.1 Refactor `handleListKey()` to use configurable bindings
- [ ] 4.2 Refactor `handleDetailKey()` to use configurable bindings
- [ ] 4.3 Refactor `handleConfirmKey()` to use configurable bindings
- [ ] 4.4 Replace hardcoded key checks with `matchesKey()`

### 5. View Updates
- [ ] 5.1 Update footer to show configured keys
- [ ] 5.2 Update help text dynamically
- [ ] 5.3 Show first key from each binding in shortcuts

### 6. Documentation
- [ ] 6.1 Document available actions and default keys
- [ ] 6.2 Add example configuration to docs
- [ ] 6.3 Document key name format (ctrl+c, shift+a, etc.)

### 7. Testing
- [ ] 7.1 Test custom keybindings are loaded
- [ ] 7.2 Test default keybindings work
- [ ] 7.3 Test multiple keys per action
- [ ] 7.4 Test invalid keybinding handling
- [ ] 7.5 Test conflicting keybindings are rejected
