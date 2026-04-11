package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

func parseTags(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")

	var tags []string

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tags = append(tags, p)
		}
	}

	return tags
}

func registerWithPrompt(cmd *cobra.Command, registry ports.RepoRegistry, alias string, entry ports.RegistryEntry, logger rollbackLogger) (string, error) {
	if cmd == nil {
		return "", cerrors.NewInvalidArgument("cmd", "command is required")
	}

	if registry == nil {
		return alias, cerrors.NewConfigInvalid("registry not configured")
	}

	target := strings.TrimSpace(alias)
	if target == "" {
		return "", cerrors.NewInvalidArgument("alias", "alias is required")
	}

	for {
		if _, exists := registry.Resolve(target); !exists {
			return registerAlias(registry, target, entry, logger)
		}

		suggested := nextAvailableAlias(registry, target)

		var err error

		target, err = promptAlias(cmd, target, suggested)
		if err != nil {
			return "", err
		}
	}
}

func nextAvailableAlias(registry ports.RepoRegistry, base string) string {
	target := base
	for idx := 2; ; idx++ {
		if _, exists := registry.Resolve(target); !exists {
			return target
		}

		target = fmt.Sprintf("%s-%d", base, idx)
	}
}

func registerAlias(registry ports.RepoRegistry, alias string, entry ports.RegistryEntry, logger rollbackLogger) (string, error) {
	if err := registry.Register(alias, entry, false); err != nil {
		return "", err
	}

	rollbackFn := func() error {
		return registry.Unregister(alias)
	}
	if err := saveRegistryWithRollback(registry, rollbackFn, "registration", logger); err != nil {
		return "", err
	}

	return alias, nil
}

// rollbackLogger is an interface for logging rollback errors.
type rollbackLogger interface {
	Errorf(format string, args ...interface{})
}

// saveRegistryWithRollback saves the registry and performs a rollback on failure.
// It logs any errors that occur during rollback and returns the save error if present.
// If logger is nil, rollback errors are silently discarded.
func saveRegistryWithRollback(
	registry ports.RepoRegistry,
	rollbackFn func() error,
	rollbackDesc string,
	logger rollbackLogger,
) error {
	if err := registry.Save(); err != nil {
		if rollbackErr := rollbackFn(); rollbackErr != nil {
			if logger != nil {
				logger.Errorf("Failed to rollback %s: %v", rollbackDesc, rollbackErr)
			}
		} else if rollbackSaveErr := registry.Save(); rollbackSaveErr != nil {
			if logger != nil {
				logger.Errorf("Failed to save rollback: %v", rollbackSaveErr)
			}
		}

		return cerrors.NewRegistryError("save", "failed to save registry", err)
	}

	return nil
}

func promptAlias(cmd *cobra.Command, alias, suggested string) (string, error) {
	reader := bufio.NewReader(cmd.InOrStdin())
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Alias '%s' already exists. Enter a new alias or press Enter to use '%s': ", alias, suggested); err != nil {
		return "", err
	}

	input, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return suggested, nil
	}

	return input, nil
}
