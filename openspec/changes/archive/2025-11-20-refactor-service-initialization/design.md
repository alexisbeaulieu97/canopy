# Design: Service Initialization Refactoring

## Context
Currently, every command handler repeats the same service initialization logic:
```go
cfg, err := config.Load()
if err != nil { return err }
gitEngine := gitx.New(cfg.ProjectsRoot)
wsEngine := workspace.New(cfg.WorkspacesRoot)
service := workspaces.NewService(cfg, gitEngine, wsEngine, logger)
```

This appears 15+ times across the codebase, making it impossible to:
- Inject mock services for unit testing commands
- Add new dependencies without updating every command
- Manage service lifecycle (startup/shutdown hooks)

## Goals / Non-Goals

### Goals
- Centralize service initialization to a single location
- Enable dependency injection for testability
- Reduce code duplication across commands
- Make it easy to add new services in the future

### Non-Goals
- Change command-line interface or user-facing behavior
- Introduce complex DI framework (keep it simple)
- Modify service layer APIs

## Decisions

### Decision 1: App Struct Pattern
Use a simple App struct with methods to create commands, rather than a heavyweight DI container.

**Rationale:**
- Go idiom: simple, explicit, no magic
- Easy to understand and maintain
- Supports testing via struct field injection
- No external dependencies needed

**Alternative considered:** Wire/Dig frameworks
- **Rejected:** Overkill for this project's size; adds complexity and build-time dependencies

### Decision 2: Method Receivers for Commands
Commands will be methods on `*App` rather than closures.

```go
type App struct {
    cfg     *config.Config
    service *workspaces.Service
    logger  *logging.Logger
}

func (a *App) buildWorkspaceNewCmd() *cobra.Command {
    return &cobra.Command{
        Use: "new <ID>",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Use a.service directly
        },
    }
}
```

**Rationale:**
- Clear ownership: commands belong to App
- Easy to test: create App with mock service, call method
- Follows common cobra patterns in larger projects

**Alternative considered:** Closure pattern
```go
func buildWorkspaceNewCmd(service *workspaces.Service) *cobra.Command { ... }
```
- **Rejected:** Less testable (can't access command internals), requires passing many parameters

### Decision 3: Lazy Error Handling
App creation happens in PersistentPreRunE, not main(), to preserve current error behavior where config issues show helpful messages rather than panicking.

**Rationale:**
- Maintains current UX (user sees "config not found" rather than panic)
- Allows `yard init` to run before config exists
- Cobra's error handling already works well

### Decision 4: No Singleton Pattern
App instance is created once in main() and passed through, not a global singleton.

**Rationale:**
- Testable: each test can create its own App
- Explicit dependencies
- No hidden global state

## Implementation Approach

### Phase 1: Create App Infrastructure
```go
// internal/app/app.go
package app

type App struct {
    cfg     *config.Config
    service *workspaces.Service
    logger  *logging.Logger
}

func New(debug bool) (*App, error) {
    cfg, err := config.Load()
    if err != nil {
        return nil, err
    }

    logger := logging.New(debug)
    gitEngine := gitx.New(cfg.ProjectsRoot)
    wsEngine := workspace.New(cfg.WorkspacesRoot)

    return &App{
        cfg:     cfg,
        service: workspaces.NewService(cfg, gitEngine, wsEngine, logger),
        logger:  logger,
    }, nil
}
```

### Phase 2: Update Main
```go
// cmd/yard/main.go
func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

var rootCmd = &cobra.Command{
    Use: "yard",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        app, err := app.New(debug)
        if err != nil {
            return err
        }
        cmd.SetContext(context.WithValue(cmd.Context(), appKey, app))
        return nil
    },
}
```

### Phase 3: Update Commands
```go
// cmd/yard/workspace.go
func init() {
    rootCmd.AddCommand(buildWorkspaceCmd())
}

func buildWorkspaceCmd() *cobra.Command {
    cmd := &cobra.Command{Use: "workspace"}
    cmd.AddCommand(buildWorkspaceNewCmd())
    return cmd
}

func buildWorkspaceNewCmd() *cobra.Command {
    return &cobra.Command{
        Use: "new <ID>",
        RunE: func(cmd *cobra.Command, args []string) error {
            app := cmd.Context().Value(appKey).(*App)
            // Use app.service, app.cfg, app.logger
        },
    }
}
```

## Risks / Trade-offs

### Risk: Breaking Internal Tests
**Mitigation:** Update tests incrementally, one command file at a time. Keep integration tests as safety net.

### Risk: Context Type Assertion Panics
**Mitigation:** Add helper `mustGetApp(cmd)` with clear panic message. Consider using cobra's built-in context passing.

### Trade-off: Slightly More Boilerplate in init()
Previously commands were defined inline; now they're built by functions. This is acceptable for the testability gain.

## Migration Plan

1. Create `internal/app/app.go` with tests
2. Update `main.go` and verify `yard --version` still works
3. Migrate workspace commands one-by-one, running tests after each
4. Migrate repo commands
5. Migrate remaining commands
6. Remove old global variables
7. Update documentation

### Rollback Strategy
Each commit migrates one command file. Can revert individual commits if issues arise.

## Open Questions
- Should App expose individual engines (gitEngine, wsEngine) or only the service? **Decision:** Only expose service and logger; engines are implementation details.
- Do we need graceful shutdown hooks? **Decision:** Not yet; add when needed (e.g., for future daemon mode).
