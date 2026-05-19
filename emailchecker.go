// Package emailchecker validates email addresses through a configurable
// pipeline of checks: regex format, MX record presence, and SMTP RCPT TO.
//
// The package depends only on the Go standard library.
//
// Basic usage:
//
//	ok := emailchecker.Valid("kevin@disneur.me")
//
// The default pipeline runs Format then MX. To run a custom pipeline:
//
//	ok := emailchecker.Validate("kevin@disneur.me", emailchecker.Format, emailchecker.MX)
//
// SMTP checking is disabled by default because most providers no longer
// honour RCPT TO probes:
//
//	ok := emailchecker.Validate(addr, emailchecker.Format, emailchecker.MX, emailchecker.SMTP)
package emailchecker

import "time"

// Check is a single validation step. It returns true when email passes the
// step. A Check should be safe for concurrent use.
type Check func(email string) bool

// DefaultChecks is the pipeline used by Valid: format then MX.
var DefaultChecks = []Check{Format, MX}

// Config controls the network-dependent checks (MX, SMTP).
//
// The zero value is valid and matches the historical Elixir defaults: a
// system resolver, no retries beyond what each check performs, and no
// timeout. Override fields as needed and pass to NewChecker.
type Config struct {
	// Timeout bounds each network operation (DNS lookup, SMTP dial, SMTP
	// read/write). Zero means no timeout.
	Timeout time.Duration

	// SMTPRetries is the number of times SMTP will retry on a network
	// error. Values <= 0 default to 2, matching the Elixir library.
	SMTPRetries int

	// HelloName is the hostname announced in the SMTP HELO. If empty,
	// the recipient's domain is used (matching the Elixir behaviour).
	HelloName string

	// MailFrom is the sender used in the SMTP MAIL FROM command. If
	// empty, "fake@email.com" is used, matching the Elixir default.
	MailFrom string
}

// Checker bundles configuration so MX/SMTP checks can be created with
// non-default timeouts or retry counts.
type Checker struct {
	cfg Config
}

// NewChecker returns a Checker bound to cfg. Use its Format, MX and SMTP
// methods to build a custom pipeline.
func NewChecker(cfg Config) *Checker {
	return &Checker{cfg: cfg}
}

// Valid runs the default pipeline (Format, MX) against email.
func Valid(email string) bool {
	return Validate(email, DefaultChecks...)
}

// Validate runs each check against email, short-circuiting on the first
// failure. A pipeline with no checks always returns true.
func Validate(email string, checks ...Check) bool {
	for _, c := range checks {
		if !c(email) {
			return false
		}
	}
	return true
}
