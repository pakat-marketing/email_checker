package emailchecker

import "testing"

func TestSuggest(t *testing.T) {
	cases := []struct {
		name     string
		email    string
		wantOK   bool
		wantFull string
	}{
		{"full domain typo", "test@gmial.com", true, "test@gmail.com"},
		{"missing tld char", "test@gmail.co", true, "test@gmail.com"},
		{"top level typo", "test@hotmail.cmo", true, "test@hotmail.com"},
		{"second level typo", "test@hotmial.com", true, "test@hotmail.com"},
		{"already valid full domain", "test@gmail.com", false, ""},
		{"already valid second+top", "test@hotmail.com", false, ""},
		{"empty input", "", false, ""},
		{"no at sign", "plainaddress", false, ""},
		{"empty domain", "test@", false, ""},
		{"empty local", "@gmail.com", false, ""},
		{"unknown domain no close match", "test@somerandomdomain.com", false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := Suggest(tc.email)
			if ok != tc.wantOK {
				t.Fatalf("Suggest(%q) ok = %v, want %v (got %+v)", tc.email, ok, tc.wantOK, got)
			}
			if ok && got.Full != tc.wantFull {
				t.Errorf("Suggest(%q).Full = %q, want %q", tc.email, got.Full, tc.wantFull)
			}
		})
	}
}

func TestSuggestParts(t *testing.T) {
	got, ok := Suggest("test@gmial.com")
	if !ok {
		t.Fatal("expected a suggestion")
	}
	if got.Address != "test" {
		t.Errorf("Address = %q, want %q", got.Address, "test")
	}
	if got.Domain != "gmail.com" {
		t.Errorf("Domain = %q, want %q", got.Domain, "gmail.com")
	}
	if got.Full != "test@gmail.com" {
		t.Errorf("Full = %q, want %q", got.Full, "test@gmail.com")
	}
}

func TestSuggestLowercasesInput(t *testing.T) {
	got, ok := Suggest("Test@GMIAL.COM")
	if !ok {
		t.Fatal("expected a suggestion")
	}
	if got.Full != "test@gmail.com" {
		t.Errorf("Full = %q, want %q", got.Full, "test@gmail.com")
	}
}

func TestSuggesterCustomDomains(t *testing.T) {
	s := NewSuggester(SuggestOptions{
		Domains: []string{"example.com"},
	})
	got, ok := s.Suggest("user@exambackground.com")
	// "exambackground.com" is far from "example.com"; should not match.
	if ok {
		t.Fatalf("did not expect a suggestion, got %+v", got)
	}

	got, ok = s.Suggest("user@exampl.com")
	if !ok {
		t.Fatalf("expected suggestion for exampl.com")
	}
	if got.Full != "user@example.com" {
		t.Errorf("Full = %q, want %q", got.Full, "user@example.com")
	}
}

func TestSuggesterCustomDistance(t *testing.T) {
	called := false
	s := NewSuggester(SuggestOptions{
		Distance: func(a, b string) int {
			called = true
			// Everything is "far", so no suggestion should be made.
			return 100
		},
	})
	if _, ok := s.Suggest("test@gmial.com"); ok {
		t.Error("expected no suggestion with always-far distance")
	}
	if !called {
		t.Error("custom Distance function was not invoked")
	}
}

func TestSuggestConfidence(t *testing.T) {
	cases := []struct {
		email string
		want  Confidence
	}{
		{"test@gmial.com", ConfidenceHigh},  // whole-domain distance 1
		{"test@hotmial.com", ConfidenceLow}, // component (second-level) path
	}
	for _, tc := range cases {
		got, ok := Suggest(tc.email)
		if !ok {
			t.Fatalf("Suggest(%q): expected a suggestion", tc.email)
		}
		if got.Confidence != tc.want {
			t.Errorf("Suggest(%q).Confidence = %v, want %v", tc.email, got.Confidence, tc.want)
		}
	}
}

func TestDefaultListsNotMutated(t *testing.T) {
	before := len(DefaultDomains)
	s := NewSuggester(SuggestOptions{})
	_, _ = s.Suggest("test@gmial.com")
	if len(DefaultDomains) != before {
		t.Errorf("DefaultDomains length changed from %d to %d", before, len(DefaultDomains))
	}
}
