package commands

import (
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	appFilename = "primbon-ajaib-backend" // filepath.Base(os.Args[0])
	configFile  = ""

	rootCmd = &cobra.Command{
		Use:     appFilename,
		Version: "1.0",
		Short:   "Go App Skeleton",
		Long: `Go App Skeletonn

This is go app skeleton applications`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func initConfig() error {
	// load custom config file if specified via command flags
	if configFile != "" {
		abs, err := filepath.Abs(configFile)
		if err != nil {
			return err
		}
		base := filepath.Base(abs)
		path := filepath.Dir(abs)
		viper.SetConfigName(strings.Split(base, ".")[0])
		viper.AddConfigPath(path)
	} else {
		viper.SetConfigName(appFilename)
		viper.AddConfigPath(".")
	}

	// from here, able use viper to get configuration vars
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	// enable reading from environment
	viper.SetEnvPrefix(strings.ToUpper(appFilename))
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	if viper.GetBool("verbose") {
		log.SetLevel(log.DebugLevel)
	}

	logFile := viper.GetString("log")
	if logFile != "" {
		switch logFile {
		case "stdout":
			log.SetOutput(os.Stdout)
		case "stderr":
			log.SetOutput(os.Stderr)
		default:
			log.SetOutput(&lumberjack.Logger{
				Filename: logFile,
				MaxSize:  100,
				MaxAge:   30,
			})
		}
	}

	return nil
}
