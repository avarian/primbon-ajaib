package util

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
)

type ibDestination struct {
	To string `json:"to"`
}

type ibMessage struct {
	NotifyUrl    string         `json:"notifyUrl"`
	From         string         `json:"from"`
	Destinations *ibDestination `json:"destinations"`
	Text         string         `json:"text"`
}

type ibRequest struct {
	Messages []ibMessage `json:"messages"`
}

type ibResponse struct {
	Messages []ibResponseMessage `json:"messages"`
}

type ibResponseMessage struct {
	To     string `json:"to"`
	Status struct {
		GroupId     int    `json:"groupId"`
		GroupName   string `json:"groupName"`
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"status"`
	MessageId string `json:"messageId"`
}

type InfobipMessenger struct {
	apiKey      string
	callbackUrl string
	sender      string
	httpClient  *http.Client
}

func NewInfobipMessenger(apiKey string, callbackUrl string, sender string) *InfobipMessenger {
	httpCookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 1024,
		},
		Jar: httpCookieJar,
	}

	return &InfobipMessenger{
		apiKey:      apiKey,
		callbackUrl: callbackUrl,
		sender:      sender,
		httpClient:  httpClient,
	}
}

func (m *InfobipMessenger) SendSMS(to string, text string) error {

	apiUrl := "https://zj1jq3.api.infobip.com/sms/2/text/advanced"
	logCtx := log.WithFields(log.Fields{
		"from":     m.sender,
		"to":       to,
		"text":     text,
		"url":      apiUrl,
		"callback": m.callbackUrl,
	})

	destination := &ibDestination{
		To: to,
	}

	message := &ibMessage{
		NotifyUrl:    m.callbackUrl,
		From:         m.sender,
		Destinations: destination,
		Text:         text,
	}

	request := &ibRequest{
		Messages: []ibMessage{*message},
	}

	json, err := json.Marshal(request)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*15))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", apiUrl, bytes.NewBuffer(json))
	if err != nil {
		logCtx.Error(err.Error())
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "App "+m.apiKey)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		logCtx.Error(err.Error())
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logCtx.Error(err.Error())
		return err
	}

	logCtx.WithFields(log.Fields{
		"statusCode": resp.StatusCode,
		"response":   string(body),
	}).Info("sms sent")

	return nil
}
