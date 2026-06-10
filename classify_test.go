package emailchecker

import "testing"

func TestIsDisposable(t *testing.T) {
	cases := []struct {
		email string
		want  bool
	}{
		{"user@mailinator.com", true},
		{"user@MAILINATOR.COM", true}, // case-insensitive
		{"user@guerrillamail.com", true},
		{"user@gmail.com", false},
		{"user@mycompany.com", false},
		{"not-an-email", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := IsDisposable(tc.email); got != tc.want {
			t.Errorf("IsDisposable(%q) = %v, want %v", tc.email, got, tc.want)
		}
	}
}

func TestIsRoleBased(t *testing.T) {
	cases := []struct {
		email string
		want  bool
	}{
		{"info@example.com", true},
		{"INFO@example.com", true}, // case-insensitive
		{"support@example.com", true},
		{"noreply@example.com", true},
		{"kevin@example.com", false},
		{"info.kevin@example.com", false}, // not an exact role local part
		{"not-an-email", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := IsRoleBased(tc.email); got != tc.want {
			t.Errorf("IsRoleBased(%q) = %v, want %v", tc.email, got, tc.want)
		}
	}
}

func TestDefaultRolePrefixesNotMutated(t *testing.T) {
	before := len(DefaultRolePrefixes)
	_ = IsRoleBased("info@example.com")
	if len(DefaultRolePrefixes) != before {
		t.Errorf("DefaultRolePrefixes length changed from %d", before)
	}
}
