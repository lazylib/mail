package mail

import (
	"strings"
	"testing"
)

func TestNewReturnsInstance(t *testing.T) {
	inst := New(Settings{Auth: &Auth{Login: "user", Password: "pass"}, SMTP: "smtp.example.com"})
	if _, ok := inst.(config); !ok {
		t.Fatalf("New() did not return a config instance")
	}
}

func TestRenderLetter(t *testing.T) {
	letter := Letter{
		HtmlTemplate: "<h1>Hello, {{.Name}}!</h1><p>{{.Count}}</p>",
		Props:        map[string]any{"Name": "World", "Count": 42},
	}
	out, err := renderLetter(letter)
	if err != nil {
		t.Fatalf("renderLetter: %v", err)
	}
	if !strings.Contains(out, "Hello, World!") {
		t.Errorf("missing rendered name, got: %q", out)
	}
	if !strings.Contains(out, "42") {
		t.Errorf("missing rendered count, got: %q", out)
	}
}

func TestRenderLetterEscapesHTML(t *testing.T) {
	letter := Letter{
		HtmlTemplate: "<p>{{.}}</p>",
		Props:        "<script>alert(1)</script>",
	}
	out, err := renderLetter(letter)
	if err != nil {
		t.Fatalf("renderLetter: %v", err)
	}
	if strings.Contains(out, "<script>") {
		t.Errorf("props not escaped, got: %q", out)
	}
}

func TestRenderLetterBadTemplate(t *testing.T) {
	_, err := renderLetter(Letter{HtmlTemplate: "{{.Missing"})
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}

func TestEncodeSubject(t *testing.T) {
	if got := encodeWord("Hello"); got != "Hello" {
		t.Errorf("ascii subject should be unchanged, got %q", got)
	}
	if got := encodeWord("Привет"); !strings.Contains(strings.ToUpper(got), "=?UTF-8?Q?") {
		t.Errorf("non-ascii subject should be encoded, got %q", got)
	}
}

func TestEncodeWordASCII(t *testing.T) {
	if got := encodeWord("Lazy Mail Team"); got != "Lazy Mail Team" {
		t.Errorf("ascii word should be unchanged, got %q", got)
	}
}

func TestBuildMessage(t *testing.T) {
	s := Settings{Auth: &Auth{Login: "user@example.com"}, SMTP: "smtp.example.com"}
	msg := buildMessage(s, "to@example.com", "Hello", "Body text", "text/plain; charset=UTF-8")
	text := string(msg)

	for _, want := range []string{
		"From: user@example.com",
		"To: to@example.com",
		"Subject: Hello",
		"Content-Type: text/plain; charset=UTF-8",
		"\r\n\r\nBody text",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("message missing %q, got:\n%s", want, text)
		}
	}
}

func TestBuildMessageWithName(t *testing.T) {
	s := Settings{Auth: &Auth{Login: "user@example.com"}, SMTP: "smtp.example.com", Name: "Lazy Mail"}
	msg := buildMessage(s, "to@example.com", "Hello", "Body", "text/plain; charset=UTF-8")
	text := string(msg)

	if !strings.Contains(text, "From: Lazy Mail <user@example.com>") {
		t.Errorf("From header should contain display name, got:\n%s", text)
	}
}

func TestBuildMessageWithFrom(t *testing.T) {
	s := Settings{Auth: &Auth{Login: "user@example.com"}, From: "sender@example.com", SMTP: "smtp.example.com"}
	msg := buildMessage(s, "to@example.com", "Hello", "Body", "text/plain; charset=UTF-8")
	text := string(msg)

	if !strings.Contains(text, "From: sender@example.com") {
		t.Errorf("From should take precedence over Login, got:\n%s", text)
	}
}
