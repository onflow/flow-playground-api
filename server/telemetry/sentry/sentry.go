package sentry

import (
	"github.com/dapperlabs/flow-playground-api/server/router/gqlHandler/middleware/errors"
	"github.com/getsentry/sentry-go"
	"github.com/kelseyhightower/envconfig"
	"log"
	"time"
)

var sentryConf struct {
	Dsn              string `default:"https://e8ff473e48aa4962b1a518411489ec5d@o114654.ingest.sentry.io/6398442"`
	Debug            bool   `default:"true"`
	AttachStacktrace bool   `default:"true"`
}

func InitializeSentry() {
	if err := envconfig.Process("SENTRY", &sentryConf); err != nil {
		log.Fatal(err)
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryConf.Dsn,
		Debug:            sentryConf.Debug,
		AttachStacktrace: sentryConf.AttachStacktrace,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			if hint.Context != nil {
				if sentryLevel, ok := errors.SentryLogLevel(hint.Context); ok {
					event.Level = sentryLevel
				}
			}
			return event
		},
	})

	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
}

func Cleanup() {
	sentry.Flush(2 * time.Second)
	sentry.Recover()
}
