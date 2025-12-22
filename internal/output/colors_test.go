package output

import "testing"

func TestColorEnabledOverride(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("CANOPY_COLOR", "0")

	if ColorEnabled() {
		t.Fatal("expected ColorEnabled to be false when CANOPY_COLOR=0")
	}

	t.Setenv("CANOPY_COLOR", "1")

	if !ColorEnabled() {
		t.Fatal("expected ColorEnabled to be true when CANOPY_COLOR=1")
	}
}

func TestColorEnabledInvalidValue(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("CANOPY_COLOR", "invalid")

	if !ColorEnabled() {
		t.Fatal("expected invalid CANOPY_COLOR to default to true")
	}
}

func TestColorEnabledNoColorPrecedence(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("CANOPY_COLOR", "1")

	if ColorEnabled() {
		t.Fatal("expected NO_COLOR to disable color output")
	}
}

func TestColumnTruncation(t *testing.T) {
	t.Setenv("CANOPY_COLOR", "0")

	if got := Column("abcdef", 4, MutedStyle); got != "abcd" {
		t.Fatalf("expected truncated column, got %q", got)
	}

	if got := Column("ab", 4, MutedStyle); got != "ab  " {
		t.Fatalf("expected padded column, got %q", got)
	}
}

func TestSeparatorLine(t *testing.T) {
	line := SeparatorLine(4)
	if line != "────" {
		t.Fatalf("unexpected separator line: %q", line)
	}
}
