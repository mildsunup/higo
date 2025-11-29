package observability

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 预定义 Buckets
var (
	DurationBuckets = []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
	SizeBuckets     = []float64{100, 1000, 10000, 100000, 1000000, 10000000}
)

// Counter 计数器接口
type Counter interface {
	Inc(labels ...string)
	Add(v float64, labels ...string)
	With(labels ...string) BoundCounter // 预绑定标签
}

// BoundCounter 预绑定标签的计数器
type BoundCounter interface {
	Inc()
	Add(v float64)
}

// Gauge 仪表盘接口
type Gauge interface {
	Set(v float64, labels ...string)
	Inc(labels ...string)
	Dec(labels ...string)
	Add(v float64, labels ...string)
	With(labels ...string) BoundGauge
}

// BoundGauge 预绑定标签的仪表盘
type BoundGauge interface {
	Set(v float64)
	Inc()
	Dec()
	Add(v float64)
}

// Histogram 直方图接口
type Histogram interface {
	Observe(v float64, labels ...string)
	Since(start time.Time, labels ...string)
	Timer(labels ...string) func()
	With(labels ...string) BoundHistogram
}

// BoundHistogram 预绑定标签的直方图
type BoundHistogram interface {
	Observe(v float64)
	Since(start time.Time)
	Timer() func()
}

// MetricsProvider 指标提供者接口
type MetricsProvider interface {
	Counter(name, help string, labels ...string) Counter
	Gauge(name, help string, labels ...string) Gauge
	Histogram(name, help string, buckets []float64, labels ...string) Histogram
	Handler() http.Handler
}

// --- Prometheus 实现 ---

type prometheusProvider struct {
	registry *prometheus.Registry
}

// NewPrometheusProvider 创建 Prometheus 指标提供者
func NewPrometheusProvider(registry *prometheus.Registry) MetricsProvider {
	return &prometheusProvider{registry: registry}
}

func (p *prometheusProvider) Counter(name, help string, labels ...string) Counter {
	vec := prometheus.NewCounterVec(prometheus.CounterOpts{Name: name, Help: help}, labels)
	p.registry.MustRegister(vec)
	return &promCounter{vec: vec}
}

func (p *prometheusProvider) Gauge(name, help string, labels ...string) Gauge {
	vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: name, Help: help}, labels)
	p.registry.MustRegister(vec)
	return &promGauge{vec: vec}
}

func (p *prometheusProvider) Histogram(name, help string, buckets []float64, labels ...string) Histogram {
	vec := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: name, Help: help, Buckets: buckets}, labels)
	p.registry.MustRegister(vec)
	return &promHistogram{vec: vec}
}

func (p *prometheusProvider) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{EnableOpenMetrics: true})
}

// --- Prometheus Counter ---

type promCounter struct{ vec *prometheus.CounterVec }

func (c *promCounter) Inc(labels ...string)            { c.vec.WithLabelValues(labels...).Inc() }
func (c *promCounter) Add(v float64, labels ...string) { c.vec.WithLabelValues(labels...).Add(v) }
func (c *promCounter) With(labels ...string) BoundCounter {
	return &boundPromCounter{c: c.vec.WithLabelValues(labels...)}
}

type boundPromCounter struct{ c prometheus.Counter }

func (b *boundPromCounter) Inc()           { b.c.Inc() }
func (b *boundPromCounter) Add(v float64)  { b.c.Add(v) }

// --- Prometheus Gauge ---

type promGauge struct{ vec *prometheus.GaugeVec }

func (g *promGauge) Set(v float64, labels ...string) { g.vec.WithLabelValues(labels...).Set(v) }
func (g *promGauge) Inc(labels ...string)            { g.vec.WithLabelValues(labels...).Inc() }
func (g *promGauge) Dec(labels ...string)            { g.vec.WithLabelValues(labels...).Dec() }
func (g *promGauge) Add(v float64, labels ...string) { g.vec.WithLabelValues(labels...).Add(v) }
func (g *promGauge) With(labels ...string) BoundGauge {
	return &boundPromGauge{g: g.vec.WithLabelValues(labels...)}
}

type boundPromGauge struct{ g prometheus.Gauge }

func (b *boundPromGauge) Set(v float64) { b.g.Set(v) }
func (b *boundPromGauge) Inc()          { b.g.Inc() }
func (b *boundPromGauge) Dec()          { b.g.Dec() }
func (b *boundPromGauge) Add(v float64) { b.g.Add(v) }

// --- Prometheus Histogram ---

type promHistogram struct{ vec *prometheus.HistogramVec }

func (h *promHistogram) Observe(v float64, labels ...string) { h.vec.WithLabelValues(labels...).Observe(v) }
func (h *promHistogram) Since(start time.Time, labels ...string) {
	h.Observe(time.Since(start).Seconds(), labels...)
}
func (h *promHistogram) Timer(labels ...string) func() {
	start := time.Now()
	return func() { h.Since(start, labels...) }
}
func (h *promHistogram) With(labels ...string) BoundHistogram {
	return &boundPromHistogram{h: h.vec.WithLabelValues(labels...)}
}

type boundPromHistogram struct{ h prometheus.Observer }

func (b *boundPromHistogram) Observe(v float64)       { b.h.Observe(v) }
func (b *boundPromHistogram) Since(start time.Time)   { b.h.Observe(time.Since(start).Seconds()) }
func (b *boundPromHistogram) Timer() func() {
	start := time.Now()
	return func() { b.Since(start) }
}

// --- Noop 实现 ---

type noopProvider struct{}

// NoopMetricsProvider 返回空实现
func NoopMetricsProvider() MetricsProvider { return &noopProvider{} }

func (p *noopProvider) Counter(name, help string, labels ...string) Counter   { return &noopCounter{} }
func (p *noopProvider) Gauge(name, help string, labels ...string) Gauge       { return &noopGauge{} }
func (p *noopProvider) Histogram(name, help string, buckets []float64, labels ...string) Histogram {
	return &noopHistogram{}
}
func (p *noopProvider) Handler() http.Handler { return http.NotFoundHandler() }

type noopCounter struct{}

func (c *noopCounter) Inc(labels ...string)            {}
func (c *noopCounter) Add(v float64, labels ...string) {}
func (c *noopCounter) With(labels ...string) BoundCounter { return &noopBoundCounter{} }

type noopBoundCounter struct{}

func (c *noopBoundCounter) Inc()          {}
func (c *noopBoundCounter) Add(v float64) {}

type noopGauge struct{}

func (g *noopGauge) Set(v float64, labels ...string) {}
func (g *noopGauge) Inc(labels ...string)            {}
func (g *noopGauge) Dec(labels ...string)            {}
func (g *noopGauge) Add(v float64, labels ...string) {}
func (g *noopGauge) With(labels ...string) BoundGauge { return &noopBoundGauge{} }

type noopBoundGauge struct{}

func (g *noopBoundGauge) Set(v float64) {}
func (g *noopBoundGauge) Inc()          {}
func (g *noopBoundGauge) Dec()          {}
func (g *noopBoundGauge) Add(v float64) {}

type noopHistogram struct{}

func (h *noopHistogram) Observe(v float64, labels ...string)     {}
func (h *noopHistogram) Since(start time.Time, labels ...string) {}
func (h *noopHistogram) Timer(labels ...string) func()           { return func() {} }
func (h *noopHistogram) With(labels ...string) BoundHistogram    { return &noopBoundHistogram{} }

type noopBoundHistogram struct{}

func (h *noopBoundHistogram) Observe(v float64)     {}
func (h *noopBoundHistogram) Since(start time.Time) {}
func (h *noopBoundHistogram) Timer() func()         { return func() {} }
