package util

type Mailer interface {
	SendEmail(fromName string, from string, subject string, to string, bodyHtml string) error
}
