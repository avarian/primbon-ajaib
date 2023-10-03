package commands

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Postgres
// go get -u gorm.io/driver/postgres
func getPostgresDSN(profile string) string {
	user := viper.GetString(profile + ".user")
	password := viper.GetString(profile + ".password")
	host := viper.GetString(profile + ".host")
	port := viper.GetString(profile + ".port")
	database := viper.GetString(profile + ".database")

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, password, database, port)
}

// MySQL
// go get -u gorm.io/driver/mysql
func getMysqlDSN(profile string) string {
	user := viper.GetString(profile + ".user")
	password := viper.GetString(profile + ".password")
	host := viper.GetString(profile + ".host")
	port := viper.GetString(profile + ".port")
	database := viper.GetString(profile + ".database")
	socket := viper.GetString(profile + ".socket")

	args := url.Values{}
	args.Add("charset", "utf8mb4")
	args.Add("collation", "utf8mb4_unicode_ci")
	args.Add("parseTime", "true")
	args.Add("loc", "Local")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s:%s@", user, password))
	if socket == "" {
		sb.WriteString(fmt.Sprintf("tcp(%s:%s)", host, port))
	} else {
		sb.WriteString(fmt.Sprintf("unix(%s)", socket))
	}
	sb.WriteString(fmt.Sprintf("/%s?%s", database, args.Encode()))

	return sb.String()
}

// Return gorm.DB based by MySQL driver
func newMysqlDB(profile string) *gorm.DB {
	var dbLogMode logger.LogLevel
	if viper.GetBool("verbose") {
		dbLogMode = logger.Info
	} else {
		dbLogMode = logger.Error
	}

	dsn := getMysqlDSN(profile)
	logCtx := log.WithField("dsn", dsn)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(dbLogMode),
	})

	if err != nil {
		logCtx.Fatal(err)
	}
	log.WithField("dsn", dsn).Info("database connected")

	return db
}

// Return gorm.DB based by Postgres driver
func newPostgresDB(profile string) *gorm.DB {
	var dbLogMode logger.LogLevel
	if viper.GetBool("verbose") {
		dbLogMode = logger.Info
	} else {
		dbLogMode = logger.Error
	}

	dsn := getPostgresDSN(profile)
	logCtx := log.WithField("dsn", dsn)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(dbLogMode),
	})

	if err != nil {
		logCtx.Fatal(err)
	}
	log.WithField("dsn", dsn).Info("database connected")

	return db
}

// Return an usable redis client
func newRedisClient(redisUrl string) *redis.Client {

	logCtx := log.WithField("url", redisUrl)

	// Initialize redis client
	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		logCtx.Fatal(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	r := redis.NewClient(opt)
	if _, err := r.Ping(ctx).Result(); err != nil {
		r.Close()
		logCtx.Fatal(err.Error())
	}

	logCtx.Info("redis client connected")
	return r
}

// Return new s3 session
func newS3Session(profile string) *session.Session {
	s3Endpoint := viper.GetString(profile + ".endpoint")
	s3Region := viper.GetString(profile + ".region")
	s3AccessKey := viper.GetString(profile + ".accessKeyId")
	s3SecretAccessKey := viper.GetString(profile + ".secretAccessKey")
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(s3AccessKey, s3SecretAccessKey, ""),
		Endpoint:         aws.String(s3Endpoint),
		Region:           aws.String(s3Region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}

	s3Session, err := session.NewSession(s3Config)

	if err != nil {
		log.Fatal(err.Error())
	}

	log.WithFields(log.Fields{
		"endpoint": viper.GetString(profile + ".endpoint"),
	}).Info("s3 client initialized")

	return s3Session
}
