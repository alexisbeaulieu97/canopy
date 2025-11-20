# Implementation Tasks

## 1. Wizard UI Components
- [ ] 1.1 Create internal/tui/wizard.go with WizardModel
- [ ] 1.2 Implement multi-step state machine (ID → Template → Repos → Branch → Confirm)
- [ ] 1.3 Add text input component for workspace ID
- [ ] 1.4 Add list selection component for templates
- [ ] 1.5 Add multi-select component for repositories
- [ ] 1.6 Add confirmation screen with summary

## 2. Wizard Logic
- [ ] 2.1 Implement real-time validation for workspace ID
- [ ] 2.2 Check for existing workspaces and show warning
- [ ] 2.3 Load available templates from config
- [ ] 2.4 Load available repos from registry + canonical storage
- [ ] 2.5 Apply template repos when template selected
- [ ] 2.6 Allow skipping template selection

## 3. CLI Integration
- [ ] 3.1 Add `yard workspace create` command
- [ ] 3.2 Add --wizard flag to `yard workspace new`
- [ ] 3.3 Wire wizard UI to workspace service
- [ ] 3.4 Pass collected inputs to CreateWorkspace()
- [ ] 3.5 Show progress spinner during creation

## 4. UX Polish
- [ ] 4.1 Add keyboard shortcuts help at bottom of screen
- [ ] 4.2 Use color coding for validation states (red=error, green=valid)
- [ ] 4.3 Show repo count in multi-select header
- [ ] 4.4 Add back/cancel navigation between steps
- [ ] 4.5 Display creation success message with path

## 5. Testing
- [ ] 5.1 Manual testing of full wizard flow
- [ ] 5.2 Test wizard with various input combinations
- [ ] 5.3 Test validation edge cases
- [ ] 5.4 Test keyboard navigation
- [ ] 5.5 Verify integration with existing workspace creation
