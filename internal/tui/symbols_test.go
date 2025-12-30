package tui

import "testing"

func TestSymbols_NerdFontMode(t *testing.T) {
	symbols := NewSymbols(true)

	tests := []struct {
		name   string
		method func() string
		want   string
	}{
		{name: "Workspaces", method: symbols.Workspaces, want: ""},
		{name: "Disk", method: symbols.Disk, want: ""},
		{name: "Folder", method: symbols.Folder, want: ""},
		{name: "Warning", method: symbols.Warning, want: ""},
		{name: "Check", method: symbols.Check, want: ""},
		{name: "Search", method: symbols.Search, want: ""},
		{name: "Loading", method: symbols.Loading, want: ""},
		{name: "Repo", method: symbols.Repo, want: ""},
		{name: "Branch", method: symbols.Branch, want: ""},
		{name: "Dirty", method: symbols.Dirty, want: ""},
		{name: "Error", method: symbols.Error, want: ""},
		{name: "Unpushed", method: symbols.Unpushed, want: ""},
		{name: "Behind", method: symbols.Behind, want: ""},
		{name: "Stale", method: symbols.Stale, want: ""},
		{name: "Time", method: symbols.Time, want: ""},
		{name: "Cursor", method: symbols.Cursor, want: ""},
		{name: "NoCursor", method: symbols.NoCursor, want: " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method(); got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestSymbols_ASCIIMode(t *testing.T) {
	symbols := NewSymbols(false)

	tests := []struct {
		name   string
		method func() string
		want   string
	}{
		{name: "Workspaces", method: symbols.Workspaces, want: "[W]"},
		{name: "Disk", method: symbols.Disk, want: "[D]"},
		{name: "Folder", method: symbols.Folder, want: "[>]"},
		{name: "Warning", method: symbols.Warning, want: "!"},
		{name: "Check", method: symbols.Check, want: "ok"},
		{name: "Search", method: symbols.Search, want: "[?]"},
		{name: "Loading", method: symbols.Loading, want: "~"},
		{name: "Repo", method: symbols.Repo, want: "[R]"},
		{name: "Branch", method: symbols.Branch, want: "[B]"},
		{name: "Dirty", method: symbols.Dirty, want: "*"},
		{name: "Error", method: symbols.Error, want: "X"},
		{name: "Unpushed", method: symbols.Unpushed, want: "^"},
		{name: "Behind", method: symbols.Behind, want: "v"},
		{name: "Stale", method: symbols.Stale, want: "~"},
		{name: "Time", method: symbols.Time, want: "@"},
		{name: "Cursor", method: symbols.Cursor, want: ">"},
		{name: "NoCursor", method: symbols.NoCursor, want: " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method(); got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
