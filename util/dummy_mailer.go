package util

import log "github.com/sirupsen/logrus"

type DummyMailer struct{}

func NewDummyMailer() *DummyMailer {
	return &DummyMailer{}
}

func (m *DummyMailer) SendEmail(fromName string, from string, subject string, to string, bodyHtml string) error {

	log.WithFields(log.Fields{
		"fromName": fromName,
		"from":     from,
		"subject":  subject,
		"to":       to,
	}).Info("dummy mail")

	return nil
}
