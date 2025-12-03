# Implementation Tasks

## 1. Enhanced Status Data
- [x] 1.1 Add LastModified timestamp to Workspace struct
- [x] 1.2 Add DiskUsage field (in bytes)
- [x] 1.3 Implement CalculateDiskUsage() method
- [x] 1.4 Add BehindRemote count to RepoStatus
- [x] 1.5 Update Status() to check behind-remote commits

## 2. Stale Workspace Detection
- [x] 2.1 Add stale_threshold_days to config
- [x] 2.2 Implement IsStale() method on Workspace
- [x] 2.3 Read workspace directory mtime for last modified
- [x] 2.4 Add visual indicator (badge/icon) for stale workspaces

## 3. Disk Usage Tracking
- [x] 3.1 Implement recursive directory size calculation
- [x] 3.2 Format bytes as human-readable (MB/GB)
- [x] 3.3 Display per-workspace usage in list
- [x] 3.4 Show total usage in header/footer

## 4. Quick Actions
- [x] 4.1 Implement push-all action (p key)
- [x] 4.2 Add confirmation prompt for push-all
- [x] 4.3 Implement open-in-editor action (o key)
- [x] 4.4 Respect $EDITOR and $VISUAL environment variables
- [x] 4.5 Add loading spinner during push operations

## 5. Filtering & Search
- [x] 5.1 Add search mode triggered by / key
- [x] 5.2 Filter workspaces by ID substring
- [x] 5.3 Add filter for stale-only (s key toggle)
- [x] 5.4 Show filter status in header

## 6. Visual Enhancements
- [x] 6.1 Add color-coded health indicators
- [x] 6.2 Show behind-remote badge for repos
- [x] 6.3 Add summary statistics header
- [x] 6.4 Improve item rendering with icons/badges

## 7. Testing
- [x] 7.1 Manual testing of all new quick actions
- [x] 7.2 Test disk usage calculation accuracy
- [x] 7.3 Test stale detection with various mtimes
- [x] 7.4 Test filtering and search
