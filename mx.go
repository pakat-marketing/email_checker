package emailchecker

import (
	"context"
	"net"
	"sort"
	"strings"
	"time"
)

// MX reports whether the domain in email has at least one MX record.
// It uses the default system resolver and no timeout.
//
// MX collapses "no MX records" and "DNS lookup failed" into the same
// false return. If you need to distinguish them (so transient DNS
// errors don't get treated as deliverability signals), use CheckMX.
func MX(email string) bool {
	return (&Checker{}).MX(email)
}

// CheckMX is like MX but distinguishes "no records" from "lookup error".
//   - (true,  nil)  → the domain has at least one MX record
//   - (false, nil)  → the domain has zero MX records (or no domain at all)
//   - (false, err)  → the lookup failed (timeout, NXDOMAIN, refused, etc.)
//
// Callers that pre-filter emails before paying for a per-address API
// should forward on err != nil, since a transient resolver hiccup is
// indistinguishable from a real domain that has no MX.
func CheckMX(email string) (bool, error) {
	return (&Checker{}).CheckMX(email)
}

// MX reports whether the domain in email has at least one MX record,
// respecting the Checker's Timeout. See package-level MX for caveats.
func (c *Checker) MX(email string) bool {
	host, _ := c.lookupMX(email)
	return host != ""
}

// CheckMX is the (hasMX, err) variant of MX. See package-level CheckMX.
func (c *Checker) CheckMX(email string) (bool, error) {
	host, err := c.lookupMX(email)
	if err != nil {
		return false, err
	}
	return host != "", nil
}

// lookupMX returns the host of the lowest-priority MX record for the
// domain in email, plus any lookup error.
//   - "", nil         → domain has no MX records (or email had no domain)
//   - host, nil       → MX records present
//   - "", err         → DNS lookup failed
func (c *Checker) lookupMX(email string) (string, error) {
	domain := Domain(email)
	if domain == "" {
		return "", nil
	}

	ctx := context.Background()
	if c.cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.cfg.Timeout)
		defer cancel()
	}

	resolver := net.DefaultResolver
	records, err := resolver.LookupMX(ctx, domain)
	if err != nil {
		return "", err
	}
	if len(records) == 0 {
		return "", nil
	}

	sort.SliceStable(records, func(i, j int) bool {
		return records[i].Pref < records[j].Pref
	})

	// LookupMX returns FQDNs with a trailing dot; strip it so the host
	// can be used directly with net.Dial.
	return strings.TrimSuffix(records[0].Host, "."), nil
}

// dialTimeout returns the per-attempt timeout, dividing the overall
// Timeout by the retry count to mirror the Elixir behaviour.
func (c *Checker) dialTimeout() time.Duration {
	if c.cfg.Timeout <= 0 {
		return 0
	}
	retries := c.cfg.SMTPRetries
	if retries <= 0 {
		retries = defaultSMTPRetries
	}
	return c.cfg.Timeout / time.Duration(retries)
}
