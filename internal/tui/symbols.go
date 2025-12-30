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

// Workspaces returns the workspaces header symbol.
func (s Symbols) Workspaces() string {
	if s.useNerdFont {
		return "\uf0c9" // nf-fa-bars (stack icon)
	}

	return "[W]"
}

// Disk returns the disk usage symbol.
func (s Symbols) Disk() string {
	if s.useNerdFont {
		return "\uf1c0" // nf-fa-database
	}

	return "[D]"
}

// Folder returns the folder/workspace detail symbol.
func (s Symbols) Folder() string {
	if s.useNerdFont {
		return "\uf07b" // nf-fa-folder
	}

	return "[>]"
}

// Warning returns the warning symbol.
func (s Symbols) Warning() string {
	if s.useNerdFont {
		return "\uf071" // nf-fa-exclamation_triangle
	}

	return "!"
}

// Check returns the success/check symbol.
func (s Symbols) Check() string {
	if s.useNerdFont {
		return "\uf00c" // nf-fa-check
	}

	return "ok"
}

// Search returns the search symbol.
func (s Symbols) Search() string {
	if s.useNerdFont {
		return "\uf002" // nf-fa-search
	}

	return "[?]"
}

// Loading returns the loading symbol.
func (s Symbols) Loading() string {
	if s.useNerdFont {
		return "\uf110" // nf-fa-spinner
	}

	return "..."
}

// Repo returns the repository symbol.
func (s Symbols) Repo() string {
	if s.useNerdFont {
		return "\uf1d3" // nf-fa-git-square
	}

	return "[R]"
}

// Branch returns the branch symbol.
func (s Symbols) Branch() string {
	if s.useNerdFont {
		return "\ue725" // nf-dev-git_branch
	}

	return "[B]"
}

// Dirty returns the dirty/modified symbol.
func (s Symbols) Dirty() string {
	if s.useNerdFont {
		return "\uf040" // nf-fa-pencil
	}

	return "*"
}

// Error returns the error symbol.
func (s Symbols) Error() string {
	if s.useNerdFont {
		return "\uf057" // nf-fa-times_circle
	}

	return "X"
}

// Unpushed returns the unpushed commits symbol.
func (s Symbols) Unpushed() string {
	if s.useNerdFont {
		return "\uf062" // nf-fa-arrow_up
	}

	return "^"
}

// Behind returns the behind remote symbol.
func (s Symbols) Behind() string {
	if s.useNerdFont {
		return "\uf063" // nf-fa-arrow_down
	}

	return "v"
}

// Stale returns the stale/outdated symbol.
func (s Symbols) Stale() string {
	if s.useNerdFont {
		return "\uf017" // nf-fa-clock_o
	}

	return "o" // distinct from Loading ("...")
}

// Time returns the time/clock symbol.
func (s Symbols) Time() string {
	if s.useNerdFont {
		return "\uf017" // nf-fa-clock_o
	}

	return "@"
}

// Cursor returns the cursor/selection indicator.
func (s Symbols) Cursor() string {
	if s.useNerdFont {
		return "\uf054" // nf-fa-chevron_right
	}

	return ">"
}

// NoCursor returns an empty cursor indicator (space).
func (s Symbols) NoCursor() string {
	return " "
}
