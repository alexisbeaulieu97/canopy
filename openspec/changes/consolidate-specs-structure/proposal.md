# Change: Consolidate Specs Structure

## Why
The current specs structure has redundant specifications:
- `tui` and `tui-interface` have significant overlap and should be merged
- `core` and `core-architecture` distinction is unclear and causes confusion

This creates maintenance burden and confusion about which spec is authoritative.

## What Changes
- BREAKING: Merge `tui-interface` spec into `tui` spec
- BREAKING: Remove `tui-interface` directory after merge
- Add clear distinction between `core` (business rules) and `core-architecture` (patterns)
- Update spec purpose sections to clarify scope

## Impact
- Affected specs: tui, tui-interface, core, core-architecture
- No code changes required
- This is a documentation/organization change only

