## 1. Flag Implementation
- [x] 1.1 Add `--config` persistent flag to root command in `main.go`
- [x] 1.2 Pass config path to App initialization
- [x] 1.3 Update config loader to accept optional path parameter

## 2. Environment Variable Support
- [x] 2.1 Check `CANOPY_CONFIG` environment variable in config loader
- [x] 2.2 Implement priority order: flag > env > default

## 3. Testing
- [x] 3.1 Add tests for config flag override
- [x] 3.2 Add tests for environment variable override
- [x] 3.3 Add tests for priority order

## 4. Documentation
- [x] 4.1 Update `docs/configuration.md` with new options
- [x] 4.2 Update help text for root command

