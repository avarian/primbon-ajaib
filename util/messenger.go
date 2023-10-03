package util

type Messenger interface {
	SendSMS(to string, text string) error
}
