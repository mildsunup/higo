package storage

import (
	"go.opentelemetry.io/otel/trace"
)

// Builder 存储构建器
type Builder struct {
	storage   Storage
	reconnect *ReconnectConfig
	tracer    trace.Tracer
	metrics   *Metrics
}

// NewBuilder 创建构建器
func NewBuilder(s Storage) *Builder {
	return &Builder{storage: s}
}

// WithReconnect 启用自动重连
func (b *Builder) WithReconnect(cfg ReconnectConfig) *Builder {
	b.reconnect = &cfg
	return b
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

// Build 构建最终存储（装饰器顺序：Reconnectable -> Traced -> Metriced）
func (b *Builder) Build() Storage {
	s := b.storage

	if b.reconnect != nil {
		s = NewReconnectable(s, *b.reconnect)
	}
	if b.tracer != nil {
		s = NewTraced(s, b.tracer)
	}
	if b.metrics != nil {
		s = NewMetriced(s, b.metrics)
	}

	return s
}

// Unwrap 解包装饰器获取原始存储
func Unwrap(s Storage) Storage {
	for {
		switch v := s.(type) {
		case *Reconnectable:
			s = v.Unwrap()
		case *Traced:
			s = v.Unwrap()
		case *Metriced:
			s = v.Unwrap()
		default:
			return s
		}
	}
}
