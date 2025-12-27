## 1. Design System Foundation

- [ ] 1.1 Define new color palette constants in `components/components.go`
- [ ] 1.2 Create box/panel style presets (header, content, footer, modal)
- [ ] 1.3 Define spacing constants (padding, margins, gaps)
- [ ] 1.4 Update symbol set with Nerd Font icons (with ASCII fallback)

## 2. Header Redesign

- [ ] 2.1 Create `renderHeader()` component with logo, title, and summary stats
- [ ] 2.2 Add breadcrumb trail showing current view context
- [ ] 2.3 Integrate filter/search status indicators into header

## 3. List View Overhaul

- [ ] 3.1 Refactor `WorkspaceDelegate.Render` for two-line compact layout
- [ ] 3.2 Create inline status pill components (dirty, behind, stale)
- [ ] 3.3 Add alternating row backgrounds for readability
- [ ] 3.4 Update selection highlighting with accent border

## 4. Detail View Redesign

- [ ] 4.1 Create card-based layout with bordered sections
- [ ] 4.2 Redesign repository list with status columns
- [ ] 4.3 Add visual grouping for metadata vs content sections
- [ ] 4.4 Create orphan warning banner component

## 5. Confirmation Dialog Styling

- [ ] 5.1 Create centered modal overlay component
- [ ] 5.2 Add icon and color coding based on action type
- [ ] 5.3 Style confirmation buttons with proper highlighting

## 6. Footer Help Bar

- [ ] 6.1 Create context-aware footer component
- [ ] 6.2 Define key legend formatting (key: action pairs)
- [ ] 6.3 Add pagination/scroll indicators when applicable

## 7. Testing & Polish

- [ ] 7.1 Test rendering at various terminal sizes (80x24 minimum)
- [ ] 7.2 Verify ASCII fallback mode renders correctly
- [ ] 7.3 Test color scheme in light and dark terminals
- [ ] 7.4 Update snapshot tests for new layouts
