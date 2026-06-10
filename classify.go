package emailchecker

import (
	"bufio"
	_ "embed"
	"strings"
)

//go:embed disposable_domains.txt
var disposableDomainsRaw string

// disposableDomains is the parsed set of known disposable provider domains,
// keyed by lowercase domain.
var disposableDomains = parseDomainSet(disposableDomainsRaw)

func parseDomainSet(raw string) map[string]struct{} {
	set := make(map[string]struct{})
	sc := bufio.NewScanner(strings.NewReader(raw))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		set[strings.ToLower(line)] = struct{}{}
	}
	return set
}

// DefaultRolePrefixes lists local parts that indicate a role/distribution
// address rather than an individual mailbox. It is exported so callers can
// extend it.
var DefaultRolePrefixes = []string{
	"admin", "administrator", "billing", "contact", "info", "help",
	"hello", "hr", "jobs", "mail", "marketing", "noreply", "no-reply",
	"office", "postmaster", "sales", "security", "support", "team",
	"webmaster", "abuse", "privacy", "press",
}

var rolePrefixSet = func() map[string]struct{} {
	set := make(map[string]struct{}, len(DefaultRolePrefixes))
	for _, p := range DefaultRolePrefixes {
		set[p] = struct{}{}
	}
	return set
}()

// IsDisposableDomain reports whether domain is a known disposable provider.
// Comparison is case-insensitive.
func IsDisposableDomain(domain string) bool {
	_, ok := disposableDomains[strings.ToLower(strings.TrimSpace(domain))]
	return ok
}

// IsDisposable reports whether email's domain is a known disposable provider.
// Unparseable input and unknown domains return false.
func IsDisposable(email string) bool {
	parts, ok := splitEmail(email)
	if !ok {
		return false
	}
	return IsDisposableDomain(parts.domain)
}

// IsRoleBased reports whether email's local part is a role/distribution
// address (info@, support@, …) rather than an individual mailbox.
// Comparison is case-insensitive. Unparseable input returns false.
func IsRoleBased(email string) bool {
	parts, ok := splitEmail(email)
	if !ok {
		return false
	}
	_, isRole := rolePrefixSet[strings.ToLower(parts.address)]
	return isRole
}
