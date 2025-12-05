# Tasks: Add Workspace Export/Import

## Implementation Checklist

### Phase 1: Define Export Schema
- [ ] Define `WorkspaceExport` struct:
  ```go
  type WorkspaceExport struct {
      Version    string            `yaml:"version"`
      ID         string            `yaml:"id"`
      Branch     string            `yaml:"branch"`
      Repos      []RepoExport      `yaml:"repos"`
      ExportedAt time.Time         `yaml:"exported_at"`
      Notes      string            `yaml:"notes,omitempty"`
  }
  
  type RepoExport struct {
      Name  string `yaml:"name"`
      URL   string `yaml:"url"`
      Alias string `yaml:"alias,omitempty"`  // registry alias if known
  }
  ```

### Phase 2: Export Command
- [ ] Add `workspaceExportCmd` subcommand
- [ ] Load workspace by ID
- [ ] Build `WorkspaceExport` from workspace data
- [ ] Output to stdout (pipe-friendly) or `--output <file>`
- [ ] Support `--format yaml|json` (default yaml)

### Phase 3: Import Command
- [ ] Add `workspaceImportCmd` subcommand
- [ ] Read from file argument or stdin
- [ ] Parse YAML/JSON to `WorkspaceExport`
- [ ] Validate version compatibility
- [ ] Add `--id <new-id>` flag to override workspace ID
- [ ] Add `--branch <name>` flag to override branch

### Phase 4: Import Logic
- [ ] For each repo in export:
  - [ ] Try to resolve via registry alias
  - [ ] Fall back to URL
  - [ ] Clone canonical if needed
- [ ] Create workspace with resolved repos
- [ ] Handle conflicts (workspace already exists)

### Phase 5: Error Handling
- [ ] Validate export file schema
- [ ] Handle missing repos gracefully
- [ ] Provide clear error messages for resolution failures
- [ ] Add `--force` to overwrite existing workspace

### Phase 6: Testing
- [ ] Test export produces valid YAML
- [ ] Test import creates correct workspace
- [ ] Test round-trip: export then import
- [ ] Test import with missing repos
- [ ] Test import with ID override
