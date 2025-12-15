## 1. Schema Updates
- [ ] 1.1 Add `Version int` field to `domain.Workspace` struct
- [ ] 1.2 Define current schema version constant (e.g., `CurrentWorkspaceVersion = 1`)

## 2. Load/Save Updates
- [ ] 2.1 Update workspace save to always write current version
- [ ] 2.2 Update workspace load to default missing version to 0
- [ ] 2.3 Add version validation on load (warn if unknown future version)

## 3. Migration Framework
- [ ] 3.1 Add migration registry for version-to-version migrations
- [ ] 3.2 Implement auto-upgrade from version 0 to 1 (no-op, just adds version field)
- [ ] 3.3 Add tests for migration path

## 4. Export/Import Updates
- [ ] 4.1 Include version in workspace export
- [ ] 4.2 Validate version compatibility on import
- [ ] 4.3 Add tests for export/import versioning

## 5. Documentation
- [ ] 5.1 Document workspace.yaml schema including version field

