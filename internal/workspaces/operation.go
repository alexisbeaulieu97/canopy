// Package workspaces contains workspace-level business logic.
package workspaces

import (
	"fmt"

	"github.com/alexisbeaulieu97/canopy/internal/logging"
)

type operationStep struct {
	action   func() error
	rollback func() error
}

// Operation runs a sequence of steps and rolls back completed steps on failure.
type Operation struct {
	steps  []operationStep
	logger *logging.Logger
}

// NewOperation creates a new operation helper.
func NewOperation(logger *logging.Logger) *Operation {
	return &Operation{logger: logger}
}

// AddStep adds an action with an optional rollback.
func (o *Operation) AddStep(action, rollback func() error) {
	o.steps = append(o.steps, operationStep{action: action, rollback: rollback})
}

// Execute runs steps in order and performs rollbacks on failure.
func (o *Operation) Execute() error {
	for idx, step := range o.steps {
		if step.action == nil {
			return fmt.Errorf("operation step %d has no action", idx)
		}

		if err := step.action(); err != nil {
			o.rollback(idx, err)
			return err
		}
	}

	return nil
}

func (o *Operation) rollback(failedIndex int, originalErr error) {
	for i := failedIndex; i >= 0; i-- {
		step := o.steps[i]
		if step.rollback == nil {
			continue
		}

		rollbackErr := step.rollback()

		if o.logger == nil {
			continue
		}

		status := "ok"
		if rollbackErr != nil {
			status = "failed"
		}

		o.logger.Debug("operation rollback",
			"step", i,
			"rollback_status", status,
			"original_error", originalErr,
			"rollback_error", rollbackErr,
		)
	}
}
