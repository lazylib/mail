package mail

import (
	"strings"
	"testing"
)

func TestNewReturnsInstance(t *testing.T) {
	inst := New(Settings{Login: "user", Password: "pass", SMTP: "smtp.example.com"})
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
	if got := encodeSubject("Hello"); got != "Hello" {
		t.Errorf("ascii subject should be unchanged, got %q", got)
	}
	if got := encodeSubject("Привет"); !strings.Contains(strings.ToUpper(got), "=?UTF-8?Q?") {
		t.Errorf("non-ascii subject should be encoded, got %q", got)
	}
}

func TestBuildMessage(t *testing.T) {
	s := Settings{Login: "user@example.com", SMTP: "smtp.example.com"}
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
	if !strings.Contains(text, "\r\n\r\nBody text") {
		t.Errorf("message must use CRLF line endings, got:\n%q", text)
	}
}
