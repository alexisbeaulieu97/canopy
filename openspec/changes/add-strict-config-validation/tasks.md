## 1. Implementation

- [x] 1.1 Add mapstructure decode hook for strict field matching
- [x] 1.2 Configure viper to use strict unmarshal (ErrorUnused, ErrorUnset options)
- [x] 1.3 Add validation for hook fields (timeout > 0, shell non-empty if set)
- [x] 1.4 Add `canopy config validate` command to check config without running other commands
- [x] 1.5 Add warning mechanism for deprecated config keys

## 2. Testing

- [x] 2.1 Add unit test: unknown config field triggers error
- [x] 2.2 Add unit test: typo in known field triggers error
- [x] 2.3 Add unit test: valid config passes strict validation
- [x] 2.4 Add unit test: hook timeout validation
- [x] 2.5 Add unit test: config validate command output

## 3. Documentation

- [x] 3.1 Update configuration.md with strict validation behavior
- [x] 3.2 Document common config mistakes and how to fix them
- [x] 3.3 Document `canopy config validate` command
