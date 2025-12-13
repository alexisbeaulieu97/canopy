// Package testutil provides shared test helpers for Canopy tests.
//
// This package contains commonly used test utilities extracted from individual
// test files to eliminate code duplication and ensure consistent test setup
// across the codebase.
//
// # Git Helpers
//
// The git helpers provide functions for creating and manipulating git repositories
// in tests:
//
//	testutil.CreateRepoWithCommit(t, "/tmp/repo")  // Initialize repo with commit
//	testutil.RunGit(t, "/tmp/repo", "status")      // Run git command
//	output := testutil.RunGitOutput(t, dir, "log") // Get git output
//
// # Filesystem Helpers
//
// The filesystem helpers provide functions for common file operations:
//
//	testutil.MustMkdir(t, "/tmp/dir")           // Create directory
//	testutil.MustWriteFile(t, "/tmp/f", "data") // Write file
//	content := testutil.MustReadFile(t, path)   // Read file
//
// # Usage Guidelines
//
// All functions in this package call t.Helper() to ensure proper test failure
// attribution. Functions prefixed with "Must" will fail the test immediately
// on error using t.Fatalf().
package testutil
