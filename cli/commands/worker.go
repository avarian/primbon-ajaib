package commands

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/avarian/primbon-ajaib-backend/jobs"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/taylorchu/work"
	"github.com/taylorchu/work/middleware/discard"
	"github.com/taylorchu/work/middleware/logrus"
)

var (
	workerCmd = &cobra.Command{
		Use:   "worker",
		Short: "Worker command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return workerCommand()
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			rand.Seed(time.Now().UnixNano())
		},
	}
)

func newWorker(client *redis.Client) *work.Worker {
	// Default job options
	maxExecutionTime := time.Duration(viper.GetInt("queue.max_execution_time")) * time.Second
	idleWait := time.Duration(viper.GetInt("queue.idle_wait")) * time.Second
	numGoroutines := viper.GetInt64("queue.num_goroutines")
	maxRetry := viper.GetInt64("queue.max_retry")

	// Initialize worker
	w := work.NewWorker(&work.WorkerOptions{
		Namespace: jobs.Namespace,
		Queue:     work.NewRedisQueue(client),
		ErrorFunc: func(err error) {
			log.WithError(err).Error("redis client error")
		},
	})

	//
	// Register job handlers
	//
	err := w.RegisterWithContext(jobs.ExampleJobQueueId, func(ctx context.Context, j *work.Job, do *work.DequeueOptions) error {
		var example jobs.ExampleJob

		if err := j.UnmarshalJSONPayload(&example); err != nil {
			return err
		}

		if err := example.Handle(ctx); err != nil {
			return err
		}

		return nil
	}, &work.JobOptions{
		MaxExecutionTime: maxExecutionTime,
		IdleWait:         idleWait,
		NumGoroutines:    numGoroutines,
		HandleMiddleware: []work.HandleMiddleware{
			logrus.HandleFuncLogger,
			discard.MaxRetry(maxRetry),
		},
	})

	if err != nil {
		log.WithError(err).Fatal("fail to register queue job handler")
	}

	log.WithFields(log.Fields{
		"namespace":        jobs.Namespace,
		"maxExecutionTime": maxExecutionTime,
		"idleWait":         idleWait,
		"numGoroutines":    numGoroutines,
		"maxRetry":         maxRetry,
	}).Info("worker initialized")
	return w
}

func workerCommand() error {
	// Redis client
	redis := newRedisClient(viper.GetString("redis.url"))
	defer redis.Close()
	log.WithField("url", viper.GetString("redis.url")).Info("redis client initialized")

	w := newWorker(redis)
	w.Start()

	done := make(chan os.Signal, 10)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-done

	log.Info("stopping workers...")
	w.Stop()
	log.Info("all workers stopped")

	return nil
}
