package workspaces

import (
	"context"
	"errors"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// joinErrors joins multiple errors into a single error.
// Returns nil if no errors are provided or all are nil.
func joinErrors(errs ...error) error {
	return errors.Join(errs...)
}

// isWorkspaceNotFound checks if the error is a workspace not found error.
func isWorkspaceNotFound(err error) bool {
	return errors.Is(err, cerrors.WorkspaceNotFound)
}

// isCanopyError checks if the error is a CanopyError and assigns it to the target.
func isCanopyError(err error, target **cerrors.CanopyError) bool {
	return errors.As(err, target)
}

// isDeadlineExceeded checks if the error is a context deadline exceeded error.
func isDeadlineExceeded(err error) bool {
	return errors.Is(err, context.DeadlineExceeded)
}
