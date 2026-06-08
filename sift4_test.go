package emailchecker

import "testing"

func TestSift4(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"gmail.com", "gmail.com", 0},
		{"", "", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1}, // single substitution
		{"ab", "ba", 1},   // transposition
	}
	for _, tc := range cases {
		if got := sift4(tc.a, tc.b); got != tc.want {
			t.Errorf("sift4(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestSift4Symmetricish(t *testing.T) {
	// Close typos should score within the default threshold of 2.
	close := []struct{ a, b string }{
		{"gmial.com", "gmail.com"},
		{"gmail.co", "gmail.com"},
		{"hotmal", "hotmail"},
		{"cmo", "com"},
	}
	for _, tc := range close {
		if d := sift4(tc.a, tc.b); d > 2 {
			t.Errorf("sift4(%q, %q) = %d, want <= 2", tc.a, tc.b, d)
		}
	}
}
