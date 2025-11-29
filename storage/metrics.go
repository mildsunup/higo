package storage

import (
	"context"
	"time"

	"github.com/mildsunup/higo/observability"
)

// PoolMetrics 连接池指标
type PoolMetrics struct {
	maxOpen      observability.Gauge
	open         observability.Gauge
	inUse        observability.Gauge
	idle         observability.Gauge
	waitCount    observability.Counter
	waitDuration observability.Histogram
}

// NewPoolMetrics 创建连接池指标
func NewPoolMetrics(mp observability.MetricsProvider, subsystem string) *PoolMetrics {
	prefix := "storage_pool_"
	if subsystem != "" {
		prefix = subsystem + "_pool_"
	}
	return &PoolMetrics{
		maxOpen:      mp.Gauge(prefix+"max_open", "Maximum number of open connections", "name", "type"),
		open:         mp.Gauge(prefix+"open", "Current open connections", "name", "type"),
		inUse:        mp.Gauge(prefix+"in_use", "Connections currently in use", "name", "type"),
		idle:         mp.Gauge(prefix+"idle", "Idle connections", "name", "type"),
		waitCount:    mp.Counter(prefix+"wait_total", "Total wait count for connections", "name", "type"),
		waitDuration: mp.Histogram(prefix+"wait_seconds", "Wait duration for connections", observability.DurationBuckets, "name", "type"),
	}
}

// Record 记录连接池统计
func (m *PoolMetrics) Record(name string, typ Type, stats Stats) {
	labels := []string{name, string(typ)}
	m.maxOpen.Set(float64(stats.MaxOpenConnections), labels...)
	m.open.Set(float64(stats.OpenConnections), labels...)
	m.inUse.Set(float64(stats.InUse), labels...)
	m.idle.Set(float64(stats.Idle), labels...)
}

// Collector 连接池指标收集器
type Collector struct {
	metrics  *PoolMetrics
	storages []Storage
	interval time.Duration
	stopCh   chan struct{}
}

// NewCollector 创建收集器
func NewCollector(metrics *PoolMetrics, interval time.Duration) *Collector {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	return &Collector{
		metrics:  metrics,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Register 注册存储
func (c *Collector) Register(s Storage) {
	c.storages = append(c.storages, s)
}

// Start 启动收集
func (c *Collector) Start(ctx context.Context) {
	go c.run(ctx)
}

// Stop 停止收集
func (c *Collector) Stop() {
	close(c.stopCh)
}

func (c *Collector) run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	c.collect() // 立即收集一次

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.collect()
		}
	}
}

func (c *Collector) collect() {
	for _, s := range c.storages {
		if sp, ok := s.(StatsProvider); ok {
			c.metrics.Record(s.Name(), s.Type(), sp.Stats())
		}
	}
}

// Metrics 存储操作指标 (用于装饰器)
type Metrics struct {
	operations observability.Counter
	duration   observability.Histogram
	errors     observability.Counter
}

// NewMetrics 创建存储操作指标
func NewMetrics(mp observability.MetricsProvider) *Metrics {
	return &Metrics{
		operations: mp.Counter("storage_operations_total", "Total storage operations", "name", "type", "operation"),
		duration:   mp.Histogram("storage_operation_seconds", "Storage operation duration", observability.DurationBuckets, "name", "type", "operation"),
		errors:     mp.Counter("storage_errors_total", "Total storage errors", "name", "type", "operation"),
	}
}

// Metriced 带指标的存储装饰器
type Metriced struct {
	storage Storage
	metrics *Metrics
}

// NewMetriced 创建带指标的存储
func NewMetriced(s Storage, m *Metrics) *Metriced {
	return &Metriced{storage: s, metrics: m}
}

func (m *Metriced) Connect(ctx context.Context) error {
	start := time.Now()
	err := m.storage.Connect(ctx)
	m.record("connect", start, err)
	return err
}

func (m *Metriced) Ping(ctx context.Context) error {
	start := time.Now()
	err := m.storage.Ping(ctx)
	m.record("ping", start, err)
	return err
}

func (m *Metriced) Close(ctx context.Context) error {
	start := time.Now()
	err := m.storage.Close(ctx)
	m.record("close", start, err)
	return err
}

func (m *Metriced) Name() string  { return m.storage.Name() }
func (m *Metriced) Type() Type    { return m.storage.Type() }
func (m *Metriced) State() State  { return m.storage.State() }
func (m *Metriced) Unwrap() Storage { return m.storage }

func (m *Metriced) record(op string, start time.Time, err error) {
	labels := []string{m.storage.Name(), string(m.storage.Type()), op}
	m.metrics.operations.Inc(labels...)
	m.metrics.duration.Since(start, labels...)
	if err != nil {
		m.metrics.errors.Inc(labels...)
	}
}

var _ Storage = (*Metriced)(nil)
