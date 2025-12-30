package tui

import "testing"

func TestSymbols_NerdFontMode(t *testing.T) {
	symbols := NewSymbols(true)

	tests := []struct {
		name   string
		method func() string
		want   string
	}{
		{name: "Workspaces", method: symbols.Workspaces, want: "\uf0c9"},
		{name: "Disk", method: symbols.Disk, want: "\uf1c0"},
		{name: "Folder", method: symbols.Folder, want: "\uf07b"},
		{name: "Warning", method: symbols.Warning, want: "\uf071"},
		{name: "Check", method: symbols.Check, want: "\uf00c"},
		{name: "Search", method: symbols.Search, want: "\uf002"},
		{name: "Loading", method: symbols.Loading, want: "\uf110"},
		{name: "Repo", method: symbols.Repo, want: "\uf1d3"},
		{name: "Branch", method: symbols.Branch, want: "\ue725"},
		{name: "Dirty", method: symbols.Dirty, want: "\uf040"},
		{name: "Error", method: symbols.Error, want: "\uf057"},
		{name: "Unpushed", method: symbols.Unpushed, want: "\uf062"},
		{name: "Behind", method: symbols.Behind, want: "\uf063"},
		{name: "Stale", method: symbols.Stale, want: "\uf017"},
		{name: "Time", method: symbols.Time, want: "\uf017"},
		{name: "Cursor", method: symbols.Cursor, want: "\uf054"},
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
		{name: "Loading", method: symbols.Loading, want: "..."},
		{name: "Repo", method: symbols.Repo, want: "[R]"},
		{name: "Branch", method: symbols.Branch, want: "[B]"},
		{name: "Dirty", method: symbols.Dirty, want: "*"},
		{name: "Error", method: symbols.Error, want: "X"},
		{name: "Unpushed", method: symbols.Unpushed, want: "^"},
		{name: "Behind", method: symbols.Behind, want: "v"},
		{name: "Stale", method: symbols.Stale, want: "o"},
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
