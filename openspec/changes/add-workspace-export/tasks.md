# Tasks: Add Workspace Export/Import

## Implementation Checklist

### 1. Define Export Schema
- [x] 1.1 Define `WorkspaceExport` struct:
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
- [x] 2.1 Add `workspaceExportCmd` subcommand
- [x] 2.2 Load workspace by ID
- [x] 2.3 Build `WorkspaceExport` from workspace data
- [x] 2.4 Output to stdout (pipe-friendly) or `--output <file>`
- [x] 2.5 Support `--format yaml|json` (default yaml)

### 3. Import Command
- [x] 3.1 Add `workspaceImportCmd` subcommand
- [x] 3.2 Read from file argument or stdin
- [x] 3.3 Parse YAML/JSON to `WorkspaceExport`
- [x] 3.4 Validate version compatibility
- [x] 3.5 Add `--id <new-id>` flag to override workspace ID
- [x] 3.6 Add `--branch <name>` flag to override branch

### 4. Import Logic
- [x] 4.1 For each repo in export, try to resolve via registry alias
- [x] 4.2 Fall back to URL if alias not found
- [x] 4.3 Clone canonical if needed
- [x] 4.4 Create workspace with resolved repos
- [x] 4.5 Handle conflicts (workspace already exists)

### 5. Error Handling
- [x] 5.1 Validate export file schema
- [x] 5.2 Handle missing repos gracefully
- [x] 5.3 Provide clear error messages for resolution failures
- [x] 5.4 Add `--force` to overwrite existing workspace

### 6. Testing
- [x] 6.1 Test export produces valid YAML
- [x] 6.2 Test import creates correct workspace
- [x] 6.3 Test round-trip: export then import
- [x] 6.4 Test import with missing repos
- [x] 6.5 Test import with ID override
