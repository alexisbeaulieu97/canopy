## 1. Design System Foundation

- [x] 1.1 Define new color palette constants in `components/components.go`
- [x] 1.2 Create box/panel style presets (header, content, footer, modal)
- [x] 1.3 Define spacing constants (padding, margins, gaps)
- [x] 1.4 Update symbol set with Nerd Font icons (with ASCII fallback)

## 2. Header Redesign

- [x] 2.1 Create `renderHeader()` component with logo, title, and summary stats
- [x] 2.2 Add breadcrumb trail showing current view context
- [x] 2.3 Integrate filter/search status indicators into header

## 3. List View Overhaul

- [x] 3.1 Refactor `WorkspaceDelegate.Render` for two-line compact layout
- [x] 3.2 Create inline status pill components (dirty, behind, stale)
- [x] 3.3 Add alternating row backgrounds for readability
- [x] 3.4 Update selection highlighting with accent border

## 4. Detail View Redesign

- [x] 4.1 Create card-based layout with bordered sections
- [x] 4.2 Redesign repository list with status columns
- [x] 4.3 Add visual grouping for metadata vs content sections
- [x] 4.4 Create orphan warning banner component

## 5. Confirmation Dialog Styling

- [x] 5.1 Create centered modal overlay component
- [x] 5.2 Add icon and color coding based on action type
- [x] 5.3 Style confirmation buttons with proper highlighting

## 6. Footer Help Bar

- [x] 6.1 Create context-aware footer component
- [x] 6.2 Define key legend formatting (key: action pairs)
- [x] 6.3 Add pagination/scroll indicators when applicable

## 7. Testing & Polish

- [x] 7.1 Test rendering at various terminal sizes (80x24 minimum)
- [x] 7.2 Verify ASCII fallback mode renders correctly
- [x] 7.3 Test color scheme in light and dark terminals
- [x] 7.4 Update snapshot tests for new layouts
