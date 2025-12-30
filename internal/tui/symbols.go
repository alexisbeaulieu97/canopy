package tui

// Symbols provides configurable symbol mappings for TUI display.
// When Nerd Fonts are available, uses Nerd Font icons for rich display.
// Otherwise, falls back to ASCII for terminal compatibility.
type Symbols struct {
	useNerdFont bool
}

// NewSymbols creates a new Symbols with the specified Nerd Font mode.
// When useNerdFont is true, Nerd Font icons are used for richer display.
// When false, ASCII fallbacks are used for basic terminal compatibility.
func NewSymbols(useNerdFont bool) Symbols {
	return Symbols{useNerdFont: useNerdFont}
}

// Workspaces returns the workspaces header symbol ( or [W]).
func (s Symbols) Workspaces() string {
	if s.useNerdFont {
		return "" // nf-oct-stack
	}

	return "[W]"
}

// Disk returns the disk usage symbol ( or [D]).
func (s Symbols) Disk() string {
	if s.useNerdFont {
		return "" // nf-fa-database
	}

	return "[D]"
}

// Folder returns the folder/workspace detail symbol ( or [>]).
func (s Symbols) Folder() string {
	if s.useNerdFont {
		return "" // nf-oct-file_directory
	}

	return "[>]"
}

// Warning returns the warning symbol ( or !).
func (s Symbols) Warning() string {
	if s.useNerdFont {
		return "" // nf-fa-exclamation_triangle
	}

	return "!"
}

// Check returns the success/check symbol ( or ok).
func (s Symbols) Check() string {
	if s.useNerdFont {
		return "" // nf-fa-check
	}

	return "ok"
}

// Search returns the search symbol ( or [?]).
func (s Symbols) Search() string {
	if s.useNerdFont {
		return "" // nf-fa-search
	}

	return "[?]"
}

// Loading returns the loading symbol ( or ~).
func (s Symbols) Loading() string {
	if s.useNerdFont {
		return "" // nf-fa-clock_o
	}

	return "~"
}

// Repo returns the repository symbol ( or [R]).
func (s Symbols) Repo() string {
	if s.useNerdFont {
		return "" // nf-oct-repo
	}

	return "[R]"
}

// Branch returns the branch symbol ( or [B]).
func (s Symbols) Branch() string {
	if s.useNerdFont {
		return "" // nf-oct-git_branch
	}

	return "[B]"
}

// Dirty returns the dirty/modified symbol ( or *).
func (s Symbols) Dirty() string {
	if s.useNerdFont {
		return "" // nf-fa-pencil
	}

	return "*"
}

// Error returns the error symbol ( or X).
func (s Symbols) Error() string {
	if s.useNerdFont {
		return "" // nf-fa-times_circle
	}

	return "X"
}

// Unpushed returns the unpushed commits symbol ( or ^).
func (s Symbols) Unpushed() string {
	if s.useNerdFont {
		return "" // nf-fa-arrow_up
	}

	return "^"
}

// Behind returns the behind remote symbol ( or v).
func (s Symbols) Behind() string {
	if s.useNerdFont {
		return "" // nf-fa-arrow_down
	}

	return "v"
}

// Stale returns the stale/outdated symbol ( or ~).
func (s Symbols) Stale() string {
	if s.useNerdFont {
		return "" // nf-fa-clock_o
	}

	return "~"
}

// Time returns the time/clock symbol ( or @).
func (s Symbols) Time() string {
	if s.useNerdFont {
		return "" // nf-fa-clock_o
	}

	return "@"
}

// Cursor returns the cursor/selection indicator ( or >).
func (s Symbols) Cursor() string {
	if s.useNerdFont {
		return "" // nf-fa-chevron_right
	}

	return ">"
}

// NoCursor returns an empty cursor indicator (space).
func (s Symbols) NoCursor() string {
	return " "
}
