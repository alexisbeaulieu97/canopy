// Package workspace manages workspace metadata and directories.
package workspace

import (
	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// MigrationFunc is a function that migrates a workspace from one version to the next.
// It receives the workspace and returns the migrated workspace or an error.
type MigrationFunc func(w *domain.Workspace) error

// migrationRegistry holds version-to-version migration functions.
// Key is the source version, value is the migration function to the next version.
var migrationRegistry = map[int]MigrationFunc{
	// Migration from version 0 to 1 is a no-op: just adds the version field
	// which is handled automatically by saveMetadata setting CurrentWorkspaceVersion
	0: migrateV0ToV1,
}

// migrateV0ToV1 migrates a version 0 workspace to version 1.
// This is a no-op migration since version 1 only adds the version field itself,
// which is set automatically on save.
func migrateV0ToV1(w *domain.Workspace) error {
	// No structural changes needed - the version field is set on save
	w.Version = 1
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
		migration, ok := migrationRegistry[w.Version]
		if !ok {
			// No migration path - this shouldn't happen with proper version management
			break
		}

		if err := migration(w); err != nil {
			return migrated, err
		}

		migrated = true
	}

	return migrated, nil
}

// NeedsMigration returns true if the workspace needs to be migrated to a newer version.
func NeedsMigration(w *domain.Workspace) bool {
	return w.Version < domain.CurrentWorkspaceVersion
}
