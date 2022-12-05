package telemetry

import (
	"fmt"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/server/config"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"time"
)

// staleDuration is the amount a time before a project is considered stale if not accessed
var staleDuration = (time.Hour * 24) * time.Duration(config.Playground().StaleProjectDays)

// staleProjectScanner scans for stale projects in the database
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

	var stales []*model.Project
	err := staleProjectScanner(staleDuration, &stales)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sentry.CaptureException(errors.Wrap(err, "failed to get stale project count"))
		return
	}

	staleDurationDays := staleDuration.Hours() / 24

	_, _ = w.Write([]byte(fmt.Sprintf("stale_project_duration_days %s\n",
		strconv.FormatFloat(staleDurationDays, 'f', -1, 64))))

	_, _ = w.Write([]byte(fmt.Sprintf("stale_project_count %d\n", len(stales))))

	_, _ = w.Write([]byte("stale_project_ids"))
	if (len(stales)) == 0 {
		_, _ = w.Write([]byte(" none"))
	}
	for _, proj := range stales {
		_, _ = w.Write([]byte(fmt.Sprintf("\n  %s", proj.ID.String())))
	}

	w.WriteHeader(http.StatusOK)
}
