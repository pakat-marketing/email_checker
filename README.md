# emailchecker

A small, zero-dependency Go package that validates email addresses through
a configurable pipeline of checks. Ported from the Elixir
[`email_checker`](https://github.com/maennchen/email_checker) library.

Checks are run in order and short-circuit on the first failure:

- **Format** — the address matches a permissive RFC 5322 regex.
- **MX** — the domain publishes at least one MX record.
- **SMTP** *(opt-in)* — the lowest-priority MX accepts the address in
  response to `RCPT TO`. No mail is sent.

> Most providers no longer honour the SMTP probe and many treat it as
> abuse. Some accept every address (catch-all). Use it only when you know
> the target supports it.

## Installation

```sh
go get github.com/pakat-marketing/email_checker
```

The package has no module dependencies — only the Go standard library.

## Usage

```go
import "github.com/pakat-marketing/email_checker"

// Default pipeline: Format + MX.
ok := emailchecker.Valid("kevin@disneur.me")

// Custom pipeline.
ok = emailchecker.Validate("kevin@disneur.me",
    emailchecker.Format,
    emailchecker.MX,
    emailchecker.SMTP,
)

// Format-only (no network).
ok = emailchecker.Format("test@test.ch")

// Distinguish "no MX" from "DNS error". Useful for pre-filters that
// shouldn't drop addresses just because the local resolver hiccuped:
//
//   (true,  nil)  → domain has MX records
//   (false, nil)  → domain has zero MX records (safe to filter)
//   (false, err)  → lookup failed; treat as unknown, not as no-MX
hasMX, err := emailchecker.CheckMX("kevin@disneur.me")
```

### Did-you-mean suggestions

`Suggest` corrects likely domain typos (a port of
[mailcheck.js](https://github.com/pakat-marketing/mailcheck)). It does no
network I/O — it matches the domain against a list of known good domains
using a string-distance function.

```go
// Default domain lists.
if s, ok := emailchecker.Suggest("kevin@gmial.com"); ok {
    fmt.Println(s.Full)       // "kevin@gmail.com"
    fmt.Println(s.Address)    // "kevin"
    fmt.Println(s.Domain)     // "gmail.com"
    fmt.Println(s.Confidence) // "high", "medium", or "low"
}

// ok is false when the address is unparseable, already good, or has no
// close match:
_, ok := emailchecker.Suggest("kevin@gmail.com") // ok == false (already good)

// Custom domain lists / thresholds / distance function. Extend the
// built-in lists rather than replacing them:
s := emailchecker.NewSuggester(emailchecker.SuggestOptions{
    Domains: append(emailchecker.DefaultDomains, "mycompany.com"),
})
suggestion, ok := s.Suggest("kevin@mycompny.com") // -> kevin@mycompany.com
```

The `Suggestion` struct also carries a `Confidence` field so callers can
decide what to act on:

| `Confidence` value | String     | When set                                                    |
| ------------------ | ---------- | ----------------------------------------------------------- |
| `ConfidenceHigh`   | `"high"`   | Whole-domain match, distance 1 (single-char typo).          |
| `ConfidenceMedium` | `"medium"` | Whole-domain match, distance 2.                             |
| `ConfidenceLow`    | `"low"`    | Component-level reconstruction (TLD or SLD swap/fix).       |

`Confidence` is ordered so `ConfidenceHigh > ConfidenceMedium > ConfidenceLow`,
allowing comparisons like `if s.Confidence >= emailchecker.ConfidenceMedium`.

| Field                  | Default                     | Meaning                                       |
| ---------------------- | --------------------------- | --------------------------------------------- |
| `Domains`              | `DefaultDomains`            | Full domains matched against the whole domain. |
| `SecondLevelDomains`   | `DefaultSecondLevelDomains` | Known second-level parts (e.g. `gmail`).       |
| `TopLevelDomains`      | `DefaultTopLevelDomains`    | Known top-level parts (e.g. `com`, `co.uk`).   |
| `DomainThreshold`      | `2`                         | Max distance for a full-domain match.          |
| `SecondLevelThreshold` | `2`                         | Max distance for a second-level match.         |
| `TopLevelThreshold`    | `2`                         | Max distance for a top-level match.            |
| `Distance`             | `sift4`                     | Pluggable `func(a, b string) int`.             |

### High-confidence auto-correction

`Correct` is a stricter variant intended for **silently rewriting** an
address (no user confirmation). It only returns a correction when the whole
domain is a single-character (distance 1) typo of a known good domain — it
never does the component-level TLD/SLD reconstruction that `Suggest` does,
and the distance bound is fixed (not configurable), so it can't be loosened
by accident.

```go
if c, ok := emailchecker.Correct("Kevin@gmial.com"); ok {
    email = c.Full // "Kevin@gmail.com" — safe to write back
}
```

Differences from `Suggest`:

- **Preserves the local part exactly** (only the domain is lowercased for
  matching), so `Kevin.O@GMIAL.COM` → `Kevin.O@gmail.com`.
- Matches against `CorrectDomains` (= `DefaultDomains` + major providers
  `yahoo.com`, `hotmail.com`, `outlook.com`, `live.com`, `proton.me`,
  `protonmail.com`, …) so Yahoo/Hotmail/Outlook typos are covered.
- Returns `ok=false` for valid addresses, distance ≥ 2, modern TLDs with no
  close match (`mycompany.io`, `startup.ai`), and malformed input — in all
  those cases, leave the address untouched.

Extend the target list rather than replacing it:

```go
s := emailchecker.NewSuggester(emailchecker.SuggestOptions{
    Domains: append(emailchecker.CorrectDomains, "mycompany.com"),
})
c, ok := s.Correct("kevin@mycompny.com") // -> kevin@mycompany.com
```

### Disposable & role-based detection

Two offline, zero-network checks for common low-quality address classes.

```go
// Disposable / throwaway domain check.
emailchecker.IsDisposable("user@mailinator.com")  // true
emailchecker.IsDisposable("user@gmail.com")        // false

// Bare-domain variant (useful when you already hold the domain):
emailchecker.IsDisposableDomain("mailinator.com")  // true

// Role / distribution address check (info@, support@, noreply@, …).
emailchecker.IsRoleBased("info@example.com")    // true
emailchecker.IsRoleBased("kevin@example.com")   // false
```

Both checks are case-insensitive and return `false` for unparseable input.

The disposable-domain list is an offline starter set embedded from
`disposable_domains.txt`. Drop a larger list (e.g. the community
`disposable-email-domains` blocklist) into that file to extend it — the
parser skips blank lines and `#` comments.

`DefaultRolePrefixes` is exported so you can extend it at startup:

```go
emailchecker.DefaultRolePrefixes = append(emailchecker.DefaultRolePrefixes, "enquiries")
```

Note: `IsRoleBased` matches the **exact** local part (so `info@` matches but
`info.team@` does not), which avoids false positives on real names.

### Configuration

For non-default timeouts or retry counts, build a `Checker` and use its
methods inside `Validate`:

```go
c := emailchecker.NewChecker(emailchecker.Config{
    Timeout:     6 * time.Second,
    SMTPRetries: 1,
    HelloName:   "client.invalid",
    MailFrom:    "probe@client.invalid",
})

ok := emailchecker.Validate("kevin@disneur.me", c.MX, c.SMTP)
```

| Field         | Default          | Meaning                                                              |
| ------------- | ---------------- | -------------------------------------------------------------------- |
| `Timeout`     | none             | Bound on each network operation (DNS, SMTP dial, SMTP read/write).   |
| `SMTPRetries` | `2`              | Retries on network errors during the SMTP probe.                     |
| `HelloName`   | recipient domain | Hostname announced in `HELO`.                                        |
| `MailFrom`    | `fake@email.com` | Sender used in `MAIL FROM`.                                          |

When `Timeout > 0`, the per-attempt SMTP timeout is `Timeout / SMTPRetries`,
matching the behaviour of the Elixir library.

## License

[MIT](./LICENSE)
