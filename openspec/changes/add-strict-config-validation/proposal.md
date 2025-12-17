# Change: Add Strict Config Validation

## Why
Currently, the config loading uses `viper.Unmarshal()` which silently ignores unknown fields. This means typos in config keys go undetected:

```yaml
# Typo: 'parrallel_workers' instead of 'parallel_workers'
parrallel_workers: 8  # Silently ignored, default of 4 used
```

Users have no way to know their config is being ignored until they notice unexpected behavior. Strict validation would catch these issues immediately at startup.

Current code at `internal/config/config.go:266`:
```go
if err := viper.Unmarshal(&cfg); err != nil {
    return nil, cerrors.NewConfigInvalid(...)
}
```

## What Changes
- Use strict unmarshaling to detect unknown config fields
- Emit warnings for deprecated config keys with migration guidance
- Validate hook field values (timeout must be positive, shell must be non-empty if specified)
- Add config validation command: `canopy config validate`

## Impact
- Affected specs: `cli`
- Affected code:
  - `internal/config/config.go` - Use strict unmarshaling with mapstructure
  - `cmd/canopy/config.go` - Add validate subcommand (new file)
- **Breaking change**: Users with typos in config will see errors (desired behavior)
