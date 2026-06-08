// This file ports the domain typo-suggestion feature from mailcheck.js
// (github.com/pakat-marketing/mailcheck). Given a likely-misspelled address
// such as "kevin@gmial.com" it suggests "kevin@gmail.com" by matching the
// domain against a list of known good domains using a string-distance
// function (sift4 by default).
//
// Suggestion is independent of the validation pipeline (Format/MX/SMTP): it
// answers "did you mean…?", not "is this deliverable?". It performs no I/O.
package emailchecker

import "strings"

// Default domain lists, copied verbatim from mailcheck.js. They are exported
// so callers can extend them (e.g. append company-specific domains) rather
// than having to replace the whole list.
var (
	DefaultDomains = []string{
		"msn.com", "bellsouth.net", "telus.net", "comcast.net", "optusnet.com.au",
		"earthlink.net", "qq.com", "sky.com", "icloud.com", "mac.com",
		"sympatico.ca", "googlemail.com", "att.net", "xtra.co.nz", "web.de",
		"cox.net", "gmail.com", "ymail.com", "aim.com", "rogers.com",
		"verizon.net", "rocketmail.com", "google.com", "optonline.net",
		"sbcglobal.net", "aol.com", "me.com", "btinternet.com", "charter.net",
		"shaw.ca",
	}

	DefaultSecondLevelDomains = []string{
		"yahoo", "hotmail", "mail", "live", "outlook", "gmx",
	}

	DefaultTopLevelDomains = []string{
		"com", "com.au", "com.tw", "ca", "co.nz", "co.uk", "de", "fr", "it",
		"ru", "net", "org", "edu", "gov", "jp", "nl", "kr", "se", "eu", "ie",
		"co.il", "us", "at", "be", "dk", "hk", "es", "gr", "ch", "no", "cz",
		"in", "net.au", "info", "biz", "mil", "co.jp", "sg", "hu", "uk",
	}
)

// defaultThreshold is the maximum distance, copied from mailcheck.js, at
// which a candidate is considered a plausible correction.
const defaultThreshold = 2

// Suggestion is a proposed correction for a misspelled email domain.
type Suggestion struct {
	Address string // local part, e.g. "kevin"
	Domain  string // corrected domain, e.g. "gmail.com"
	Full    string // corrected full email, e.g. "kevin@gmail.com"
}

// SuggestOptions configures a Suggester. The zero value is valid: nil slices
// fall back to the built-in defaults and non-positive thresholds fall back
// to 2.
type SuggestOptions struct {
	Domains              []string     // full domains; nil → DefaultDomains
	SecondLevelDomains   []string     // nil → DefaultSecondLevelDomains
	TopLevelDomains      []string     // nil → DefaultTopLevelDomains
	DomainThreshold      int          // <=0 → 2
	SecondLevelThreshold int          // <=0 → 2
	TopLevelThreshold    int          // <=0 → 2
	Distance             DistanceFunc // nil → sift4
}

// Suggester holds resolved options so suggestions can be made repeatedly
// without re-resolving defaults. It is safe for concurrent use.
type Suggester struct {
	domains     []string
	secondLevel []string
	topLevel    []string
	domainThr   int
	secondThr   int
	topThr      int
	distance    DistanceFunc
}

// NewSuggester returns a Suggester bound to opts, with defaults applied.
func NewSuggester(opts SuggestOptions) *Suggester {
	s := &Suggester{
		domains:     opts.Domains,
		secondLevel: opts.SecondLevelDomains,
		topLevel:    opts.TopLevelDomains,
		domainThr:   opts.DomainThreshold,
		secondThr:   opts.SecondLevelThreshold,
		topThr:      opts.TopLevelThreshold,
		distance:    opts.Distance,
	}
	if s.domains == nil {
		s.domains = DefaultDomains
	}
	if s.secondLevel == nil {
		s.secondLevel = DefaultSecondLevelDomains
	}
	if s.topLevel == nil {
		s.topLevel = DefaultTopLevelDomains
	}
	if s.domainThr <= 0 {
		s.domainThr = defaultThreshold
	}
	if s.secondThr <= 0 {
		s.secondThr = defaultThreshold
	}
	if s.topThr <= 0 {
		s.topThr = defaultThreshold
	}
	if s.distance == nil {
		s.distance = sift4
	}
	return s
}

// defaultSuggester backs the package-level Suggest function.
var defaultSuggester = NewSuggester(SuggestOptions{})

// Suggest returns a correction for email using the built-in default domain
// lists, or ok=false if the address is unparseable, already good, or has no
// close match.
func Suggest(email string) (Suggestion, bool) {
	return defaultSuggester.Suggest(email)
}

// Suggest returns a correction for email, or ok=false if the address is
// unparseable, already good, or has no close match.
func (s *Suggester) Suggest(email string) (Suggestion, bool) {
	email = strings.ToLower(email)

	parts, ok := splitEmail(email)
	if !ok {
		return Suggestion{}, false
	}

	// If the domain is already a known second-level + top-level combination
	// (e.g. hotmail.com), it's good as-is.
	if contains(s.secondLevel, parts.secondLevelDomain) && contains(s.topLevel, parts.topLevelDomain) {
		return Suggestion{}, false
	}

	// First try matching the whole domain.
	if closest, found := findClosestDomain(parts.domain, s.domains, s.distance, s.domainThr); found {
		if closest == parts.domain {
			return Suggestion{}, false
		}
		return newSuggestion(parts.address, closest), true
	}

	// Otherwise correct the second-level and top-level parts separately.
	closestSLD, foundSLD := findClosestDomain(parts.secondLevelDomain, s.secondLevel, s.distance, s.secondThr)
	closestTLD, foundTLD := findClosestDomain(parts.topLevelDomain, s.topLevel, s.distance, s.topThr)

	if parts.domain == "" {
		return Suggestion{}, false
	}

	domain := parts.domain
	changed := false
	if foundSLD && closestSLD != parts.secondLevelDomain {
		domain = strings.Replace(domain, parts.secondLevelDomain, closestSLD, 1)
		changed = true
	}
	if foundTLD && closestTLD != parts.topLevelDomain && parts.secondLevelDomain != "" {
		domain = replaceSuffix(domain, parts.topLevelDomain, closestTLD)
		changed = true
	}
	if changed {
		return newSuggestion(parts.address, domain), true
	}
	return Suggestion{}, false
}

func newSuggestion(address, domain string) Suggestion {
	return Suggestion{Address: address, Domain: domain, Full: address + "@" + domain}
}

// emailParts mirrors the structure mailcheck.js extracts from an address.
type emailParts struct {
	topLevelDomain    string
	secondLevelDomain string
	domain            string
	address           string
}

// splitEmail parses email into its parts. It returns ok=false on empty
// input, a missing "@", or any empty component (e.g. "a@" or "@b.com").
func splitEmail(email string) (emailParts, bool) {
	email = strings.TrimSpace(email)
	parts := strings.Split(email, "@")
	if len(parts) < 2 {
		return emailParts{}, false
	}
	for _, p := range parts {
		if p == "" {
			return emailParts{}, false
		}
	}

	domain := parts[len(parts)-1]
	address := strings.Join(parts[:len(parts)-1], "@")
	sld, tld := splitDomain(domain)

	return emailParts{
		topLevelDomain:    tld,
		secondLevelDomain: sld,
		domain:            domain,
		address:           address,
	}, true
}

// splitDomain splits a domain into its second-level and top-level parts,
// matching mailcheck.js: the first label is the second-level part and the
// rest (joined by ".") is the top-level part. A single-label domain has an
// empty second-level part.
func splitDomain(domain string) (sld, tld string) {
	domainParts := strings.Split(domain, ".")
	if len(domainParts) == 1 {
		return "", domainParts[0]
	}
	return domainParts[0], strings.Join(domainParts[1:], ".")
}

// findClosestDomain returns the entry in domains closest to domain, provided
// it is within threshold. An exact match short-circuits. Ties are resolved
// by first-encountered, matching mailcheck.js iteration order.
func findClosestDomain(domain string, domains []string, dist DistanceFunc, threshold int) (string, bool) {
	closest, _, ok := closestDomainWithin(domain, domains, dist, threshold)
	return closest, ok
}

// closestDomainWithin is findClosestDomain that also returns the matching
// distance (0 for an exact match), so callers can apply a stricter policy
// than the threshold alone.
func closestDomainWithin(domain string, domains []string, dist DistanceFunc, threshold int) (string, int, bool) {
	if domain == "" || len(domains) == 0 {
		return "", 0, false
	}

	minDist := -1
	closest := ""
	for _, d := range domains {
		if domain == d {
			return domain, 0, true
		}
		score := dist(domain, d)
		if minDist < 0 || score < minDist {
			minDist = score
			closest = d
		}
	}

	if closest != "" && minDist <= threshold {
		return closest, minDist, true
	}
	return "", 0, false
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// replaceSuffix replaces oldSuffix with newSuffix only when s ends with
// oldSuffix, mirroring mailcheck.js's end-anchored top-level replacement.
func replaceSuffix(s, oldSuffix, newSuffix string) string {
	if strings.HasSuffix(s, oldSuffix) {
		return s[:len(s)-len(oldSuffix)] + newSuffix
	}
	return s
}
