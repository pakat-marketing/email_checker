package emailchecker

import "regexp"

// emailPattern is the same regex used by the original Elixir library,
// adapted for Go's RE2 syntax (named capture uses (?P<name>...) and the
// caseless flag is expressed inline as (?i)).
//
// The pattern is split so the literal backtick character inside the
// local-part class can be embedded without breaking the raw string.
const emailPattern = `(?i)^(?:[a-z0-9!#$%&'*+/=?^_` + "`" +
	`{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_` + "`" +
	`{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?P<domain>(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\]))$`

var emailRegex = regexp.MustCompile(emailPattern)

var domainSubmatchIndex = emailRegex.SubexpIndex("domain")

// Format reports whether email matches the address regex. It performs no
// network I/O.
func Format(email string) bool {
	return emailRegex.MatchString(email)
}

// Domain extracts the domain portion of email. It returns the empty
// string if email does not match the format regex.
func Domain(email string) string {
	m := emailRegex.FindStringSubmatch(email)
	if m == nil || domainSubmatchIndex < 0 {
		return ""
	}
	return m[domainSubmatchIndex]
}
