# Implementation Tasks

Note: This change documents shortcuts already implemented in `enhance-status-dashboard`. Tasks are marked complete as they exist in the codebase.

## 1. Push Shortcut
- [x] 1.1 Implement push handler (`p` key) with confirmation
- [x] 1.2 Show spinner during push operation
- [x] 1.3 Display operation result/error

## 2. Open Editor Shortcut
- [x] 2.1 Implement open in editor (`o` key)
- [x] 2.2 Respect `$VISUAL` and `$EDITOR` environment variables
- [x] 2.3 Handle editor types (GUI vs terminal)

## 3. Stale Filter Shortcut
- [x] 3.1 Implement stale filter toggle (`s` key)
- [x] 3.2 Apply filter to workspace list
- [x] 3.3 Show active filter in header

## 4. Search Filter
- [x] 4.1 Enable Bubble Tea list filtering (`/` key)
- [x] 4.2 Show search in help keys

## 5. Close Shortcut
- [x] 5.1 Implement close handler (`c` key) with confirmation
- [x] 5.2 Support y/n confirmation
- [x] 5.3 Reload workspace list after close

## 6. Testing
- [x] 6.1 Manual test push shortcut
- [x] 6.2 Manual test open editor shortcut
- [x] 6.3 Manual test stale filter toggle
- [x] 6.4 Manual test search filter
- [x] 6.5 Manual test close shortcut
