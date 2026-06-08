package emailchecker

import "strings"

// CorrectDomains is the curated full-domain target list used by the
// package-level Correct. It is DefaultDomains plus the major consumer
// providers that mailcheck.js models as second-level parts rather than full
// domains (yahoo.com, hotmail.com, outlook.com, live.com, …), so that
// off-by-one typos of those providers are caught too.
//
// It is exported so callers can extend it, e.g.:
//
//	s := emailchecker.NewSuggester(emailchecker.SuggestOptions{
//	    Domains: append(emailchecker.CorrectDomains, "mycompany.com"),
//	})
var CorrectDomains = func() []string {
	extra := []string{
		"yahoo.com", "hotmail.com", "outlook.com", "live.com",
		"hotmail.co.uk", "yahoo.co.uk", "proton.me", "protonmail.com",
	}
	out := make([]string, 0, len(DefaultDomains)+len(extra))
	out = append(out, DefaultDomains...)
	out = append(out, extra...)
	return out
}()

// highConfidenceMaxDistance is the fixed distance bound for Correct. It is
// intentionally not configurable: Correct rewrites real addresses silently,
// so the bar must not be loosened by accident. Distance 1 means a
// single-character typo of a known good domain.
const highConfidenceMaxDistance = 1

// defaultCorrector backs the package-level Correct function. It matches
// against CorrectDomains.
var defaultCorrector = NewSuggester(SuggestOptions{Domains: CorrectDomains})

// Correct returns a high-confidence domain correction suitable for silently
// rewriting an address. It returns ok=true only when the whole domain is a
// single-character (distance 1) typo of a known good domain in
// CorrectDomains. It never performs component-level (TLD/SLD) reconstruction
// and never widens beyond distance 1.
//
// Unlike Suggest, Correct preserves the local part exactly as given (only
// the domain is lowercased for matching), so it is safe to write the result
// back as the user's address.
//
// ok=false when the address is unparseable, already valid, or has no
// distance-1 match — in those cases leave the original address untouched.
func Correct(email string) (Suggestion, bool) {
	return defaultCorrector.Correct(email)
}

// Correct is the Suggester method form, matching against the Suggester's
// configured Domains. See the package-level Correct for semantics.
func (s *Suggester) Correct(email string) (Suggestion, bool) {
	parts, ok := splitEmail(email)
	if !ok {
		return Suggestion{}, false
	}

	domain := strings.ToLower(parts.domain)

	// Never rewrite an address that is already a valid known second-level +
	// top-level combination (e.g. mail.com, live.com, gmx.com), even when a
	// full domain happens to be distance 1 away (mail.com is 1 from
	// gmail.com). Without this guard the silent path would corrupt valid
	// addresses on providers that the lists model via their parts.
	sld, tld := splitDomain(domain)
	if contains(s.secondLevel, sld) && contains(s.topLevel, tld) {
		return Suggestion{}, false
	}

	closest, dist, found := closestDomainWithin(domain, s.domains, s.distance, highConfidenceMaxDistance)
	if !found || dist == 0 {
		// No match, or already an exact known domain: nothing to fix.
		return Suggestion{}, false
	}

	// parts.address keeps the original local-part casing.
	return newSuggestion(parts.address, closest), true
}
