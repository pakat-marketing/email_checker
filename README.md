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
```

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
