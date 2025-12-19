## 1. Core Implementation
- [ ] 1.1 Add `GetWorkspaceStatusBatch` method to `workspaces.Service` that accepts slice of workspace IDs
- [ ] 1.2 Implement worker pool using `errgroup` with bounded concurrency
- [ ] 1.3 Use channel to collect results and maintain ordering
- [ ] 1.4 Respect `parallel_workers` config setting for worker count

## 2. CLI Integration
- [ ] 2.1 Refactor `workspace list --status` to use batch status method
- [ ] 2.2 Add `--sequential-status` flag to force sequential fetching
- [ ] 2.3 Ensure output order is deterministic (sort by workspace ID)
- [ ] 2.4 Update JSON output to include status alongside each workspace

## 3. Testing
- [ ] 3.1 Add unit tests for `GetWorkspaceStatusBatch` with mock git operations
- [ ] 3.2 Add test for result ordering guarantee
- [ ] 3.3 Add integration test comparing sequential vs parallel output equality
- [ ] 3.4 Add benchmark test measuring speedup with multiple workspaces

## 4. Documentation
- [ ] 4.1 Update `docs/usage.md` to mention parallel status fetching
- [ ] 4.2 Document `--sequential-status` flag and when to use it
