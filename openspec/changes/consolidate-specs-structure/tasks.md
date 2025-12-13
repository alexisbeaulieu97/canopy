# Tasks: Consolidate Specs Structure

## Implementation Checklist

### 1. Merge TUI Specs
- [ ] 1.1 Review all requirements in `tui/spec.md`
- [ ] 1.2 Review all requirements in `tui-interface/spec.md`
- [ ] 1.3 Identify duplicate requirements
- [ ] 1.4 Merge unique requirements from `tui-interface` into `tui`
- [ ] 1.5 Update `tui/spec.md` purpose section
- [ ] 1.6 Delete `tui-interface/` directory

### 2. Clarify Core vs Core-Architecture
- [ ] 2.1 Review `core/spec.md` content
- [ ] 2.2 Review `core-architecture/spec.md` content
- [ ] 2.3 Update `core/spec.md` purpose to focus on domain rules
- [ ] 2.4 Update `core-architecture/spec.md` purpose to focus on patterns
- [ ] 2.5 Move any misplaced requirements to correct spec

### 3. Validation
- [ ] 3.1 Run `openspec validate --strict` on all specs
- [ ] 3.2 Verify no broken references
- [ ] 3.3 Update any openspec changes that reference removed specs

