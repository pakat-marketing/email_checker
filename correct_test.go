package emailchecker

import "testing"

func TestCorrect(t *testing.T) {
	cases := []struct {
		name     string
		email    string
		wantOK   bool
		wantFull string
	}{
		// High-confidence: whole-domain off-by-one against a known domain.
		{"gmail transposition", "kevin@gmial.com", true, "kevin@gmail.com"},
		{"gmail missing letter", "kevin@gmai.com", true, "kevin@gmail.com"},
		{"gmail wrong letter", "kevin@gnail.com", true, "kevin@gmail.com"},
		{"gmail tld typo", "kevin@gmail.con", true, "kevin@gmail.com"},
		{"hotmail typo", "kevin@hotmial.com", true, "kevin@hotmail.com"},
		{"outlook typo", "kevin@outlok.com", true, "kevin@outlook.com"},
		{"icloud tld typo", "kevin@icloud.con", true, "kevin@icloud.com"},
		{"comcast tld typo", "kevin@comcast.nett", true, "kevin@comcast.net"},
		{"protonmail typo", "kevin@protonmial.com", true, "kevin@protonmail.com"},

		// Already valid: no correction.
		{"valid gmail", "kevin@gmail.com", false, ""},
		{"valid hotmail", "kevin@hotmail.com", false, ""},
		{"valid yahoo", "kevin@yahoo.com", false, ""},
		{"valid outlook", "kevin@outlook.com", false, ""},
		// Valid providers modeled via second-level + top-level lists must
		// never be silently rewritten, even if a full domain is distance 1
		// away (mail.com is 1 from gmail.com).
		{"valid mail.com", "kevin@mail.com", false, ""},
		{"valid live.com", "kevin@live.com", false, ""},
		{"valid gmx.com", "kevin@gmx.com", false, ""},
		{"valid mail.co.uk", "kevin@mail.co.uk", false, ""},
		{"valid uppercase mail.com", "kevin@MAIL.COM", false, ""},

		// Low confidence: must NOT silently rewrite.
		{"valid uncommon tld", "kevin@mycompany.io", false, ""},
		{"unknown domain", "kevin@somerandomdomain.com", false, ""},

		// Malformed: no correction.
		{"empty", "", false, ""},
		{"no at", "plainaddress", false, ""},
		{"empty domain", "kevin@", false, ""},
		{"empty local", "@gmail.com", false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := Correct(tc.email)
			if ok != tc.wantOK {
				t.Fatalf("Correct(%q) ok = %v, want %v (got %+v)", tc.email, ok, tc.wantOK, got)
			}
			if ok && got.Full != tc.wantFull {
				t.Errorf("Correct(%q).Full = %q, want %q", tc.email, got.Full, tc.wantFull)
			}
			if ok && got.Confidence != ConfidenceHigh {
				t.Errorf("Correct(%q).Confidence = %v, want high", tc.email, got.Confidence)
			}
		})
	}
}

func TestCorrectPreservesLocalPartCase(t *testing.T) {
	got, ok := Correct("Kevin.O'Brien@GMIAL.COM")
	if !ok {
		t.Fatal("expected a correction")
	}
	if got.Address != "Kevin.O'Brien" {
		t.Errorf("Address = %q, want %q (local part case must be preserved)", got.Address, "Kevin.O'Brien")
	}
	if got.Domain != "gmail.com" {
		t.Errorf("Domain = %q, want %q", got.Domain, "gmail.com")
	}
	if got.Full != "Kevin.O'Brien@gmail.com" {
		t.Errorf("Full = %q, want %q", got.Full, "Kevin.O'Brien@gmail.com")
	}
}

func TestCorrectIgnoresConfigurableThreshold(t *testing.T) {
	// Even with a generous DomainThreshold, Correct stays at distance 1.
	s := NewSuggester(SuggestOptions{
		Domains:         []string{"example.com"},
		DomainThreshold: 5,
	})
	// "exmple.com" is distance 1 -> corrected.
	if got, ok := s.Correct("user@exmple.com"); !ok || got.Full != "user@example.com" {
		t.Errorf("Correct(exmple.com) = %+v, %v; want user@example.com, true", got, ok)
	}
	// "exampls.org" is farther than distance 1 -> not corrected, despite threshold 5.
	if got, ok := s.Correct("user@something.com"); ok {
		t.Errorf("Correct(something.com) = %+v; want no correction", got)
	}
}

func TestCorrectDomainsIncludesMajorProviders(t *testing.T) {
	for _, d := range []string{"yahoo.com", "hotmail.com", "outlook.com", "live.com"} {
		if !contains(CorrectDomains, d) {
			t.Errorf("CorrectDomains is missing %q", d)
		}
	}
}
