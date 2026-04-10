package config

import (
	"fmt"
	"strings"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

var knownConfigFields = []string{
	"projects_root",
	"workspaces_root",
	"closed_root",
	"workspace_close_default",
	"workspace_naming",
	"stale_threshold_days",
	"parallel_workers",
	"lock_timeout",
	"lock_stale_threshold",
	"defaults",
	"defaults.workspace_patterns",
	"templates",
	"templates.repos",
	"templates.default_branch",
	"templates.description",
	"templates.setup_commands",
	"hooks",
	"hooks.post_create",
	"hooks.pre_close",
	"tui",
	"tui.keybindings",
	"tui.use_emoji",
	"git",
	"git.retry",
	"git.retry.max_attempts",
	"git.retry.initial_delay",
	"git.retry.max_delay",
	"git.retry.multiplier",
	"git.retry.jitter_factor",
	"command",
	"description",
	"repos",
	"shell",
	"timeout",
	"continue_on_error",
	"quit",
	"search",
	"sync",
	"push",
	"close",
	"open_editor",
	"toggle_stale",
	"details",
	"select",
	"select_all",
	"deselect_all",
	"confirm",
	"cancel",
	"pattern",
}

func findSimilarField(unknown string) string {
	bestMatch := ""
	bestDistance := 4

	for _, known := range knownConfigFields {
		parts := strings.Split(known, ".")
		fieldName := parts[len(parts)-1]

		dist := levenshteinDistance(strings.ToLower(unknown), strings.ToLower(fieldName))
		if dist < bestDistance {
			bestDistance = dist
			bestMatch = fieldName
		}

		if len(parts) > 1 {
			dist = levenshteinDistance(strings.ToLower(unknown), strings.ToLower(known))
			if dist < bestDistance {
				bestDistance = dist
				bestMatch = known
			}
		}
	}

	return bestMatch
}

func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}

	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func formatUnknownFieldError(unknownFields []string) string {
	var msgs []string

	for _, field := range unknownFields {
		similar := findSimilarField(field)
		if similar != "" {
			msgs = append(msgs, fmt.Sprintf("unknown config field %q, did you mean %q?", field, similar))
			continue
		}

		msgs = append(msgs, fmt.Sprintf("unknown config field %q", field))
	}

	return strings.Join(msgs, "; ")
}

func extractUnknownFields(errMsg string) []string {
	idx := strings.Index(errMsg, "invalid keys:")
	if idx == -1 {
		return nil
	}

	keysStr := strings.TrimSpace(errMsg[idx+len("invalid keys:"):])

	var fields []string
	for _, field := range strings.Split(keysStr, ",") {
		field = strings.TrimSpace(field)
		if field != "" {
			fields = append(fields, field)
		}
	}

	return fields
}

func handleUnmarshalError(err error) error {
	errMsg := err.Error()
	if strings.Contains(errMsg, "has invalid keys") {
		if unknownFields := extractUnknownFields(errMsg); len(unknownFields) > 0 {
			return cerrors.NewConfigValidation("config", formatUnknownFieldError(unknownFields))
		}
	}

	return cerrors.NewConfigInvalid(fmt.Sprintf("failed to unmarshal: %v", err))
}
