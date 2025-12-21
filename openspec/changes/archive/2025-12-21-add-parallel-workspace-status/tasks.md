## 1. Core Implementation
- [x] 1.1 Add `GetWorkspaceStatusBatch` method to `workspaces.Service` that accepts slice of workspace IDs
- [x] 1.2 Implement worker pool using `errgroup` with bounded concurrency
- [x] 1.3 Use channel to collect results and maintain ordering
- [x] 1.4 Respect `parallel_workers` config setting for worker count

## 2. CLI Integration
- [x] 2.1 Refactor `workspace list --status` to use batch status method
- [x] 2.2 Add `--sequential-status` flag to force sequential fetching
- [x] 2.3 Ensure output order is deterministic (sort by workspace ID)
- [x] 2.4 Update JSON output to include status alongside each workspace

## 3. Testing
- [x] 3.1 Add unit tests for `GetWorkspaceStatusBatch` with mock git operations
- [x] 3.2 Add test for result ordering guarantee
- [x] 3.3 Add integration test comparing sequential vs parallel output equality
- [x] 3.4 Add benchmark test measuring speedup with multiple workspaces

## 4. Documentation
- [x] 4.1 Update `docs/usage.md` to mention parallel status fetching
- [x] 4.2 Document `--sequential-status` flag and when to use it
