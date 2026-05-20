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

// TestCheckMXBadFormat: a bad address short-circuits to (false, nil) —
// no DNS attempt, no error. The "no domain" case is treated as "no MX",
// not as a lookup error.
func TestCheckMXBadFormat(t *testing.T) {
	ok, err := CheckMX("not-an-email")
	if ok {
		t.Fatal("CheckMX should report false for unparseable input")
	}
	if err != nil {
		t.Fatalf("CheckMX should not return an error for unparseable input; got %v", err)
	}
	ok, err = CheckMX("")
	if ok || err != nil {
		t.Fatalf("CheckMX(\"\") = (%v, %v); want (false, nil)", ok, err)
	}
}
