package emailchecker

import "testing"

func TestCheckFormat(t *testing.T) {
	if got := Check("not-an-email"); got.FormatOK {
		t.Error("expected FormatOK=false for malformed input")
	}
	if got := Check("kevin@example.com"); !got.FormatOK {
		t.Error("expected FormatOK=true for a well-formed address")
	}
}

func TestCheckClassification(t *testing.T) {
	d := Check("user@mailinator.com")
	if !d.Disposable {
		t.Error("expected Disposable=true for mailinator.com")
	}
	r := Check("info@example.com")
	if !r.RoleBased {
		t.Error("expected RoleBased=true for info@")
	}
	n := Check("kevin@example.com")
	if n.Disposable || n.RoleBased {
		t.Errorf("expected no classification flags for kevin@example.com, got %+v", n)
	}
}

func TestCheckSuggestionAndCorrection(t *testing.T) {
	got := Check("kevin@gmial.com")
	if got.Correction == nil || got.Correction.Full != "kevin@gmail.com" {
		t.Errorf("expected high-confidence correction to kevin@gmail.com, got %+v", got.Correction)
	}
	if got.Suggestion == nil || got.Suggestion.Full != "kevin@gmail.com" {
		t.Errorf("expected suggestion to kevin@gmail.com, got %+v", got.Suggestion)
	}
	// A valid address yields no suggestion/correction.
	clean := Check("kevin@gmail.com")
	if clean.Suggestion != nil || clean.Correction != nil {
		t.Errorf("expected no suggestion/correction for a valid address, got %+v", clean)
	}
}

func TestCheckTrimsInput(t *testing.T) {
	if got := Check("  kevin@example.com  "); got.Address != "kevin@example.com" {
		t.Errorf("Address = %q, want trimmed", got.Address)
	}
}
