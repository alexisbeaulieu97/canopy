## ADDED Requirements

### Requirement: No Deprecated Code Aliases
The codebase SHALL NOT contain deprecated type aliases or wrapper functions. All code SHALL use canonical implementations directly.

#### Scenario: Domain types used directly
- **WHEN** code needs to reference domain types like Workspace, ClosedWorkspace, HookContext
- **THEN** the code SHALL import and use `domain.TypeName` directly
- **AND** package-local aliases SHALL NOT be defined

#### Scenario: Utility functions used directly
- **WHEN** code needs URL parsing utilities
- **THEN** the code SHALL import and use `giturl.FunctionName` directly
- **AND** wrapper functions SHALL NOT be defined in other packages

