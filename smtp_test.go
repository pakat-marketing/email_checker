package emailchecker

import (
	"bufio"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// newFakeServer starts a one-shot TCP listener that plays scripted
// replies. It returns a dialer (matching the package-level `dial`
// signature), and a function returning the commands the client sent
// once the conversation is finished.
func newFakeServer(t *testing.T, replies []string) (func(string, string, time.Duration) (net.Conn, error), func() string) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { listener.Close() })

	var (
		mu       sync.Mutex
		received strings.Builder
		done     = make(chan struct{})
	)

	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		for i, reply := range replies {
			if _, err := conn.Write([]byte(reply + "\r\n")); err != nil {
				return
			}
			if i == len(replies)-1 {
				return
			}
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			mu.Lock()
			received.WriteString(line)
			mu.Unlock()
		}
	}()

	dialer := func(network, address string, timeout time.Duration) (net.Conn, error) {
		if timeout > 0 {
			return net.DialTimeout("tcp", listener.Addr().String(), timeout)
		}
		return net.Dial("tcp", listener.Addr().String())
	}

	return dialer, func() string {
		<-done
		mu.Lock()
		defer mu.Unlock()
		return received.String()
	}
}

func patchDial(t *testing.T, d func(string, string, time.Duration) (net.Conn, error)) {
	t.Helper()
	orig := dial
	dial = d
	t.Cleanup(func() { dial = orig })
}

func TestSMTPAccepts250(t *testing.T) {
	dialer, transcript := newFakeServer(t, []string{
		"220 fake.example.com ready",
		"250 hello",
		"250 sender ok",
		"250 recipient ok",
	})
	patchDial(t, dialer)

	c := &Checker{}
	ok, retry := c.smtpAttempt("user@example.com", "fake.example.com")
	if retry {
		t.Fatal("did not expect retry")
	}
	if !ok {
		t.Fatal("expected RCPT TO to be accepted")
	}

	sent := transcript()
	for _, want := range []string{"HELO example.com", "mail from:<fake@email.com>", "rcpt to:<user@example.com>"} {
		if !strings.Contains(sent, want) {
			t.Errorf("transcript missing %q; got:\n%s", want, sent)
		}
	}
}

func TestSMTPRejects550(t *testing.T) {
	dialer, _ := newFakeServer(t, []string{
		"220 fake.example.com ready",
		"250 hello",
		"250 sender ok",
		"550 no such user",
	})
	patchDial(t, dialer)

	c := &Checker{}
	ok, retry := c.smtpAttempt("user@example.com", "fake.example.com")
	if retry {
		t.Fatal("did not expect retry")
	}
	if ok {
		t.Fatal("expected RCPT TO to be rejected")
	}
}

func TestSMTPRetryOnDialFailure(t *testing.T) {
	attempts := 0
	patchDial(t, func(network, address string, timeout time.Duration) (net.Conn, error) {
		attempts++
		return nil, errors.New("fake dial error")
	})

	c := &Checker{cfg: Config{SMTPRetries: 3}}
	if c.smtpProbe("user@example.com", "fake.example.com") {
		t.Fatal("expected SMTP to fail when every dial errors")
	}
	if attempts != 3 {
		t.Fatalf("expected 3 dial attempts, got %d", attempts)
	}
}

func TestSMTPRespectsHelloAndMailFrom(t *testing.T) {
	dialer, transcript := newFakeServer(t, []string{
		"220 ready",
		"250 hello",
		"250 sender ok",
		"250 recipient ok",
	})
	patchDial(t, dialer)

	c := &Checker{cfg: Config{HelloName: "client.invalid", MailFrom: "probe@client.invalid"}}
	ok, _ := c.smtpAttempt("user@example.com", "fake.example.com")
	if !ok {
		t.Fatal("expected acceptance")
	}
	sent := transcript()
	if !strings.Contains(sent, "HELO client.invalid") {
		t.Errorf("expected custom HELO; got:\n%s", sent)
	}
	if !strings.Contains(sent, "mail from:<probe@client.invalid>") {
		t.Errorf("expected custom MAIL FROM; got:\n%s", sent)
	}
}
