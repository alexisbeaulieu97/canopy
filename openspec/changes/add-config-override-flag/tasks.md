## 1. Flag Implementation
- [ ] 1.1 Add `--config` persistent flag to root command in `main.go`
- [ ] 1.2 Pass config path to App initialization
- [ ] 1.3 Update config loader to accept optional path parameter

## 2. Environment Variable Support
- [ ] 2.1 Check `CANOPY_CONFIG` environment variable in config loader
- [ ] 2.2 Implement priority order: flag > env > default

## 3. Testing
- [ ] 3.1 Add tests for config flag override
- [ ] 3.2 Add tests for environment variable override
- [ ] 3.3 Add tests for priority order

## 4. Documentation
- [ ] 4.1 Update `docs/configuration.md` with new options
- [ ] 4.2 Update help text for root command

