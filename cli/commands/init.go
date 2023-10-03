package commands

import (
	"github.com/spf13/viper"
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", configFile, "path to config file")
	rootCmd.PersistentFlags().String("log", "", "log output destination (stdout, stderr, or filepath)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output for debugging")
	rootCmd.PersistentFlags().String("redis", "redis://redis:6379", "redis connection url")
	viper.BindPFlag("log", rootCmd.PersistentFlags().Lookup("log"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("redis.url", rootCmd.Flags().Lookup("redis"))

	// Command flags for "serve"
	serveCmd.Flags().String("listen", ":8080", "http server listen address")
	viper.BindPFlag("listen_address", serveCmd.Flags().Lookup("listen"))

	// Command flags for "queue"
	//queueCmd.Flags().BoolVar(&queueWorker, "worker", false, "run queue worker (default: "+strconv.FormatBool(queueWorker)+")")
	workerCmd.Flags().Int("num-goroutines", 4, "number of goroutines")
	workerCmd.Flags().Int("max-execution-time", 180, "max execution time in seconds")
	workerCmd.Flags().Int("idle-wait", 10, "idle wait in seconds")
	workerCmd.Flags().Int("max-retry", 10, "max retrying attempt before discarding a job")
	viper.BindPFlag("queue.num_goroutines", workerCmd.Flags().Lookup("num-goroutines"))
	viper.BindPFlag("queue.max_execution_time", workerCmd.Flags().Lookup("max-execution-time"))
	viper.BindPFlag("queue.idle_wait", workerCmd.Flags().Lookup("idle-wait"))
	viper.BindPFlag("queue.max_retry", workerCmd.Flags().Lookup("max-retry"))

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(workerCmd)
	rootCmd.AddCommand(migrateCmd)
}
