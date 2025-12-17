package output

import (
	"bytes"
	"io"
	"os"
	"sync"
	"testing"
)

// stdoutMutex protects stdout capture operations across tests.
var stdoutMutex sync.Mutex

// stderrMutex protects stderr capture operations across tests.
var stderrMutex sync.Mutex

// captureOutput captures stdout output from a function.
// Uses a mutex to prevent parallel test interference.
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	stdoutMutex.Lock()
	defer stdoutMutex.Unlock()

	old := os.Stdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	// Ensure cleanup happens even if fn() panics
	defer func() { _ = r.Close() }()
	defer func() { os.Stdout = old }()

	os.Stdout = w

	fn()

	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to copy output: %v", err)
	}

	return buf.String()
}

// captureStderr captures stderr output from a function.
// Uses a mutex to prevent parallel test interference.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	stderrMutex.Lock()
	defer stderrMutex.Unlock()

	old := os.Stderr

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	// Ensure cleanup happens even if fn() panics
	defer func() { _ = r.Close() }()
	defer func() { os.Stderr = old }()

	os.Stderr = w

	fn()

	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to copy output: %v", err)
	}

	return buf.String()
}

func TestSuccess(t *testing.T) {
	tests := []struct {
		name   string
		action string
		target string
		want   string
	}{
		{
			name:   "basic success message",
			action: "Created workspace",
			target: "my-workspace",
			want:   "Created workspace my-workspace\n",
		},
		{
			name:   "with special characters",
			action: "Removed repository",
			target: "org/repo-name",
			want:   "Removed repository org/repo-name\n",
		},
		{
			name:   "empty target",
			action: "Done",
			target: "",
			want:   "Done \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureOutput(t, func() {
				Success(tt.action, tt.target)
			})
			if got != tt.want {
				t.Errorf("Success() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSuccessWithPath(t *testing.T) {
	tests := []struct {
		name   string
		action string
		target string
		path   string
		want   string
	}{
		{
			name:   "workspace created with path",
			action: "Created workspace",
			target: "my-workspace",
			path:   "/home/user/workspaces/my-workspace",
			want:   "Created workspace my-workspace in /home/user/workspaces/my-workspace\n",
		},
		{
			name:   "imported workspace",
			action: "Imported workspace",
			target: "imported-ws",
			path:   "/workspaces/imported-ws",
			want:   "Imported workspace imported-ws in /workspaces/imported-ws\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureOutput(t, func() {
				SuccessWithPath(tt.action, tt.target, tt.path)
			})
			if got != tt.want {
				t.Errorf("SuccessWithPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "simple info",
			message: "No orphaned worktrees found.",
			want:    "No orphaned worktrees found.\n",
		},
		{
			name:    "empty message",
			message: "",
			want:    "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureOutput(t, func() {
				Info(tt.message)
			})
			if got != tt.want {
				t.Errorf("Info() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInfof(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "formatted info",
			format: "Found %d orphaned worktree(s):",
			args:   []interface{}{5},
			want:   "Found 5 orphaned worktree(s):\n",
		},
		{
			name:   "multiple args",
			format: "Workspace: %s, Branch: %s",
			args:   []interface{}{"my-ws", "main"},
			want:   "Workspace: my-ws, Branch: main\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureOutput(t, func() {
				Infof(tt.format, tt.args...)
			})
			if got != tt.want {
				t.Errorf("Infof() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWarn(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "simple warning",
			message: "Configuration may be incomplete",
			want:    "Configuration may be incomplete\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureStderr(t, func() {
				Warn(tt.message)
			})
			if got != tt.want {
				t.Errorf("Warn() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWarnf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "formatted warning",
			format: "Missing %d files",
			args:   []interface{}{3},
			want:   "Missing 3 files\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureStderr(t, func() {
				Warnf(tt.format, tt.args...)
			})
			if got != tt.want {
				t.Errorf("Warnf() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "path without newline",
			message: "/home/user/workspace",
			want:    "/home/user/workspace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureOutput(t, func() {
				Print(tt.message)
			})
			if got != tt.want {
				t.Errorf("Print() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "formatted output",
			format: "%s/%s",
			args:   []interface{}{"/workspaces", "my-ws"},
			want:   "/workspaces/my-ws",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureOutput(t, func() {
				Printf(tt.format, tt.args...)
			})
			if got != tt.want {
				t.Errorf("Printf() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintln(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "path with newline",
			message: "/home/user/workspace",
			want:    "/home/user/workspace\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureOutput(t, func() {
				Println(tt.message)
			})
			if got != tt.want {
				t.Errorf("Println() = %q, want %q", got, tt.want)
			}
		})
	}
}
