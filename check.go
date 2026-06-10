package emailchecker

import "strings"

// Result is the combined outcome of checking an email address. It is the
// "front door" of the package, bundling the individual checks into one value.
type Result struct {
	Address  string // the input address, trimmed of surrounding space
	FormatOK bool   // matches the address format regex
	HasMX    bool   // the domain publishes at least one MX record
	MXError  error  // non-nil if the DNS lookup itself failed (transient);
	// distinct from HasMX=false meaning "no MX records"
	Disposable bool        // domain is a known disposable provider
	RoleBased  bool        // local part is a role address (info@, support@, …)
	Suggestion *Suggestion // best typo suggestion (any confidence), or nil
	Correction *Suggestion // high-confidence correction safe to auto-apply, or nil
}

// Check runs the full set of checks against email using default settings and
// the system resolver. It performs a DNS MX lookup (network I/O). For
// non-default timeouts, build a Checker with NewChecker and call its Check.
func Check(email string) Result {
	return (&Checker{}).Check(email)
}

// Check is the Checker method form; it honors the Checker's Timeout for the
// MX lookup. See the package-level Check.
func (c *Checker) Check(email string) Result {
	addr := strings.TrimSpace(email)
	r := Result{Address: addr}

	r.FormatOK = Format(addr)

	hasMX, err := c.CheckMX(addr)
	r.HasMX, r.MXError = hasMX, err

	r.Disposable = IsDisposable(addr)
	r.RoleBased = IsRoleBased(addr)

	if s, ok := Suggest(addr); ok {
		r.Suggestion = &s
	}
	if cor, ok := Correct(addr); ok {
		r.Correction = &cor
	}

	return r
}
