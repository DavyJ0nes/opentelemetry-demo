package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
)

func main() {
	config, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	teleConfig, err := newTelemetryConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := teleConfig.Exporter.Shutdown(config.Ctx)
		if err != nil {
			log.Fatalf("failed to stop exporter: %v", err)
		}
	}()

	tracer, err := newTracer(teleConfig)
	if err != nil {
		log.Fatal(err)
	}

	metricProcessor, err := newMetricProcessor(teleConfig)
	if err != nil {
		log.Fatal(err)
	}

	router := newRouter(config, tracer, metricProcessor)

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

type telemetryConfig struct {
	Ctx      context.Context
	Exporter *otlp.Exporter
	Resource *resource.Resource
	SvcName  string
}

func newTelemetryConfig(config Config) (telemetryConfig, error) {
	driver := otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(config.OtelCollectorAddr),
	)

	exp, err := otlp.NewExporter(config.Ctx, driver)
	if err != nil {
		return telemetryConfig{}, err
	}

	res, err := resource.New(config.Ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(config.ServiceName),
		),
	)
	if err != nil {
		return telemetryConfig{}, err
	}

	return telemetryConfig{
		Ctx:      config.Ctx,
		Exporter: exp,
		Resource: res,
	}, nil
}

func newTracer(config telemetryConfig) (*sdktrace.TracerProvider, error) {
	idg := xray.NewIDGenerator()
	bsp := sdktrace.NewBatchSpanProcessor(config.Exporter)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(config.Resource),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithIDGenerator(idg),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tracerProvider, nil
}

func newMetricProcessor(config telemetryConfig) (*controller.Controller, error) {
	cont := controller.New(
		processor.New(
			simple.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries([]float64{
					0.001, 0.01, 0.1, 1, 10, 100, 1000,
				}),
			),
			config.Exporter,
			processor.WithMemory(true),
		),
		controller.WithResource(config.Resource),
		controller.WithExporter(config.Exporter),
	)

	if err := cont.Start(config.Ctx); err != nil {
		return nil, err
	}

	return cont, nil
}
