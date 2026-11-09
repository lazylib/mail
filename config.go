package mail

type Auth struct {
	Login    string // SMTP username used for authentication
	Password string // SMTP password used for authentication
}

// Settings holds everything required to initialize the mail library
// and send emails through the configured SMTP server.
type Settings struct {
	SMTP string // SMTP server host, e.g. "smtp.gmail.com"
	Port int    // SMTP server port: 587 (STARTTLS) or 465 (implicit TLS). 0 defaults to 587.
	From string // sender address used in the From header, e.g. "sender@example.com"
	Name string // optional sender display name: "From: Name <sender@example.com>"
	Auth *Auth
}

type config struct {
	settings Settings
}

// New initializes the mail library with the given settings
// and returns an Instance ready to send emails.
func New(settings Settings) Instance {
	return config{
		settings: settings,
	}
}
