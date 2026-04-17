package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestBoxRenderLineIgnoresANSIEscapeWidth(t *testing.T) {
	t.Setenv("CANOPY_COLOR", "1")
	t.Setenv("CANOPY_UNICODE", "0")

	var buf bytes.Buffer
	NewBox("Status").
		WithWidth(24).
		WithWriter(&buf).
		Render([]string{"  " + Colorize(SuccessStyle, "clean")})

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	for _, line := range lines {
		if got := lipgloss.Width(line); got != 24 {
			t.Fatalf("expected visible width 24, got %d for line %q", got, line)
		}
	}
}
