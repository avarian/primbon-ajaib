package util

import log "github.com/sirupsen/logrus"

type DummyMessenger struct{}

func NewDummyMessenger() *DummyMessenger {
	return &DummyMessenger{}
}

func (m *DummyMessenger) SendSMS(to string, text string) error {

	log.WithFields(log.Fields{
		"to":   to,
		"text": text,
	}).Info("dummy sms")

	return nil
}
