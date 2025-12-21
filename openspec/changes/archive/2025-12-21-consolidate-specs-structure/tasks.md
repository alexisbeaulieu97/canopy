# Tasks: Consolidate Specs Structure

## Implementation Checklist

### 1. Merge TUI Specs
- [x] 1.1 Review all requirements in `tui/spec.md`
- [x] 1.2 Review all requirements in `tui-interface/spec.md`
- [x] 1.3 Identify duplicate requirements
- [x] 1.4 Merge unique requirements from `tui-interface` into `tui`
- [x] 1.5 Update `tui/spec.md` purpose section
- [x] 1.6 Delete `tui-interface/` directory

### 2. Clarify Core vs Core-Architecture
- [x] 2.1 Review `core/spec.md` content
- [x] 2.2 Review `core-architecture/spec.md` content
- [x] 2.3 Update `core/spec.md` purpose to focus on domain rules
- [x] 2.4 Update `core-architecture/spec.md` purpose to focus on patterns
- [x] 2.5 Move any misplaced requirements to correct spec

### 3. Validation
- [x] 3.1 Run `openspec validate --strict` on all specs
- [x] 3.2 Verify no broken references
- [x] 3.3 Update any openspec changes that reference removed specs

