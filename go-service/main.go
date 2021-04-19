package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/propagators/aws/xray"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
)

func main() {
	config, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	tracer, tracerShutdownFunc, err := newTracer(tracerConfig{
		Ctx:           config.Ctx,
		CollectorAddr: config.OtelCollectorAddr,
		SvcName:       config.ServiceName,
	})
	defer tracerShutdownFunc()
	if err != nil {
		log.Fatal(err)
	}

	router := newRouter(config, tracer)

	log.Println(config.ServiceName + " starting...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}

type Config struct {
	Ctx                  context.Context
	ServiceName          string
	OtelCollectorAddr    string
	DependantServiceAddr string
}

func newConfig() (Config, error) {
	ctx := context.Background()
	svcName := os.Getenv("SERVICE_NAME")
	collectorAddr := os.Getenv("COLLECTOR_ADDR")
	callService := os.Getenv("CALL_SERVICE_ADDR")

	return Config{
		Ctx:                  ctx,
		ServiceName:          svcName,
		OtelCollectorAddr:    collectorAddr,
		DependantServiceAddr: callService,
	}, nil
}

func newHttpClient() http.Client {
	return http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}

type tracerConfig struct {
	Ctx           context.Context
	CollectorAddr string
	SvcName       string
}

func newTracer(config tracerConfig) (*sdktrace.TracerProvider, func(), error) {
	exp, err := otlp.NewExporter(config.Ctx,
		otlp.WithInsecure(),
		otlp.WithAddress(config.CollectorAddr),
	)
	if err != nil {
		return nil, nil, err
	}

	res, err := resource.New(config.Ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(config.SvcName),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	idg := xray.NewIDGenerator()
	bsp := sdktrace.NewBatchSpanProcessor(exp)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithIDGenerator(idg),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	shutdownFunc := func() {
		err := exp.Shutdown(config.Ctx)
		if err != nil {
			log.Fatalf("failed to stop exporter: %v", err)
		}
	}

	return tracerProvider, shutdownFunc, nil
}
