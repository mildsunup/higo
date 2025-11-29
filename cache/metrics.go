package cache

import (
	"context"
	"time"

	"higo/observability"
)

// Metrics 缓存指标
type Metrics struct {
	OpsTotal    observability.Counter
	OpsDuration observability.Histogram
	HitRate     observability.Gauge
}

// NewMetrics 创建缓存指标
func NewMetrics(p observability.MetricsProvider) *Metrics {
	return &Metrics{
		OpsTotal:    p.Counter("cache_operations_total", "Total cache operations", "name", "operation", "status"),
		OpsDuration: p.Histogram("cache_operation_duration_seconds", "Cache operation duration", []float64{.0001, .0005, .001, .005, .01, .025, .05, .1}, "name", "operation"),
		HitRate:     p.Gauge("cache_hit_rate", "Cache hit rate", "name"),
	}
}

// Metriced 指标装饰器
type Metriced struct {
	cache   Cache
	name    string
	metrics *Metrics
}

// NewMetriced 创建指标装饰器
func NewMetriced(cache Cache, name string, m *Metrics) *Metriced {
	return &Metriced{cache: cache, name: name, metrics: m}
}

func (m *Metriced) Get(ctx context.Context, key string, dest any) error {
	start := time.Now()
	err := m.cache.Get(ctx, key, dest)

	status := "hit"
	if err == ErrNotFound {
		status = "miss"
	} else if err != nil {
		status = "error"
	}

	m.metrics.OpsTotal.Inc(m.name, "get", status)
	m.metrics.OpsDuration.Since(start, m.name, "get")
	return err
}

func (m *Metriced) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	start := time.Now()
	err := m.cache.Set(ctx, key, value, ttl)

	status := "success"
	if err != nil {
		status = "error"
	}

	m.metrics.OpsTotal.Inc(m.name, "set", status)
	m.metrics.OpsDuration.Since(start, m.name, "set")
	return err
}

func (m *Metriced) Delete(ctx context.Context, keys ...string) error {
	start := time.Now()
	err := m.cache.Delete(ctx, keys...)

	status := "success"
	if err != nil {
		status = "error"
	}

	m.metrics.OpsTotal.Inc(m.name, "delete", status)
	m.metrics.OpsDuration.Since(start, m.name, "delete")
	return err
}

func (m *Metriced) Exists(ctx context.Context, key string) (bool, error) {
	return m.cache.Exists(ctx, key)
}

func (m *Metriced) Close() error  { return m.cache.Close() }
func (m *Metriced) Unwrap() Cache { return m.cache }

// UpdateHitRate 更新命中率指标
func (m *Metriced) UpdateHitRate() {
	if sp, ok := m.cache.(StatsProvider); ok {
		m.metrics.HitRate.Set(sp.Stats().HitRate, m.name)
	}
}

var _ Cache = (*Metriced)(nil)
