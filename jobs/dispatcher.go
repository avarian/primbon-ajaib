package jobs

import (
	"errors"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/taylorchu/work"
)

// The job namespace is assigned automatically using running binary name
var Namespace string = filepath.Base(os.Args[0])

// The redis queue assigned automatically
var redisQueue work.RedisQueue

func SetRedisQueue(q work.RedisQueue) {
	redisQueue = q
}

type Job interface {
	QueueID() string
}

// Dispatch a job
func Dispatch(j Job) error {

	logCtx := log.WithFields(log.Fields{
		"namespace": Namespace,
		"queueId":   j.QueueID(),
	})

	if redisQueue == nil {
		logCtx.Error("redis queue is uninitialized")
		return errors.New("redis queue is uninitialized")
	}

	job := work.NewJob()
	if err := job.MarshalJSONPayload(j); err != nil {
		logCtx.WithError(err).Error("failed to marshal job")
		return err
	}

	opts := &work.EnqueueOptions{
		Namespace: Namespace,
		QueueID:   j.QueueID(),
	}

	if err := redisQueue.Enqueue(job, opts); err != nil {
		logCtx.WithError(err).Error("failed to dispatch a job")
		return err
	}

	logCtx.Info("job dispatched successfully")
	return nil
}
