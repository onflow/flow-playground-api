package telemetry

import (
	"fmt"
	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

// staleProjectCounter returns the number of stale projects in the database
var staleProjectScanner func(stale time.Duration, projs *[]*model.Project) error = nil

func SetStaleProjectScanner(scanner func(stale time.Duration, projs *[]*model.Project) error) {
	staleProjectScanner = scanner
}

func StaleProjects(w http.ResponseWriter, _ *http.Request) {
	if staleProjectScanner == nil {
		w.WriteHeader(http.StatusInternalServerError)
		sentry.CaptureException(errors.New("stale project scanner not set"))
		return
	}

	// TODO: Display stale project count then IDs

	var stales []*model.Project
	err := staleProjectScanner(controller.StaleDuration, &stales)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sentry.CaptureException(errors.Wrap(err, "failed to get stale project count"))
		return
	}

	_, _ = w.Write([]byte(fmt.Sprintf("StaleProjectCount: %d\n", len(stales))))
	_, _ = w.Write([]byte("StaleProjectIDs:\n"))
	for _, proj := range stales {
		_, _ = w.Write([]byte(fmt.Sprintf("  %s\n", proj.ID.String())))
	}
	w.WriteHeader(http.StatusOK)
}
