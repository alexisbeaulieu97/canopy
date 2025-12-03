# Change: Enhance TUI Status Dashboard

## Why
The current TUI shows basic workspace information (ID, repo count) but lacks actionable insights. Users can't quickly identify stale workspaces, disk usage issues, or repos needing sync. Enhanced status provides at-a-glance health metrics and quick actions to improve daily workflow efficiency.

## What Changes
- Add stale workspace indicators (last modified > N days, configurable)
- Display total disk usage for worktrees in workspace list
- Show "behind remote" status for repos (needs pull)
- Add quick action: `p` key to push all repos in workspace
- Add quick action: `o` key to open workspace in $EDITOR
- Add workspace filtering: `/` key to search, `s` to show only stale
- Display summary statistics at top (total workspaces, total disk usage)
- Color-code workspace health (green=clean, yellow=needs attention, red=dirty)

## Impact
- Affected specs: `specs/tui-interface/spec.md`
- Affected code:
  - `internal/tui/tui.go:206-244` - Update View() with enhanced display
  - `internal/tui/tui.go:93-173` - Add new keyboard handlers
  - `internal/workspaces/service.go` - Add disk usage calculation
  - `internal/gitx/git.go:65-126` - Add behind-remote check to Status()
  - `internal/domain/domain.go:17-24` - Add LastModified and DiskUsage fields
