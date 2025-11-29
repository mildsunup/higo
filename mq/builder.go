package mq

import (
	"go.opentelemetry.io/otel/trace"
)

// Builder MQ 客户端构建器
type Builder struct {
	client  Client
	tracer  trace.Tracer
	metrics *Metrics
}

// NewBuilder 创建构建器
func NewBuilder(c Client) *Builder {
	return &Builder{client: c}
}

// WithTracing 启用追踪
func (b *Builder) WithTracing(tracer trace.Tracer) *Builder {
	b.tracer = tracer
	return b
}

// WithMetrics 启用指标
func (b *Builder) WithMetrics(m *Metrics) *Builder {
	b.metrics = m
	return b
}

// Build 构建最终客户端（装饰器顺序：Traced -> Metriced）
func (b *Builder) Build() Client {
	c := b.client

	if b.tracer != nil {
		c = NewTraced(c, b.tracer)
	}
	if b.metrics != nil {
		c = NewMetriced(c, b.metrics)
	}

	return c
}

// Unwrap 解包装饰器获取原始客户端
func Unwrap(c Client) Client {
	for {
		switch v := c.(type) {
		case *Traced:
			c = v.Unwrap()
		case *Metriced:
			c = v.Unwrap()
		default:
			return c
		}
	}
}
