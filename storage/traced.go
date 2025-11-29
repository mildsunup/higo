package storage

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Traced 追踪装饰器
type Traced struct {
	storage Storage
	tracer  trace.Tracer
}

// NewTraced 创建追踪装饰器
func NewTraced(s Storage, tracer trace.Tracer) *Traced {
	return &Traced{storage: s, tracer: tracer}
}

func (t *Traced) Connect(ctx context.Context) error {
	ctx, span := t.startSpan(ctx, "Connect")
	defer span.End()
	return t.recordError(span, t.storage.Connect(ctx))
}

func (t *Traced) Ping(ctx context.Context) error {
	ctx, span := t.startSpan(ctx, "Ping")
	defer span.End()
	return t.recordError(span, t.storage.Ping(ctx))
}

func (t *Traced) Close(ctx context.Context) error {
	ctx, span := t.startSpan(ctx, "Close")
	defer span.End()
	return t.recordError(span, t.storage.Close(ctx))
}

func (t *Traced) Name() string  { return t.storage.Name() }
func (t *Traced) Type() Type    { return t.storage.Type() }
func (t *Traced) State() State  { return t.storage.State() }

// Unwrap 获取底层存储
func (t *Traced) Unwrap() Storage { return t.storage }

func (t *Traced) startSpan(ctx context.Context, op string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, "storage."+op,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("storage.name", t.storage.Name()),
			attribute.String("storage.type", string(t.storage.Type())),
		),
	)
}

func (t *Traced) recordError(span trace.Span, err error) error {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	return err
}

var _ Storage = (*Traced)(nil)
