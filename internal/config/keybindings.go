package config

import (
	"fmt"
	"sort"
	"strings"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// Default keybindings for TUI actions.
var (
	DefaultQuitKeys        = []string{"q", "ctrl+c"}
	DefaultSearchKeys      = []string{"/"}
	DefaultSyncKeys        = []string{"s"}
	DefaultPushKeys        = []string{"p"}
	DefaultCloseKeys       = []string{"c"}
	DefaultOpenEditorKeys  = []string{"o"}
	DefaultToggleStaleKeys = []string{"t"}
	DefaultDetailsKeys     = []string{"enter"}
	DefaultSelectKeys      = []string{"space"}
	DefaultSelectAllKeys   = []string{"a"}
	DefaultDeselectAllKeys = []string{"A"}
	DefaultConfirmKeys     = []string{"y", "Y"}
	DefaultCancelKeys      = []string{"n", "N", "esc"}
)

func copyKeys(keys []string) []string {
	if keys == nil {
		return nil
	}

	result := make([]string, len(keys))
	copy(result, keys)

	return result
}

func applyDefaultKeys(keys *[]string, defaults []string) {
	if len(*keys) == 0 {
		*keys = copyKeys(defaults)
	}
}

// WithDefaults returns a copy of Keybindings with defaults applied for empty fields.
func (k Keybindings) WithDefaults() Keybindings {
	result := Keybindings{
		Quit:        copyKeys(k.Quit),
		Search:      copyKeys(k.Search),
		Sync:        copyKeys(k.Sync),
		Push:        copyKeys(k.Push),
		Close:       copyKeys(k.Close),
		OpenEditor:  copyKeys(k.OpenEditor),
		ToggleStale: copyKeys(k.ToggleStale),
		Details:     copyKeys(k.Details),
		Select:      copyKeys(k.Select),
		SelectAll:   copyKeys(k.SelectAll),
		DeselectAll: copyKeys(k.DeselectAll),
		Confirm:     copyKeys(k.Confirm),
		Cancel:      copyKeys(k.Cancel),
	}

	applyDefaultKeys(&result.Quit, DefaultQuitKeys)
	applyDefaultKeys(&result.Search, DefaultSearchKeys)
	applyDefaultKeys(&result.Sync, DefaultSyncKeys)
	applyDefaultKeys(&result.Push, DefaultPushKeys)
	applyDefaultKeys(&result.Close, DefaultCloseKeys)
	applyDefaultKeys(&result.OpenEditor, DefaultOpenEditorKeys)
	applyDefaultKeys(&result.ToggleStale, DefaultToggleStaleKeys)
	applyDefaultKeys(&result.Details, DefaultDetailsKeys)
	applyDefaultKeys(&result.Select, DefaultSelectKeys)
	applyDefaultKeys(&result.SelectAll, DefaultSelectAllKeys)
	applyDefaultKeys(&result.DeselectAll, DefaultDeselectAllKeys)
	applyDefaultKeys(&result.Confirm, DefaultConfirmKeys)
	applyDefaultKeys(&result.Cancel, DefaultCancelKeys)

	return result
}

var validKeys = map[string]bool{
	"a": true, "b": true, "c": true, "d": true, "e": true, "f": true, "g": true, "h": true,
	"i": true, "j": true, "k": true, "l": true, "m": true, "n": true, "o": true, "p": true,
	"q": true, "r": true, "s": true, "t": true, "u": true, "v": true, "w": true, "x": true,
	"y": true, "z": true,
	"A": true, "B": true, "C": true, "D": true, "E": true, "F": true, "G": true, "H": true,
	"I": true, "J": true, "K": true, "L": true, "M": true, "N": true, "O": true, "P": true,
	"Q": true, "R": true, "S": true, "T": true, "U": true, "V": true, "W": true, "X": true,
	"Y": true, "Z": true,
	"0": true, "1": true, "2": true, "3": true, "4": true,
	"5": true, "6": true, "7": true, "8": true, "9": true,
	"enter": true, "esc": true, "tab": true, "backspace": true, "delete": true,
	"up": true, "down": true, "left": true, "right": true,
	"home": true, "end": true, "pgup": true, "pgdown": true,
	"space": true,
	"f1":    true, "f2": true, "f3": true, "f4": true, "f5": true, "f6": true,
	"f7": true, "f8": true, "f9": true, "f10": true, "f11": true, "f12": true,
	"/": true, "\\": true, ".": true, ",": true, ";": true, "'": true, "`": true,
	"[": true, "]": true, "-": true, "=": true,
}

func isValidKey(key string) bool {
	if key == "" {
		return false
	}

	for _, prefix := range []string{"ctrl+", "alt+", "shift+"} {
		if strings.HasPrefix(key, prefix) {
			return isValidKey(strings.TrimPrefix(key, prefix))
		}
	}

	return validKeys[key]
}

// ValidateKeybindings checks for invalid and conflicting keybindings.
func (k Keybindings) ValidateKeybindings() error {
	var errors []string

	validateKeys := func(keys []string, action string) {
		for _, key := range keys {
			if !isValidKey(key) {
				errors = append(errors, fmt.Sprintf("invalid key %q for action %q", key, action))
			}
		}
	}

	validateKeys(k.Quit, "quit")
	validateKeys(k.Search, "search")
	validateKeys(k.Sync, "sync")
	validateKeys(k.Push, "push")
	validateKeys(k.Close, "close")
	validateKeys(k.OpenEditor, "open_editor")
	validateKeys(k.ToggleStale, "toggle_stale")
	validateKeys(k.Details, "details")
	validateKeys(k.Select, "select")
	validateKeys(k.SelectAll, "select_all")
	validateKeys(k.DeselectAll, "deselect_all")
	validateKeys(k.Confirm, "confirm")
	validateKeys(k.Cancel, "cancel")

	keyUsage := make(map[string][]string)
	addKeys := func(keys []string, action string) {
		seen := make(map[string]struct{}, len(keys))
		for _, key := range keys {
			if _, ok := seen[key]; ok {
				continue
			}

			seen[key] = struct{}{}
			keyUsage[key] = append(keyUsage[key], action)
		}
	}

	addKeys(k.Quit, "quit")
	addKeys(k.Search, "search")
	addKeys(k.Sync, "sync")
	addKeys(k.Push, "push")
	addKeys(k.Close, "close")
	addKeys(k.OpenEditor, "open_editor")
	addKeys(k.ToggleStale, "toggle_stale")
	addKeys(k.Details, "details")
	addKeys(k.Select, "select")
	addKeys(k.SelectAll, "select_all")
	addKeys(k.DeselectAll, "deselect_all")
	addKeys(k.Confirm, "confirm")
	addKeys(k.Cancel, "cancel")

	var conflictKeys []string

	for key, actions := range keyUsage {
		if len(actions) > 1 {
			conflictKeys = append(conflictKeys, key)
		}
	}

	sort.Strings(conflictKeys)

	for _, key := range conflictKeys {
		sort.Strings(keyUsage[key])
		errors = append(errors, fmt.Sprintf("key %q is assigned to multiple actions: %s", key, strings.Join(keyUsage[key], ", ")))
	}

	if len(errors) > 0 {
		return cerrors.NewConfigValidation("tui.keybindings", strings.Join(errors, "; "))
	}

	return nil
}
