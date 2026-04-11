package output

import "testing"

func TestSanitizeInlineMessageCollapsesWhitespace(t *testing.T) {
	t.Parallel()

	input := "  first\tsecond\nthird\r\n  fourth  "
	got := SanitizeInlineMessage(input, 0)

	if got != "first second third fourth" {
		t.Fatalf("SanitizeInlineMessage() = %q, want %q", got, "first second third fourth")
	}
}
