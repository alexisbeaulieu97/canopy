package domain

import (
	"testing"
	"time"
)

func TestWorkspaceIsStale(t *testing.T) {
	base := time.Now()

	tests := []struct {
		name      string
		ws        Workspace
		threshold int
		stale     bool
	}{
		{
			name:      "zero threshold never stale",
			ws:        Workspace{LastModified: base.AddDate(0, 0, -30)},
			threshold: 0,
			stale:     false,
		},
		{
			name:      "zero last modified",
			ws:        Workspace{},
			threshold: 14,
			stale:     false,
		},
		{
			name:      "older than threshold",
			ws:        Workspace{LastModified: base.AddDate(0, 0, -15)},
			threshold: 14,
			stale:     true,
		},
		{
			name:      "just under threshold",
			ws:        Workspace{LastModified: base.AddDate(0, 0, -14).Add(time.Hour)},
			threshold: 14,
			stale:     false,
		},
		{
			name:      "newer than threshold",
			ws:        Workspace{LastModified: base.AddDate(0, 0, -5)},
			threshold: 14,
			stale:     false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ws.IsStale(tt.threshold); got != tt.stale {
				t.Fatalf("IsStale() = %v, want %v", got, tt.stale)
			}
		})
	}
}
