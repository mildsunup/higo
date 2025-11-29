package mq

import (
	"context"
	"time"

	"github.com/mildsunup/higo/observability"
)

// Metrics MQ 指标
type Metrics struct {
	MessagesTotal   observability.Counter
	MessageDuration observability.Histogram
	MessageSize     observability.Histogram
	ConsumerLag     observability.Gauge
}

// NewMetrics 创建 MQ 指标
func NewMetrics(p observability.MetricsProvider) *Metrics {
	return &Metrics{
		MessagesTotal:   p.Counter("mq_messages_total", "Total MQ messages", "client", "type", "topic", "operation", "status"),
		MessageDuration: p.Histogram("mq_message_duration_seconds", "MQ message processing duration", observability.DurationBuckets, "client", "type", "topic", "operation"),
		MessageSize:     p.Histogram("mq_message_size_bytes", "MQ message size", observability.SizeBuckets, "client", "type", "topic"),
		ConsumerLag:     p.Gauge("mq_consumer_lag", "MQ consumer lag", "client", "type", "topic", "group"),
	}
}

// Metriced 指标装饰器
type Metriced struct {
	client  Client
	metrics *Metrics
}

// NewMetriced 创建指标装饰器
func NewMetriced(c Client, m *Metrics) *Metriced {
	return &Metriced{client: c, metrics: m}
}

func (m *Metriced) Connect(ctx context.Context) error { return m.client.Connect(ctx) }
func (m *Metriced) Ping(ctx context.Context) error    { return m.client.Ping(ctx) }
func (m *Metriced) Name() string                      { return m.client.Name() }
func (m *Metriced) Type() Type                        { return m.client.Type() }
func (m *Metriced) State() State                      { return m.client.State() }
func (m *Metriced) Stats() Stats                      { return m.client.Stats() }
func (m *Metriced) Unsubscribe(topic string) error    { return m.client.Unsubscribe(topic) }
func (m *Metriced) Close() error                      { return m.client.Close() }
func (m *Metriced) Unwrap() Client                    { return m.client }

func (m *Metriced) Publish(ctx context.Context, topic string, value []byte, opts ...PublishOption) (*PublishResult, error) {
	name, typ := m.client.Name(), string(m.client.Type())
	start := time.Now()

	result, err := m.client.Publish(ctx, topic, value, opts...)

	status := "success"
	if err != nil {
		status = "error"
	}

	m.metrics.MessagesTotal.Inc(name, typ, topic, "publish", status)
	m.metrics.MessageDuration.Since(start, name, typ, topic, "publish")
	m.metrics.MessageSize.Observe(float64(len(value)), name, typ, topic)
	return result, err
}

func (m *Metriced) PublishAsync(ctx context.Context, topic string, value []byte, callback func(*PublishResult, error), opts ...PublishOption) {
	name, typ := m.client.Name(), string(m.client.Type())
	start := time.Now()

	m.metrics.MessageSize.Observe(float64(len(value)), name, typ, topic)

	m.client.PublishAsync(ctx, topic, value, func(result *PublishResult, err error) {
		status := "success"
		if err != nil {
			status = "error"
		}
		m.metrics.MessagesTotal.Inc(name, typ, topic, "publish_async", status)
		m.metrics.MessageDuration.Since(start, name, typ, topic, "publish_async")
		if callback != nil {
			callback(result, err)
		}
	}, opts...)
}

func (m *Metriced) Subscribe(ctx context.Context, topic string, handler Handler, opts ...SubscribeOption) error {
	name, typ := m.client.Name(), string(m.client.Type())

	wrapped := func(ctx context.Context, msg *Message) error {
		start := time.Now()
		err := handler(ctx, msg)

		status := "success"
		if err != nil {
			status = "error"
		}

		m.metrics.MessagesTotal.Inc(name, typ, topic, "consume", status)
		m.metrics.MessageDuration.Since(start, name, typ, topic, "consume")
		m.metrics.MessageSize.Observe(float64(len(msg.Value)), name, typ, topic)
		return err
	}

	return m.client.Subscribe(ctx, topic, wrapped, opts...)
}

// SetConsumerLag 设置消费延迟
func (m *Metriced) SetConsumerLag(topic, group string, lag float64) {
	m.metrics.ConsumerLag.Set(lag, m.client.Name(), string(m.client.Type()), topic, group)
}

var _ Client = (*Metriced)(nil)
