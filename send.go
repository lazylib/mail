package mail

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
	"unicode"
)

// Letter is an email built from an HTML template.
// Props is passed to the template renderer, so any data
// can be injected into the HtmlTemplate dynamically.
type Letter struct {
	HtmlTemplate string
	Props        any
}

// Send sends a plain text email to the given recipient.
func (c config) Send(body, subject, to string) error {
	msg := buildMessage(c.settings, to, subject, body, "text/plain; charset=UTF-8")
	return c.send(msg, to)
}

// SendLetter renders letter.HtmlTemplate with letter.Props
// and sends the resulting HTML email to the given recipient.
func (c config) SendLetter(letter Letter, subject string, to string) error {
	htmlBody, err := renderLetter(letter)
	if err != nil {
		return err
	}
	msg := buildMessage(c.settings, to, subject, htmlBody, "text/html; charset=UTF-8")
	return c.send(msg, to)
}

// renderLetter parses the HTML template and executes it with Props.
func renderLetter(letter Letter) (string, error) {
	tmpl, err := template.New("letter").Parse(letter.HtmlTemplate)
	if err != nil {
		return "", fmt.Errorf("mail: parse template: %w", err)
	}
	var out strings.Builder
	if err := tmpl.Execute(&out, letter.Props); err != nil {
		return "", fmt.Errorf("mail: execute template: %w", err)
	}
	return out.String(), nil
}

// buildMessage assembles an RFC 5322 message with CRLF line endings.
func buildMessage(s Settings, to, subject, body, contentType string) []byte {
	from := s.From
	if from == "" && s.Auth != nil {
		from = s.Auth.Login
	}
	if s.Name != "" {
		from = encodeWord(s.Name) + " <" + from + ">"
	}
	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + encodeWord(subject) + "\r\n" +
		"Date: " + time.Now().Format(time.RFC1123Z) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: " + contentType + "\r\n" +
		"\r\n" +
		body
	return []byte(msg)
}

// encodeWord encodes non-ASCII strings using RFC 2047 word encoding,
// required by the SMTP protocol for international characters.
func encodeWord(s string) string {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return mime.QEncoding.Encode("utf-8", s)
		}
	}
	return s
}

// send delivers msg to the recipient via the configured SMTP server.
// Port 465 uses implicit TLS, any other port uses STARTTLS when available.
func (c config) send(msg []byte, to string) error {
	s := c.settings
	if s.SMTP == "" {
		return fmt.Errorf("mail: SMTP host is not set")
	}
	if s.From == "" && (s.Auth == nil || s.Auth.Login == "") {
		return fmt.Errorf("mail: sender address is not set (Settings.From or Settings.Auth.Login)")
	}

	port := s.Port
	if port == 0 {
		port = 587
	}
	addr := net.JoinHostPort(s.SMTP, fmt.Sprintf("%d", port))

	client, err := dial(s.SMTP, addr, port)
	if err != nil {
		return err
	}
	defer client.Close()

	if s.Auth != nil {
		auth := smtp.PlainAuth("", s.Auth.Login, s.Auth.Password, s.SMTP)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("mail: auth: %w", err)
		}
	}

	if err := client.Mail(s.From); err != nil {
		return fmt.Errorf("mail: from %q: %w", s.From, err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("mail: to %q: %w", to, err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("mail: data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("mail: write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("mail: close: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("mail: quit: %w", err)
	}
	return nil
}

// dial establishes an SMTP connection, using implicit TLS for port 465
// and STARTTLS for any other port when the server supports it.
func dial(host, addr string, port int) (*smtp.Client, error) {
	if port == 465 {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
		if err != nil {
			return nil, fmt.Errorf("mail: tls dial: %w", err)
		}
		client, err := smtp.NewClient(conn, host)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("mail: smtp client: %w", err)
		}
		return client, nil
	}

	client, err := smtp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("mail: dial: %w", err)
	}
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			client.Close()
			return nil, fmt.Errorf("mail: starttls: %w", err)
		}
	}
	return client, nil
}
