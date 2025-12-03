package observability

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Logger 日志接口（最小化依赖）
type Logger interface {
	Info(ctx context.Context, msg string, fields ...any)
	Error(ctx context.Context, msg string, fields ...any)
}

// Observability 可观测性管理器
type Observability struct {
	cfg      Config
	tp       *sdktrace.TracerProvider
	registry *prometheus.Registry
	metrics  MetricsProvider
	logger   Logger
}

// Option 可观测性选项
type Option func(*Observability)

// WithLogger 设置日志
func WithLogger(logger Logger) Option {
	return func(o *Observability) {
		o.logger = logger
	}
}

// WithRegistry 设置自定义 Prometheus Registry
func WithRegistry(registry *prometheus.Registry) Option {
	return func(o *Observability) {
		o.registry = registry
	}
}

// New 创建可观测性实例
func New(cfg Config, opts ...Option) (*Observability, error) {
	o := &Observability{cfg: cfg}

	for _, opt := range opts {
		opt(o)
	}

	if err := o.initTracing(); err != nil {
		return nil, fmt.Errorf("init tracing: %w", err)
	}

	o.initMetrics()

	return o, nil
}

func (o *Observability) initTracing() error {
	if !o.cfg.Tracing.Enabled {
		o.tp = sdktrace.NewTracerProvider()
		otel.SetTracerProvider(o.tp)
		return nil
	}

	exporter, err := o.createExporter()
	if err != nil {
		return err
	}

	res, err := o.createResource()
	if err != nil {
		return err
	}

	o.tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(o.createSampler()),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(o.tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

func (o *Observability) createExporter() (sdktrace.SpanExporter, error) {
	ctx := context.Background()
	cfg := o.cfg.Tracing

	switch cfg.Exporter {
	case "otlp", "otlp-grpc":
		opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.Endpoint)}
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		return otlptracegrpc.New(ctx, opts...)

	case "otlp-http":
		opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(cfg.Endpoint)}
		if cfg.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		return otlptracehttp.New(ctx, opts...)

	case "stdout":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())

	default:
		return &noopExporter{}, nil
	}
}

func (o *Observability) createResource() (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(o.cfg.ServiceName),
		semconv.DeploymentEnvironment(o.cfg.Environment),
	}
	if o.cfg.ServiceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersion(o.cfg.ServiceVersion))
	}
	if hostname, _ := os.Hostname(); hostname != "" {
		attrs = append(attrs, semconv.HostName(hostname))
	}
	return resource.NewWithAttributes(semconv.SchemaURL, attrs...), nil
}

func (o *Observability) createSampler() sdktrace.Sampler {
	switch o.cfg.Tracing.Sampler {
	case "always":
		return sdktrace.AlwaysSample()
	case "never":
		return sdktrace.NeverSample()
	case "parent":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(o.cfg.Tracing.Ratio))
	default:
		return sdktrace.TraceIDRatioBased(o.cfg.Tracing.Ratio)
	}
}

func (o *Observability) initMetrics() {
	o.registry = prometheus.NewRegistry()
	if o.cfg.Metrics.Enabled {
		o.registry.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
		o.metrics = NewPrometheusProvider(o.registry)
	} else {
		o.metrics = NoopMetricsProvider()
	}
}

// TracerProvider 返回 TracerProvider
func (o *Observability) TracerProvider() *sdktrace.TracerProvider {
	return o.tp
}

// Registry 返回 Prometheus Registry
func (o *Observability) Registry() *prometheus.Registry {
	return o.registry
}

// Metrics 返回指标提供者
func (o *Observability) Metrics() MetricsProvider {
	return o.metrics
}

// MetricsHandler 返回指标 HTTP Handler
func (o *Observability) MetricsHandler() http.Handler {
	return o.metrics.Handler()
}

// Shutdown 关闭
func (o *Observability) Shutdown(ctx context.Context) error {
	if o.tp != nil {
		return o.tp.Shutdown(ctx)
	}
	return nil
}

type noopExporter struct{}

func (e *noopExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return nil
}
func (e *noopExporter) Shutdown(ctx context.Context) error { return nil }
