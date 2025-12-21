# Change: Add Strict Config Validation

## Why
Currently, `viper.Unmarshal()` silently ignores unknown config fields, allowing typos to go undetected. Strict validation would catch these issues immediately at startup.

## What Changes
- **BREAKING** Use strict unmarshaling to detect unknown config fields
- Emit warnings for deprecated config keys with migration guidance
- Validate hook field values (timeout must be positive, shell must be non-empty if specified)
- Add config validation command: `canopy config validate`

## Impact
- Affected specs: `cli`
- Affected code:
  - `internal/config/config.go` - Use strict unmarshaling with mapstructure
  - `cmd/canopy/config.go` - Add validate subcommand (new file)
