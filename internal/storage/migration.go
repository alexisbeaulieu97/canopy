// Package storage manages workspace metadata and directories.
package storage

import (
	"fmt"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// MigrationFunc is a function that migrates a workspace from one version to the next.
// It receives the workspace and performs structural changes only.
// The migration loop handles version advancement after successful migration.
type MigrationFunc func(w *domain.Workspace) error

// migrationRegistry holds version-to-version migration functions.
// Key is the source version, value is the migration function to the next version.
var migrationRegistry = map[int]MigrationFunc{
	// Migration from version 0 to 1 is a no-op: just adds the version field
	// which is handled automatically by saveMetadata setting CurrentWorkspaceVersion
	0: migrateV0ToV1,
}

// migrateV0ToV1 migrates a version 0 workspace to version 1.
// This is a no-op migration since version 1 only adds the version field itself.
// Note: Version increment is handled by the migration loop, not here.
func migrateV0ToV1(_ *domain.Workspace) error {
	// No structural changes needed for v0 -> v1
	// The version field is incremented by MigrateWorkspace after this returns
	return nil
}

// MigrateWorkspace applies all necessary migrations to bring a workspace
// from its current version to the current schema version.
// Returns true if any migrations were applied.
func MigrateWorkspace(w *domain.Workspace) (bool, error) {
	if w.Version >= domain.CurrentWorkspaceVersion {
		return false, nil
	}

	migrated := false

	for w.Version < domain.CurrentWorkspaceVersion {
		previousVersion := w.Version

		migration, ok := migrationRegistry[w.Version]
		if !ok {
			return migrated, fmt.Errorf("missing migration path from version %d to %d",
				w.Version, domain.CurrentWorkspaceVersion)
		}

		if err := migration(w); err != nil {
			return migrated, fmt.Errorf("migration from version %d failed: %w", previousVersion, err)
		}

		// Advance version after successful migration
		w.Version = previousVersion + 1
		migrated = true
	}

	return migrated, nil
}

// NeedsMigration returns true if the workspace needs to be migrated to a newer version.
func NeedsMigration(w *domain.Workspace) bool {
	return w.Version < domain.CurrentWorkspaceVersion
}
