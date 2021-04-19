package main

import (
	"net/http"

	"github.com/gorilla/mux"
	muxmiddleware "go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func newRouter(config Config, tracer *sdktrace.TracerProvider, metricProcessor *controller.Controller) http.Handler {
	meter := metricProcessor.MeterProvider().Meter("example.test")
	counter := metric.Must(meter).NewInt64Counter(
		"http_request_count_total",
		metric.WithDescription("total count of http requests"),
		metric.WithInstrumentationVersion("v1.1.0"),
	)
	m := &measurement{
		counter:     counter,
		serviceName: config.ServiceName,
	}

	router := mux.NewRouter()
	router.Use(
		muxmiddleware.Middleware(config.ServiceName, muxmiddleware.WithTracerProvider(tracer)),
		m.Measure,
	)

	router.Handle("/", index{
		dependantService: config.DependantServiceAddr,
		httpClient:       newHttpClient(),
		serviceName:      config.ServiceName,
	})

	return router
}

func newHttpClient() http.Client {
	return http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}

type measurement struct {
	counter     metric.Int64Counter
	serviceName string
}

func (m *measurement) Measure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		labels := []attribute.KeyValue{
			attribute.String("service_name", m.serviceName),
			attribute.String("path", req.URL.Path),
		}

		m.counter.Add(req.Context(), 1, labels...)

		next.ServeHTTP(w, req)
	})
}
