package output

import "io"

// WriteStructuredReport writes a raw JSON report or renders a human-readable report.
// Unlike PrintJSON, this preserves the caller's existing top-level JSON shape.
func WriteStructuredReport(w io.Writer, report interface{}, jsonOutput bool, renderHuman func(io.Writer) error) error {
	if jsonOutput {
		return WriteIndentedJSON(w, report)
	}

	return renderHuman(w)
}
