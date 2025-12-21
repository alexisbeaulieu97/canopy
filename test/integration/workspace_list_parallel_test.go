//go:build integration

package integration

import (
	"strings"
	"testing"
)

func TestWorkspaceListParallelMatchesSequential(t *testing.T) {
	setupConfig(t)

	tc := newTestContext(t)
	tc.createWorkspace("TEST-PAR-1")
	tc.createWorkspace("TEST-PAR-2")

	sequentialOut, err := runCanopy("workspace", "list", "--status", "--json", "--sequential-status")
	if err != nil {
		t.Fatalf("sequential list failed: %v\nOutput: %s", err, sequentialOut)
	}

	parallelOut, err := runCanopy("workspace", "list", "--status", "--json", "--parallel-status")
	if err != nil {
		t.Fatalf("parallel list failed: %v\nOutput: %s", err, parallelOut)
	}

	if strings.TrimSpace(sequentialOut) != strings.TrimSpace(parallelOut) {
		t.Fatalf("parallel output did not match sequential output\nSequential: %s\nParallel: %s", sequentialOut, parallelOut)
	}
}
