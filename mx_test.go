package emailchecker

import "testing"

// TestMXBadFormat exercises only the offline path: an unparseable email
// short-circuits before any DNS lookup happens.
func TestMXBadFormat(t *testing.T) {
	if MX("not-an-email") {
		t.Fatal("MX should reject input that fails Format")
	}
	if MX("") {
		t.Fatal("MX should reject empty input")
	}
}
