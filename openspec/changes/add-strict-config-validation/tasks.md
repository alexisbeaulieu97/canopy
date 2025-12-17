## 1. Implementation

- [ ] 1.1 Add mapstructure decode hook for strict field matching
- [ ] 1.2 Configure viper to use strict unmarshal (ErrorUnused, ErrorUnset options)
- [ ] 1.3 Add validation for hook fields (timeout > 0, shell non-empty if set)
- [ ] 1.4 Add `canopy config validate` command to check config without running other commands
- [ ] 1.5 Add warning mechanism for deprecated config keys

## 2. Testing

- [ ] 2.1 Add unit test: unknown config field triggers error
- [ ] 2.2 Add unit test: typo in known field triggers error
- [ ] 2.3 Add unit test: valid config passes strict validation
- [ ] 2.4 Add unit test: hook timeout validation
- [ ] 2.5 Add unit test: config validate command output

## 3. Documentation

- [ ] 3.1 Update configuration.md with strict validation behavior
- [ ] 3.2 Document common config mistakes and how to fix them
- [ ] 3.3 Document `canopy config validate` command
