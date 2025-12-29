// Package output provides helpers for CLI output formatting.
package output

// Icons provides consistent status icons for CLI output.
// When Unicode is disabled (e.g., NO_COLOR=1 or piped output),
// ASCII fallbacks are used for terminal compatibility.
type Icons struct {
	useUnicode bool
}

// NewIcons creates a new Icons instance based on color/terminal capabilities.
func NewIcons() Icons {
	return Icons{useUnicode: ColorEnabled() && UnicodeEnabled()}
}

// NewIconsWithMode creates Icons with explicit Unicode mode.
func NewIconsWithMode(useUnicode bool) Icons {
	return Icons{useUnicode: useUnicode}
}

// Success returns the success icon (✓ or [ok]).
func (i Icons) Success() string {
	if i.useUnicode {
		return "✓"
	}

	return "[ok]"
}

// Warning returns the warning icon (⚠ or [!]).
func (i Icons) Warning() string {
	if i.useUnicode {
		return "⚠"
	}

	return "[!]"
}

// Error returns the error icon (✗ or [X]).
func (i Icons) Error() string {
	if i.useUnicode {
		return "✗"
	}

	return "[X]"
}

// Info returns the info icon (ℹ or [i]).
func (i Icons) Info() string {
	if i.useUnicode {
		return "ℹ"
	}

	return "[i]"
}

// Dirty returns the dirty/modified icon (● or *).
func (i Icons) Dirty() string {
	if i.useUnicode {
		return "●"
	}

	return "*"
}

// Clean returns the clean icon (○ or -).
func (i Icons) Clean() string {
	if i.useUnicode {
		return "○"
	}

	return "-"
}

// Unpushed returns the unpushed commits icon (↑ or ^).
func (i Icons) Unpushed() string {
	if i.useUnicode {
		return "↑"
	}

	return "^"
}

// Behind returns the behind remote icon (↓ or v).
func (i Icons) Behind() string {
	if i.useUnicode {
		return "↓"
	}

	return "v"
}

// Bullet returns a list bullet (• or *).
func (i Icons) Bullet() string {
	if i.useUnicode {
		return "•"
	}

	return "*"
}

// SpinnerFrames returns the frames for an animated spinner.
func (i Icons) SpinnerFrames() []string {
	if i.useUnicode {
		return []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	}

	return []string{".", "..", "..."}
}

// UseUnicode returns whether Unicode icons are enabled.
func (i Icons) UseUnicode() bool {
	return i.useUnicode
}

// StatusIcon returns the appropriate icon for a clean/dirty status.
func (i Icons) StatusIcon(isDirty bool) string {
	if isDirty {
		return i.Dirty()
	}

	return i.Clean()
}

// ResultIcon returns the appropriate icon for success/failure.
func (i Icons) ResultIcon(success bool) string {
	if success {
		return i.Success()
	}

	return i.Error()
}
