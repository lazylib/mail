# mail

Simple SMTP mail library for Go. Initialize the library with a config and send plain text or HTML emails — the HTML email is built from a template with dynamic properties.

## Install

```
go get github.com/lazylib/mail
```

## Usage

```go
package main

import (
	"github.com/lazylib/mail"
)

func main() {
	inst := mail.New(mail.Settings{
		Login:    "sender@example.com",   // SMTP username
		Password: "secret",               // SMTP password
		SMTP:     "smtp.example.com",     // SMTP server host
		Port:     587,                    // 587 = STARTTLS, 465 = implicit TLS (0 defaults to 587)
		From:     "sender@example.com",   // From header (defaults to Login if empty)
	})

	// Plain text email
	if err := inst.Send("Hello from lazylib!", "Greetings", "to@example.com"); err != nil {
		panic(err)
	}

	// HTML email rendered from a template with dynamic props
	err := inst.SendLetter(mail.Letter{
		HtmlTemplate: `<h1>Hello, {{.Name}}!</h1><p>Your order #{{.OrderID}} is shipped.</p>`,
		Props: map[string]any{
			"Name":    "Alice",
			"OrderID": 42,
		},
	}, "Order update", "alice@example.com")
	if err != nil {
		panic(err)
	}
}
```

## Settings

| Field    | Description                                                        |
| -------- | ------------------------------------------------------------------ |
| Login    | SMTP username used for authentication                              |
| Password | SMTP password used for authentication                              |
| SMTP     | SMTP server host, e.g. `smtp.gmail.com`                            |
| Port     | SMTP port: `587` (STARTTLS) or `465` (implicit TLS). `0` defaults to `587` |
| From     | Sender address for the `From` header. Defaults to `Login` if empty |

## API

- `mail.New(settings Settings) Instance` — initialize the library.
- `(Instance) Send(body, subject, to string) error` — send a plain text email.
- `(Instance) SendLetter(letter Letter, subject string, to string) error` — render the HTML template with `Props` and send it.
