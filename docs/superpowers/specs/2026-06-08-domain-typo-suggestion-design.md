# Domain typo suggestion for `emailchecker`

Date: 2026-06-08

## Goal

Port the domain typo-suggestion feature from
[`pakat-marketing/mailcheck`](https://github.com/pakat-marketing/mailcheck)
(a fork of Kicksend's `mailcheck.js`) into this Go library. Given a
likely-misspelled address such as `kevin@gmial.com`, suggest the corrected
`kevin@gmail.com` by matching the domain against known good domains using a
string-distance algorithm.

This is **additive**: it sits alongside the existing boolean validation
pipeline (`Format`, `MX`, `SMTP`) and shares no state with it. Validation
answers "is this address deliverable?"; suggestion answers "did you mean…?".

## Scope

In scope:
- Typo-suggestion feature with configurable domain lists and thresholds.
- A pluggable distance function (default: `sift4`, matching the fork).
- Built-in default domain lists copied verbatim from the fork.

Out of scope (intentionally dropped from the JS source):
- `encodeEmail` — a browser XSS concern. Go callers handle their own
  escaping/HTML templating.
- The jQuery plugin and callback/`run(opts)` surface — replaced by Go's
  idiomatic `(Suggestion, bool)` return value.

## Architecture & files

One concern per file, matching the existing layout (`format.go`, `mx.go`,
`smtp.go`).

- **`suggest.go`** — public API and core suggestion logic:
  `Suggestion`, `SuggestOptions`, `Suggester`, `NewSuggester`, package-level
  `Suggest`, the default domain lists, and `splitEmail` /
  `findClosestDomain` helpers.
- **`sift4.go`** — the `sift4` distance function and the exported
  `DistanceFunc` type.
- **`suggest_test.go`**, **`sift4_test.go`** — tests.

Existing files are untouched except a new README section and a CHANGELOG
entry. `Config`/`Checker` (network settings) are not modified.

## Public API

```go
// Suggestion is a proposed correction for a misspelled email domain.
type Suggestion struct {
    Address string // local part, e.g. "kevin"
    Domain  string // corrected domain, e.g. "gmail.com"
    Full    string // corrected full email, e.g. "kevin@gmail.com"
}

// DistanceFunc returns the edit distance between two strings. Lower means
// more similar. Used to score candidate domains.
type DistanceFunc func(a, b string) int

// SuggestOptions configures a Suggester. The zero value is valid: nil
// slices fall back to the built-in defaults and non-positive thresholds
// fall back to 2.
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
// without re-resolving defaults. Safe for concurrent use.
type Suggester struct { /* resolved opts */ }

// NewSuggester returns a Suggester bound to opts (with defaults applied).
func NewSuggester(opts SuggestOptions) *Suggester

// Suggest (method) returns a correction for email, or ok=false if the
// address is unparseable, already good, or has no close match.
func (s *Suggester) Suggest(email string) (Suggestion, bool)

// Suggest (package-level) uses the built-in defaults.
func Suggest(email string) (Suggestion, bool)

// Exported defaults so callers can extend rather than replace them.
var (
    DefaultDomains              []string
    DefaultSecondLevelDomains   []string
    DefaultTopLevelDomains      []string
)
```

`ok` is `false` when:
- the email is empty or has no parseable `local@domain` shape, or
- the domain is already considered good (matches a known second-level +
  top-level combination), or
- no candidate is within threshold, or
- the best candidate equals the input domain (no real change).

### Default lists (verbatim from the fork)

- `DefaultDomains`: `msn.com, bellsouth.net, telus.net, comcast.net,
  optusnet.com.au, earthlink.net, qq.com, sky.com, icloud.com, mac.com,
  sympatico.ca, googlemail.com, att.net, xtra.co.nz, web.de, cox.net,
  gmail.com, ymail.com, aim.com, rogers.com, verizon.net, rocketmail.com,
  google.com, optonline.net, sbcglobal.net, aol.com, me.com, btinternet.com,
  charter.net, shaw.ca`
- `DefaultSecondLevelDomains`: `yahoo, hotmail, mail, live, outlook, gmx`
- `DefaultTopLevelDomains`: `com, com.au, com.tw, ca, co.nz, co.uk, de, fr,
  it, ru, net, org, edu, gov, jp, nl, kr, se, eu, ie, co.il, us, at, be, dk,
  hk, es, gr, ch, no, cz, in, net.au, info, biz, mil, co.jp, sg, hu, uk`
- All three thresholds default to `2`.

## Algorithm (ported from mailcheck.js)

1. **lowercase + splitEmail(email)** — the whole address is lowercased
   first (matching mailcheck.js, so the returned `Address` is lowercased
   too), then trimmed and split: the part after the last `@` is the domain,
   everything before is the local part. Return not-ok on empty input,
   missing `@`, empty local part, or empty domain.
2. **already-good check** — split the domain into second-level and
   top-level parts. If the second-level part is in `SecondLevelDomains`
   *and* the top-level part is in `TopLevelDomains`, return no suggestion.
3. **full-domain match** — find the closest entry in `Domains` via
   `findClosestDomain(domain, Domains, DomainThreshold)`. If found, that is
   the suggested domain.
4. **component fallback** — if no full-domain match, separately find the
   closest second-level part (`SecondLevelThreshold`) and closest top-level
   part (`TopLevelThreshold`). If either part has a closer known value,
   reconstruct `<closestSL>.<closestTL>` as the suggested domain.
5. **dedupe** — only return a `Suggestion` if the suggested domain differs
   from the input domain; otherwise ok=false.

`findClosestDomain(s, list, threshold)` returns the list entry with the
minimum `Distance(s, entry)` that is `<= threshold`, or empty if none
qualifies (ties resolved by first-encountered, matching JS iteration order).

`sift4(a, b)` ported with `maxOffset = 5` (the fork's default), tracking
longest-common-subsequence runs, transpositions, and offset patterns.

## Error handling

No errors are returned; the feature is local-only (no I/O). All failure
modes collapse into `ok=false`. Inputs are never panicked on (guard empty
strings and missing `@`).

## Testing (TDD)

Port mailcheck's Jasmine spec cases to Go table tests:
- `test@gmial.com` → `test@gmail.com` (full domain)
- `test@hotmail.cmo` → `test@hotmail.com` (top-level fix)
- second-level-only typo (e.g. `test@hotmial.com` → `test@hotmail.com`).
  Note: a second-level typo only routes through the component path when no
  *full* domain is within `DomainThreshold`. e.g. `test@yaho.com` resolves
  to `test@aol.com`, not `yahoo.com`, because `aol.com` scores within
  threshold first — this matches mailcheck.js exactly.
- already-valid domain → no suggestion
- empty / no-`@` / no-domain input → ok=false
- input whose only change would equal itself → ok=false

Plus:
- `sift4` unit tests against known string pairs with expected distances.
- A custom `Distance` function is honored.
- Custom `Domains`/threshold lists are honored; defaults are not mutated.
- Defaults are exported and extendable (append doesn't corrupt internal state).

## Decisions made during design

- Field named `Full` (not JS `full`); `Address` added for the local part.
- `encodeEmail` and jQuery/callback API dropped (not meaningful in Go).
- Distance is pluggable via `DistanceFunc`, defaulting to `sift4`.
