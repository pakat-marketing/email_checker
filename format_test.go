package emailchecker

import "testing"

func TestFormat(t *testing.T) {
	cases := []struct {
		email string
		want  bool
	}{
		{"test@test.ch", true},
		{"a.b+c@example.co.uk", true},
		{"\"quoted\"@example.com", true},
		{"plainaddress", false},
		{"@no-local.com", false},
		{"no-domain@", false},
		{"two@@signs.com", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := Format(tc.email); got != tc.want {
			t.Errorf("Format(%q) = %v, want %v", tc.email, got, tc.want)
		}
	}
}

func TestDomain(t *testing.T) {
	cases := []struct {
		email string
		want  string
	}{
		{"kevin@disneur.me", "disneur.me"},
		{"a.b@sub.example.com", "sub.example.com"},
		{"not-an-email", ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := Domain(tc.email); got != tc.want {
			t.Errorf("Domain(%q) = %q, want %q", tc.email, got, tc.want)
		}
	}
}

func TestValidateShortCircuits(t *testing.T) {
	calls := 0
	count := func(string) bool { calls++; return true }
	fail := func(string) bool { return false }

	if Validate("x@y.com", count, fail, count) {
		t.Fatal("expected Validate to return false when a check fails")
	}
	if calls != 1 {
		t.Fatalf("expected exactly one call before short-circuit, got %d", calls)
	}
}

func TestValidateEmpty(t *testing.T) {
	if !Validate("anything") {
		t.Fatal("Validate with no checks should return true")
	}
}
