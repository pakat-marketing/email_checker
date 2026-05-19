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
func MX(email string) bool {
	return (&Checker{}).MX(email)
}

// MX reports whether the domain in email has at least one MX record,
// respecting the Checker's Timeout.
func (c *Checker) MX(email string) bool {
	return c.lookupMX(email) != ""
}

// lookupMX returns the host of the lowest-priority MX record for the
// domain in email, or "" if the lookup fails or yields no records.
func (c *Checker) lookupMX(email string) string {
	domain := Domain(email)
	if domain == "" {
		return ""
	}

	ctx := context.Background()
	if c.cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.cfg.Timeout)
		defer cancel()
	}

	resolver := net.DefaultResolver
	records, err := resolver.LookupMX(ctx, domain)
	if err != nil || len(records) == 0 {
		return ""
	}

	sort.SliceStable(records, func(i, j int) bool {
		return records[i].Pref < records[j].Pref
	})

	// LookupMX returns FQDNs with a trailing dot; strip it so the host
	// can be used directly with net.Dial.
	return strings.TrimSuffix(records[0].Host, ".")
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
