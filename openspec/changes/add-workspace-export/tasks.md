# Tasks: Add Workspace Export/Import

## Implementation Checklist

### 1. Define Export Schema
- [ ] 1.1 Define `WorkspaceExport` struct:
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

### 2. Export Command
- [ ] 2.1 Add `workspaceExportCmd` subcommand
- [ ] 2.2 Load workspace by ID
- [ ] 2.3 Build `WorkspaceExport` from workspace data
- [ ] 2.4 Output to stdout (pipe-friendly) or `--output <file>`
- [ ] 2.5 Support `--format yaml|json` (default yaml)

### 3. Import Command
- [ ] 3.1 Add `workspaceImportCmd` subcommand
- [ ] 3.2 Read from file argument or stdin
- [ ] 3.3 Parse YAML/JSON to `WorkspaceExport`
- [ ] 3.4 Validate version compatibility
- [ ] 3.5 Add `--id <new-id>` flag to override workspace ID
- [ ] 3.6 Add `--branch <name>` flag to override branch

### 4. Import Logic
- [ ] 4.1 For each repo in export, try to resolve via registry alias
- [ ] 4.2 Fall back to URL if alias not found
- [ ] 4.3 Clone canonical if needed
- [ ] 4.4 Create workspace with resolved repos
- [ ] 4.5 Handle conflicts (workspace already exists)

### 5. Error Handling
- [ ] 5.1 Validate export file schema
- [ ] 5.2 Handle missing repos gracefully
- [ ] 5.3 Provide clear error messages for resolution failures
- [ ] 5.4 Add `--force` to overwrite existing workspace

### 6. Testing
- [ ] 6.1 Test export produces valid YAML
- [ ] 6.2 Test import creates correct workspace
- [ ] 6.3 Test round-trip: export then import
- [ ] 6.4 Test import with missing repos
- [ ] 6.5 Test import with ID override
