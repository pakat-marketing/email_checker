## Unreleased

* Add domain typo suggestion (`Suggest`, `NewSuggester`, `SuggestOptions`,
  `Suggestion`), ported from mailcheck.js. Given `kevin@gmial.com` it
  suggests `kevin@gmail.com`. Domain lists, thresholds, and the distance
  function (default `sift4`) are configurable. No network I/O.
* Add `CheckMX(email) (bool, error)` and `(*Checker).CheckMX` so callers
  can distinguish "domain has no MX records" from "DNS lookup failed".
  The original `MX(email) bool` is unchanged.

## v0.1.4 (2021-01-07)

* Fix Compiler Warning (#15)

## v0.1.3 (2019-11-11)

* Fix Compiler Warning (#12)
* Format Code

## v0.1.2 (2018-02-07)

* HELO doesn't work with empty message (#10)
* Format Code using Elixir 1.6

## v0.1.1 (2017-09-29)

* Check Type of email parameter in valid? function

## v0.1.0 (2017-02-27)

* Make Checks configurable
* Remove Elixir < 1.3 Support
* Disable SMTP Check by default
* Timeouts for Checks

## v0.0.3 (2015-11-22)

* Fix a bug when the SMTP doesn't reply
* Fix a bug when the socket is unexpectedly closed

## v0.0.2 (2015-11-21)

* Speedup the library: Remove a retry loop (hack) when MX lookup was empty
* Add a configuration entry for the default DNS used
* Make the app OTP compliant

## v0.0.1 (2015-11-19)

* Check email match a basic email REGEX
* Check the MX exists for the email's domain
* Check the SMTP behind the MX knows this email address
