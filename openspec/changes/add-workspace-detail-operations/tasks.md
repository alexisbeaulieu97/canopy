## 1. Implementation

- [x] 1.1 Update `DetailViewState` to store reference to current workspace for operations
- [x] 1.2 Add key handlers in `handleDetailKeyAction` for `p` (push), `S` (sync), `o` (open), `c` (close)
- [x] 1.3 Modify `ConfirmViewState` to track parent view state for return-after-confirm
- [x] 1.4 Update `renderDetailView` to display available shortcuts in footer
- [x] 1.5 Handle state transitions: detail -> confirm -> detail (or list if closed)
- [x] 1.6 Add tests for detail view operation key handlers
- [x] 1.7 Add tests for confirm-from-detail state transitions
