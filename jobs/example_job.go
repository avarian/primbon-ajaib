package jobs

import (
	"context"

	log "github.com/sirupsen/logrus"
)

var ExampleJobQueueId = "example"

type ExampleJob struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Subject     string `json:"subject"`
	ContentText string `json:"content_text"`
}

func NewExampleJob(from string, to string, subject string, contentText string) *ExampleJob {
	return &ExampleJob{
		From:        from,
		To:          to,
		Subject:     subject,
		ContentText: contentText,
	}
}

// Return the queue id for this job
func (j *ExampleJob) QueueID() string { return ExampleJobQueueId }

// Define job handle function, how the job is executed
// This will get executed inside a work.ContextHandleFunc
func (j *ExampleJob) Handle(ctx context.Context) error {
	log.WithFields(log.Fields{
		"From":        j.From,
		"To":          j.To,
		"Subject":     j.Subject,
		"ContentText": j.ContentText,
	}).Info("Processing ExampleJob.Handle()")
	return nil
}
