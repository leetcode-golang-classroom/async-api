package config

import (
	"log"

	"github.com/spf13/viper"
)

type Env string

const (
	Env_Test Env = "test"
	Env_Dev  Env = "dev"
)

type Config struct {
	Port                 string `mapstructure:"PORT"`
	DBURL                string `mapstructure:"DB_URL"`
	DBURLTEST            string `mapstructure:"DB_URL_TEST"`
	ENV                  Env    `mapstructure:"ENV"`
	PROJECT_ROOT         string `mapstructure:"PROJECT_ROOT"`
	JWTSecret            string `mapstructure:"JWT_SECRET"`
	JWTServerHost        string `mapstructure:"JWT_SERVER_HOST"`
	S3LocalstackEndPoint string `mapstructure:"S3_LOCALSTACK_ENDPOINT"`
	LocalstackEndPoint   string `mapstructure:"LOCALSTACK_ENDPOINT"`
	S3Bucket             string `mapstructure:"S3_BUCKET"`
	SQSQueue             string `mapstructure:"SQS_QUEUE"`
}

var AppConfig *Config

func init() {
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	FailOnError(v.BindEnv("PORT"), "failed to bind PORT")
	FailOnError(v.BindEnv("DB_URL"), "failed to bind DB_URL")
	FailOnError(v.BindEnv("DB_URL_TEST"), "failed to bind DB_URL_TEST")
	FailOnError(v.BindEnv("ENV"), "failed to bind ENV")
	FailOnError(v.BindEnv("PROJECT_ROOT"), "failed to bind PROJECT_ROOT")
	FailOnError(v.BindEnv("JWT_SECRET"), "failed to bind JWT_SECRET")
	FailOnError(v.BindEnv("JWT_SERVER_HOST"), "failed to bind JWT_SERVER_HOST")
	FailOnError(v.BindEnv("S3_LOCALSTACK_ENDPOINT"), "faield to bind S3_LOCALSTACK_ENDPOINT")
	FailOnError(v.BindEnv("LOCALSTACK_ENDPOINT"), "faield to bind LOCALSTACK_ENDPOINT")
	FailOnError(v.BindEnv("S3_BUCKET"), "faield to bind S3_BUCKET")
	FailOnError(v.BindEnv("SQS_QUEUE"), "faield to bind SQS_QUEUE")
	err := v.ReadInConfig()
	if err != nil {
		log.Println("Load from environment variable")
	}
	err = v.Unmarshal(&AppConfig)
	if err != nil {
		FailOnError(err, "Failed to read enivronment")
	}
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func (c *Config) SetupEnv(env Env) {
	c.ENV = env
}
