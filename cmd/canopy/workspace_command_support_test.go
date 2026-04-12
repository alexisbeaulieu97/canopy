package main

import (
	"bufio"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestValidateWorkspaceTargetArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		pattern   string
		all       bool
		wantError bool
	}{
		{
			name: "single workspace target",
			args: []string{"WS-1"},
		},
		{
			name:    "bulk target with pattern",
			pattern: "^WS-",
		},
		{
			name:      "bulk target rejects workspace id",
			args:      []string{"WS-1"},
			pattern:   "^WS-",
			wantError: true,
		},
		{
			name:      "all rejects explicit pattern",
			pattern:   "^WS-",
			all:       true,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := newWorkspaceTargetTestCommand()

			_ = cmd.Flags().Set("pattern", tt.pattern)
			if tt.all {
				_ = cmd.Flags().Set("all", "true")
			}

			err := validateWorkspaceTargetArgs(cmd, tt.args, 1, "id", "workspace ID is required", 0, "id", "cannot provide workspace ID with --pattern or --all")
			if tt.wantError && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantError && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestValidateHookExecutionFlags(t *testing.T) {
	t.Parallel()

	if err := validateHookExecutionFlags(false, false, false); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := validateHookExecutionFlags(true, true, false); err == nil {
		t.Fatal("expected hooks-only/no-hooks conflict")
	}

	if err := validateHookExecutionFlags(true, false, true); err == nil {
		t.Fatal("expected dry-run-hooks/no-hooks conflict")
	}

	if err := validateHookExecutionFlags(false, true, true); err == nil {
		t.Fatal("expected dry-run-hooks/hooks-only conflict")
	}
}

func TestResolveWorkspaceCloseKeepMetadata(t *testing.T) {
	t.Parallel()

	if !resolveWorkspaceCloseKeepMetadata(false, true, false, false, false, nil) {
		t.Fatal("expected explicit keep flag to win")
	}

	if resolveWorkspaceCloseKeepMetadata(true, false, true, false, false, nil) {
		t.Fatal("expected explicit delete flag to win")
	}

	if !resolveWorkspaceCloseKeepMetadata(true, false, false, false, true, bufio.NewReader(strings.NewReader("y\n"))) {
		t.Fatal("expected yes answer to keep metadata")
	}

	if resolveWorkspaceCloseKeepMetadata(true, false, false, false, true, bufio.NewReader(strings.NewReader("n\n"))) {
		t.Fatal("expected no answer to delete metadata")
	}

	if !resolveWorkspaceCloseKeepMetadata(true, false, false, false, true, bufio.NewReader(strings.NewReader("maybe\n"))) {
		t.Fatal("expected invalid answer to fall back to default")
	}

	if !resolveWorkspaceCloseKeepMetadata(true, false, false, true, true, bufio.NewReader(strings.NewReader("n\n"))) {
		t.Fatal("expected dry-run to avoid prompting and keep the default choice")
	}
}

func newWorkspaceTargetTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("pattern", "", "pattern")
	cmd.Flags().Bool("all", false, "all")

	return cmd
}
