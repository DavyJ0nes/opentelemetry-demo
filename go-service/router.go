package main

import (
	"net/http"

	"github.com/gorilla/mux"
	telemetry "github.com/iZettle/golang-telemetry"
	muxmiddleware "go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func newRouter(config Config, tracer *sdktrace.TracerProvider) http.Handler {
	tele := telemetry.New(config.ServiceName)
	middleware := tele.NewMiddleware()

	router := mux.NewRouter()
	router.Use(
		muxmiddleware.Middleware(config.ServiceName, muxmiddleware.WithTracerProvider(tracer)),
		middleware.Measure,
	)

	telemetry.RegisterHandler(tele, router)

	router.Handle("/", index{
		dependantService: config.DependantServiceAddr,
		httpClient:       newHttpClient(),
		serviceName:      config.ServiceName,
	})

	return router
}
