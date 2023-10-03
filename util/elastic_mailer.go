package util

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
)

type ElasticEmailResponse struct {
	Success bool `json:"success"`
	Data    struct {
		TransactionId string `json:"transactionid"`
		MessageId     string `json:"messageid"`
	} `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type ElasticMailer struct {
	apiKey     string
	channel    string
	httpClient *http.Client
}

func NewElasticMailer(apiKey string, channel string) *ElasticMailer {
	httpCookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 1024,
		},
		Jar: httpCookieJar,
	}

	return &ElasticMailer{
		apiKey:     apiKey,
		channel:    channel,
		httpClient: httpClient,
	}
}

func (m *ElasticMailer) SendEmail(fromName string, from string, subject string, to string, bodyHtml string) error {

	apiUrl := "https://api.elasticemail.com/v2/email/send"
	logCtx := log.WithFields(log.Fields{
		"from":    from,
		"to":      to,
		"subject": subject,
		"url":     apiUrl,
	})

	params := url.Values{}
	params.Set("from", from)
	params.Set("fromName", fromName)
	params.Set("apiKey", m.apiKey)
	params.Set("subject", subject)
	params.Set("to", to)
	params.Set("bodyHtml", bodyHtml)
	//params.Set("bodyText", "")
	params.Set("isTransactional", "false")
	params.Set("channel", m.channel)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*15))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", apiUrl, strings.NewReader(params.Encode()))
	if err != nil {
		logCtx.Error(err.Error())
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		logCtx.Error(err.Error())
		return err
	}
	defer resp.Body.Close()

	var result ElasticEmailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logCtx.Error(err.Error())
		return err
	}

	if result.Success {
		logCtx.WithFields(log.Fields{
			"statusCode":    resp.StatusCode,
			"success":       result.Success,
			"transactionId": result.Data.TransactionId,
			"messageId":     result.Data.MessageId,
		}).Info("mail sent")
		return nil
	}

	logCtx.WithFields(log.Fields{
		"statusCode": resp.StatusCode,
		"success":    result.Success,
		"error":      result.Error,
	}).Error("mail not sent")
	return errors.New(result.Error)
}
