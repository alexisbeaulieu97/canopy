# Implementation Tasks

## 1. Model State
- [ ] 1.1 Add `filterDirty bool` to Model
- [ ] 1.2 Add `filterBehind bool` to Model
- [ ] 1.3 Add `showHelp bool` to Model
- [ ] 1.4 Add `operation string` for loading state feedback

## 2. Git Operations
- [ ] 2.1 Implement fetch handler (`f` key)
- [ ] 2.2 Implement pull handler (`P` key)
- [ ] 2.3 Implement push handler (`p` key) with confirmation
- [ ] 2.4 Show spinner during git operations
- [ ] 2.5 Display operation result/error

## 3. External Actions
- [ ] 3.1 Implement open in browser (`g` key)
- [ ] 3.2 Implement open in editor (`o` key)
- [ ] 3.3 Handle editor types (GUI vs terminal)

## 4. Filtering
- [ ] 4.1 Implement dirty filter toggle (`D` key)
- [ ] 4.2 Implement behind-remote filter toggle (`B` key)
- [ ] 4.3 Apply filters to workspace list
- [ ] 4.4 Show active filters in header/status bar

## 5. Help Overlay
- [ ] 5.1 Create help view with all shortcuts listed
- [ ] 5.2 Toggle help with `?` key
- [ ] 5.3 Dismiss help with `?` or `esc`
- [ ] 5.4 Style help overlay

## 6. Refresh
- [ ] 6.1 Implement refresh handler (`r` key)
- [ ] 6.2 Show loading spinner during refresh
- [ ] 6.3 Preserve selection after refresh if possible

## 7. Testing
- [ ] 7.1 Manual test all new shortcuts
- [ ] 7.2 Test filter combinations
- [ ] 7.3 Test help overlay toggle
