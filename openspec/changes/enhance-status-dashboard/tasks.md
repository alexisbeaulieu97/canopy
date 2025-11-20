# Implementation Tasks

## 1. Enhanced Status Data
- [ ] 1.1 Add LastModified timestamp to Workspace struct
- [ ] 1.2 Add DiskUsage field (in bytes)
- [ ] 1.3 Implement CalculateDiskUsage() method
- [ ] 1.4 Add BehindRemote count to RepoStatus
- [ ] 1.5 Update Status() to check behind-remote commits

## 2. Stale Workspace Detection
- [ ] 2.1 Add stale_threshold_days to config
- [ ] 2.2 Implement IsStale() method on Workspace
- [ ] 2.3 Read workspace directory mtime for last modified
- [ ] 2.4 Add visual indicator (badge/icon) for stale workspaces

## 3. Disk Usage Tracking
- [ ] 3.1 Implement recursive directory size calculation
- [ ] 3.2 Format bytes as human-readable (MB/GB)
- [ ] 3.3 Display per-workspace usage in list
- [ ] 3.4 Show total usage in header/footer

## 4. Quick Actions
- [ ] 4.1 Implement push-all action (p key)
- [ ] 4.2 Add confirmation prompt for push-all
- [ ] 4.3 Implement open-in-editor action (o key)
- [ ] 4.4 Respect $EDITOR and $VISUAL environment variables
- [ ] 4.5 Add loading spinner during push operations

## 5. Filtering & Search
- [ ] 5.1 Add search mode triggered by / key
- [ ] 5.2 Filter workspaces by ID substring
- [ ] 5.3 Add filter for stale-only (s key toggle)
- [ ] 5.4 Show filter status in header

## 6. Visual Enhancements
- [ ] 6.1 Add color-coded health indicators
- [ ] 6.2 Show behind-remote badge for repos
- [ ] 6.3 Add summary statistics header
- [ ] 6.4 Improve item rendering with icons/badges

## 7. Testing
- [ ] 7.1 Manual testing of all new quick actions
- [ ] 7.2 Test disk usage calculation accuracy
- [ ] 7.3 Test stale detection with various mtimes
- [ ] 7.4 Test filtering and search
