package components

import (
	"strings"
	"testing"
)

func TestNewStatusBadge(t *testing.T) {
	tests := []struct {
		name     string
		input    StatusBadgeInput
		expected HealthStatus
	}{
		{
			name: "error state takes precedence",
			input: StatusBadgeInput{
				HasError: true,
				IsLoaded: true,
			},
			expected: HealthError,
		},
		{
			name: "loading when not loaded",
			input: StatusBadgeInput{
				HasError: false,
				IsLoaded: false,
			},
			expected: HealthLoading,
		},
		{
			name: "orphaned when has orphans",
			input: StatusBadgeInput{
				HasError:    false,
				IsLoaded:    true,
				OrphanCount: 2,
			},
			expected: HealthOrphaned,
		},
		{
			name: "dirty when has dirty repos",
			input: StatusBadgeInput{
				HasError:   false,
				IsLoaded:   true,
				DirtyRepos: 1,
			},
			expected: HealthDirty,
		},
		{
			name: "dirty when has unpushed repos",
			input: StatusBadgeInput{
				HasError:      false,
				IsLoaded:      true,
				UnpushedRepos: 1,
			},
			expected: HealthDirty,
		},
		{
			name: "attention when stale",
			input: StatusBadgeInput{
				HasError: false,
				IsLoaded: true,
				IsStale:  true,
			},
			expected: HealthAttention,
		},
		{
			name: "attention when behind",
			input: StatusBadgeInput{
				HasError:    false,
				IsLoaded:    true,
				BehindRepos: 1,
			},
			expected: HealthAttention,
		},
		{
			name: "clean when everything is good",
			input: StatusBadgeInput{
				HasError: false,
				IsLoaded: true,
			},
			expected: HealthClean,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badge := NewStatusBadge(tt.input)
			if badge.Status != tt.expected {
				t.Errorf("expected status %s, got %s", tt.expected, badge.Status)
			}
		})
	}
}

func TestStatusBadgeRender(t *testing.T) {
	badge := NewStatusBadge(StatusBadgeInput{
		IsLoaded: true,
		HasError: false,
	})

	result := badge.Render()
	if result == "" {
		t.Error("expected non-empty render result")
	}
}

func TestNewBadgeSet(t *testing.T) {
	tests := []struct {
		name          string
		input         BadgeSetInput
		expectEmpty   bool
		expectContain string
	}{
		{
			name: "empty when not loaded",
			input: BadgeSetInput{
				IsLoaded: false,
				HasError: false,
			},
			expectEmpty: true,
		},
		{
			name: "has ERROR badge on error",
			input: BadgeSetInput{
				IsLoaded: true,
				HasError: true,
			},
			expectContain: "ERROR",
		},
		{
			name: "has dirty badge",
			input: BadgeSetInput{
				IsLoaded:   true,
				DirtyRepos: 2,
			},
			expectContain: "dirty",
		},
		{
			name: "has unpushed badge",
			input: BadgeSetInput{
				IsLoaded:      true,
				UnpushedRepos: 3,
			},
			expectContain: "unpushed",
		},
		{
			name: "has behind badge",
			input: BadgeSetInput{
				IsLoaded:    true,
				BehindRepos: 1,
			},
			expectContain: "behind",
		},
		{
			name: "has STALE badge",
			input: BadgeSetInput{
				IsLoaded: true,
				IsStale:  true,
			},
			expectContain: "STALE",
		},
		{
			name: "has orphan badge",
			input: BadgeSetInput{
				IsLoaded:    true,
				OrphanCount: 1,
			},
			expectContain: "orphan",
		},
		{
			name: "has orphan check failed badge",
			input: BadgeSetInput{
				IsLoaded:          true,
				OrphanCheckFailed: true,
			},
			expectContain: "orphan check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := NewBadgeSet(tt.input)
			result := bs.Render()

			if tt.expectEmpty && result != "" {
				t.Errorf("expected empty result, got %s", result)
			}

			if tt.expectContain != "" && !strings.Contains(result, tt.expectContain) {
				t.Errorf("expected result to contain %q, got %s", tt.expectContain, result)
			}
		})
	}
}

func TestNewStatusLine(t *testing.T) {
	tests := []struct {
		name          string
		input         StatusLineInput
		expectContain string
	}{
		{
			name:          "clean when all zero",
			input:         StatusLineInput{},
			expectContain: "All repos clean",
		},
		{
			name: "shows dirty count",
			input: StatusLineInput{
				DirtyRepos: 2,
			},
			expectContain: "dirty",
		},
		{
			name: "shows unpushed count",
			input: StatusLineInput{
				UnpushedRepos: 3,
			},
			expectContain: "unpushed",
		},
		{
			name: "shows behind count",
			input: StatusLineInput{
				BehindRepos: 1,
			},
			expectContain: "behind",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl := NewStatusLine(tt.input)
			result := sl.Render()

			if !strings.Contains(result, tt.expectContain) {
				t.Errorf("expected result to contain %q, got %s", tt.expectContain, result)
			}
		})
	}
}
