package output

import "testing"

func TestColorEnabledOverride(t *testing.T) {
	t.Setenv("CANOPY_COLOR", "0")

	if ColorEnabled() {
		t.Fatal("expected ColorEnabled to be false when CANOPY_COLOR=0")
	}

	t.Setenv("CANOPY_COLOR", "1")

	if !ColorEnabled() {
		t.Fatal("expected ColorEnabled to be true when CANOPY_COLOR=1")
	}
}

func TestSeparatorLine(t *testing.T) {
	line := SeparatorLine(4)
	if line != "────" {
		t.Fatalf("unexpected separator line: %q", line)
	}
}
