package tui

// Symbols provides configurable symbol mappings for TUI display.
// When emoji is disabled, ASCII fallbacks are used for terminal compatibility.
type Symbols struct {
	useEmoji bool
}

// NewSymbols creates a new Symbols with the specified emoji mode.
func NewSymbols(useEmoji bool) Symbols {
	return Symbols{useEmoji: useEmoji}
}

// Workspaces returns the workspaces header symbol (🌲 or [W]).
func (s Symbols) Workspaces() string {
	if s.useEmoji {
		return "🌲"
	}

	return "[W]"
}

// Disk returns the disk usage symbol (💾 or [D]).
func (s Symbols) Disk() string {
	if s.useEmoji {
		return "💾"
	}

	return "[D]"
}

// Folder returns the folder/workspace detail symbol (📂 or [>]).
func (s Symbols) Folder() string {
	if s.useEmoji {
		return "📂"
	}

	return "[>]"
}

// Warning returns the warning symbol (⚠ or [!]).
func (s Symbols) Warning() string {
	if s.useEmoji {
		return "⚠"
	}

	return "[!]"
}

// Check returns the success/check symbol (✓ or [*]).
func (s Symbols) Check() string {
	if s.useEmoji {
		return "✓"
	}

	return "[*]"
}

// Search returns the search symbol (🔍 or [?]).
func (s Symbols) Search() string {
	if s.useEmoji {
		return "🔍"
	}

	return "[?]"
}

// Loading returns the loading symbol (⏳ or [...]).
func (s Symbols) Loading() string {
	if s.useEmoji {
		return "⏳"
	}

	return "[...]"
}

// Repo returns the repository symbol (📁 or [-]).
func (s Symbols) Repo() string {
	if s.useEmoji {
		return "📁"
	}

	return "[-]"
}
