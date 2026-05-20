package emailchecker

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	defaultSMTPRetries = 2
	defaultMailFrom    = "fake@email.com"
)

// SMTP reports whether the SMTP server for the email's domain
// acknowledges email as a recipient (a 250 reply to RCPT TO). No mail is
// sent. Most providers no longer support this probe; use with care.
func SMTP(email string) bool {
	return (&Checker{}).SMTP(email)
}

// SMTP runs the SMTP RCPT TO probe against email, retrying up to
// SMTPRetries times on network errors.
func (c *Checker) SMTP(email string) bool {
	host, err := c.lookupMX(email)
	if err != nil || host == "" {
		return false
	}
	return c.smtpProbe(email, host)
}

func (c *Checker) smtpProbe(email, host string) bool {
	retries := c.cfg.SMTPRetries
	if retries <= 0 {
		retries = defaultSMTPRetries
	}
	for i := 0; i < retries; i++ {
		ok, retry := c.smtpAttempt(email, host)
		if !retry {
			return ok
		}
	}
	return false
}

// smtpAttempt returns (accepted, retry). retry is true when a network
// error suggests trying again; accepted is the RCPT TO outcome.
func (c *Checker) smtpAttempt(email, host string) (accepted bool, retry bool) {
	domain := Domain(email)
	timeout := c.dialTimeout()

	conn, err := dial("tcp", net.JoinHostPort(host, "25"), timeout)
	if err != nil {
		return false, true
	}
	defer conn.Close()

	if timeout > 0 {
		_ = conn.SetDeadline(time.Now().Add(timeout))
	}

	r := bufio.NewReader(conn)

	if _, err := readLine(r); err != nil {
		return false, true
	}

	hello := c.cfg.HelloName
	if hello == "" {
		hello = domain
	}
	mailFrom := c.cfg.MailFrom
	if mailFrom == "" {
		mailFrom = defaultMailFrom
	}

	if err := writeCmd(conn, "HELO "+hello); err != nil {
		return false, true
	}
	if _, err := readLine(r); err != nil {
		return false, true
	}

	if err := writeCmd(conn, "mail from:<"+mailFrom+">"); err != nil {
		return false, true
	}
	if _, err := readLine(r); err != nil {
		return false, true
	}

	if err := writeCmd(conn, "rcpt to:<"+email+">"); err != nil {
		return false, true
	}
	line, err := readLine(r)
	if err != nil {
		return false, true
	}

	return strings.HasPrefix(line, "250 "), false
}

// dial is a package-level indirection so tests can substitute a fake
// dialer.
var dial = func(network, address string, timeout time.Duration) (net.Conn, error) {
	if timeout > 0 {
		return net.DialTimeout(network, address, timeout)
	}
	return net.Dial(network, address)
}

func writeCmd(w net.Conn, cmd string) error {
	_, err := fmt.Fprintf(w, "%s\r\n", cmd)
	return err
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	return strings.TrimRight(line, "\r\n"), err
}
