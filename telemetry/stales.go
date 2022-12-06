package telemetry

import (
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/server/config"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"time"
)

// staleDuration is the amount a time before a project is considered stale if not accessed
var staleDuration = (time.Hour * 24) * time.Duration(config.Playground().StaleProjectDays)

// staleProjectScanner scans for stale projects in the database
var staleProjectScanner func(stale time.Duration, projs *[]*model.Project) error = nil

func SetStaleProjectScanner(scanner func(stale time.Duration, projs *[]*model.Project) error) {
	staleProjectScanner = scanner
}

// StaleProjectCounter returns the number of stale projects
func StaleProjectCounter() float64 {
	if staleProjectScanner == nil {
		sentry.CaptureException(errors.New("stale project scanner not set"))
		return 0
	}

	var stales []*model.Project
	err := staleProjectScanner(staleDuration, &stales)
	if err != nil {
		sentry.CaptureException(errors.Wrap(err, "failed to get stale projects"))
		return 0
	}

	return float64(len(stales))
}
