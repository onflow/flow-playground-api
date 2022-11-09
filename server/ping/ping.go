package ping

import (
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"net/http"
)

// handlers holds ping handler functions for /ping endpoint to call
var handlers struct {
	initialized bool
	storagePing func() error
	// Add other handlers here
}

// SetPingHandlers sets all ping handlers functions
func SetPingHandlers(storagePing func() error) error {
	if storagePing == nil {
		return errors.New("storage ping handler is nil")
	}

	handlers.storagePing = storagePing
	handlers.initialized = true
	return nil
}

// Ping handles /ping endpoint
//
// Calls each handler in ping handlers
func Ping(w http.ResponseWriter, _ *http.Request) {
	if !handlers.initialized {
		w.WriteHeader(http.StatusInternalServerError)
		sentry.CaptureException(errors.New("unset ping handlers"))
		return
	}

	if err := handlers.storagePing(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sentry.CaptureException(errors.Wrap(err, "database ping failed"))
		return
	}

	w.WriteHeader(http.StatusOK)
}
