package config

import (
	"github.com/kelseyhightower/envconfig"
	"log"
)

// GetSentryConfig parses environment variables and returns a copy of the config
func GetSentryConfig() SentryConfig {
	if sentryConfig == nil {
		sentryConfig = &SentryConfig{}
		if err := envconfig.Process("SENTRY", sentryConfig); err != nil {
			log.Fatal(err)
		}
	}
	return *sentryConfig
}

var sentryConfig *SentryConfig = nil

type SentryConfig struct {
	Dsn              string `default:"https://e8ff473e48aa4962b1a518411489ec5d@o114654.ingest.sentry.io/6398442"`
	Debug            bool   `default:"true"`
	AttachStacktrace bool   `default:"true"`
}
