package mail

type Instance interface {
	Send(body, subject, to string) error
	SendLetter(letter Letter, subject string, to string) error
}
