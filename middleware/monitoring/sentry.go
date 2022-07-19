package monitoring

import (
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/gorilla/mux"
)

func Middleware() mux.MiddlewareFunc {
	return sentryhttp.New(sentryhttp.Options{}).Handle
}
