// Package output provides helpers for CLI output formatting.
package output

import (
	"io"
	"os"

	"golang.org/x/term"
)

func isWriterTTY(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}

	return false
}
