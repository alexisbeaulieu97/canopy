## ADDED Requirements

### Requirement: Configuration Path Override
The CLI SHALL support overriding the default configuration file path via flag or environment variable.

#### Scenario: Override config with flag
- **WHEN** user runs `canopy --config /path/to/config.yaml workspace list`
- **THEN** the configuration SHALL be loaded from `/path/to/config.yaml`
- **AND** the default config path SHALL be ignored

#### Scenario: Override config with environment variable
- **WHEN** `CANOPY_CONFIG=/path/to/config.yaml` is set
- **AND** user runs `canopy workspace list` without `--config` flag
- **THEN** the configuration SHALL be loaded from `/path/to/config.yaml`

#### Scenario: Flag takes precedence over environment variable
- **WHEN** `CANOPY_CONFIG=/env/config.yaml` is set
- **AND** user runs `canopy --config /flag/config.yaml workspace list`
- **THEN** the configuration SHALL be loaded from `/flag/config.yaml`
- **AND** the environment variable SHALL be ignored

#### Scenario: Default config when no override
- **WHEN** `--config` flag is not provided
- **AND** `CANOPY_CONFIG` environment variable is not set
- **THEN** the configuration SHALL be loaded from `~/.canopy/config.yaml`

#### Scenario: Config file not found error
- **WHEN** `--config /nonexistent/config.yaml` is specified
- **AND** the file does not exist
- **THEN** an error SHALL be returned indicating the config file was not found
- **AND** the error message SHALL include the attempted path

