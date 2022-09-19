package router

import (
	gqlPlayground "github.com/99designs/gqlgen/graphql/playground"
	"github.com/dapperlabs/flow-playground-api/server/config"
	playground "github.com/dapperlabs/flow-playground-api/server/router/gqlHandler"
	"github.com/dapperlabs/flow-playground-api/server/router/gqlHandler/resolver/controller"
	"github.com/dapperlabs/flow-playground-api/server/router/middleware/httpcontext"
	"github.com/dapperlabs/flow-playground-api/server/router/middleware/monitoring"
	"github.com/dapperlabs/flow-playground-api/server/router/middleware/sessions"
	"github.com/dapperlabs/flow-playground-api/server/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/httplog"
	"github.com/go-chi/render"
	gsessions "github.com/gorilla/sessions"
	"github.com/rs/cors"
	"net/http"
)

func InitializeRouter() *chi.Mux {
	router := chi.NewRouter()
	router.Use(monitoring.Middleware())

	if config.GetConfig().Debug {
		logger := httplog.NewLogger("playground-api", httplog.Options{Concise: true})
		router.Use(httplog.RequestLogger(logger))
		router.Handle("/", gqlPlayground.Handler("GraphQL playground", "/query"))
	}

	router.Route("/query", func(r chi.Router) {
		// Add CORS middleware around every request
		// See https://github.com/rs/cors for full option listing
		r.Use(cors.New(cors.Options{
			AllowedOrigins:   config.GetConfig().AllowedOrigins,
			AllowCredentials: true,
		}).Handler)

		sessionAuthKey := []byte(config.GetConfig().SessionAuthKey)
		cookieStore := gsessions.NewCookieStore(sessionAuthKey)
		cookieStore.MaxAge(int(config.GetConfig().SessionMaxAge.Seconds()))

		cookieStore.Options.Secure = config.GetConfig().SessionCookiesSecure
		cookieStore.Options.HttpOnly = config.GetConfig().SessionCookiesHTTPOnly

		if config.GetConfig().SessionCookiesSameSiteNone {
			cookieStore.Options.SameSite = http.SameSiteNoneMode
		}

		r.Use(httpcontext.Middleware())
		r.Use(sessions.Middleware(cookieStore))
		r.Use(monitoring.Middleware())

		gqlHandler := playground.GraphQLHandler()
		r.Handle("/", gqlHandler)
	})

	embedsHandler := controller.NewEmbedsHandler(storage.GetStorage(), config.GetConfig().PlaygroundBaseURL)
	router.Handle("/embed", embedsHandler)

	utilsHandler := controller.NewUtilsHandler()
	router.Route("/utils", func(r chi.Router) {
		// Add CORS middleware around every request
		// See https://github.com/rs/cors for full option listing
		r.Use(cors.New(cors.Options{
			AllowedOrigins: config.GetConfig().AllowedOrigins,
		}).Handler)

		r.Use(render.SetContentType(render.ContentTypeJSON))
		r.HandleFunc("/version", utilsHandler.VersionHandler)
	})

	router.HandleFunc("/ping", ping)
	return router
}

func ping(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(200)
	_, _ = w.Write([]byte("ok"))
}
