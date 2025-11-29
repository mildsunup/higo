package mq

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Traced 追踪装饰器
type Traced struct {
	client     Client
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

// NewTraced 创建追踪装饰器
func NewTraced(c Client, tracer trace.Tracer) *Traced {
	return &Traced{
		client:     c,
		tracer:     tracer,
		propagator: propagation.TraceContext{},
	}
}

func (t *Traced) Connect(ctx context.Context) error {
	ctx, span := t.startSpan(ctx, "Connect", "")
	defer span.End()
	return t.recordError(span, t.client.Connect(ctx))
}

func (t *Traced) Ping(ctx context.Context) error {
	ctx, span := t.startSpan(ctx, "Ping", "")
	defer span.End()
	return t.recordError(span, t.client.Ping(ctx))
}

func (t *Traced) Publish(ctx context.Context, topic string, value []byte, opts ...PublishOption) (*PublishResult, error) {
	ctx, span := t.startSpan(ctx, "Publish", topic)
	defer span.End()

	span.SetAttributes(
		attribute.Int("message.size", len(value)),
	)

	// 注入 trace context 到消息头
	opts = append(opts, WithHeaders(t.injectTraceContext(ctx)))

	result, err := t.client.Publish(ctx, topic, value, opts...)
	if err != nil {
		t.recordError(span, err)
		return nil, err
	}

	span.SetAttributes(
		attribute.String("message.id", result.MessageID),
		attribute.Int("message.partition", int(result.Partition)),
		attribute.Int64("message.offset", result.Offset),
	)
	span.SetStatus(codes.Ok, "")
	return result, nil
}

func (t *Traced) PublishAsync(ctx context.Context, topic string, value []byte, callback func(*PublishResult, error), opts ...PublishOption) {
	ctx, span := t.startSpan(ctx, "PublishAsync", topic)

	opts = append(opts, WithHeaders(t.injectTraceContext(ctx)))

	t.client.PublishAsync(ctx, topic, value, func(result *PublishResult, err error) {
		defer span.End()
		if err != nil {
			t.recordError(span, err)
		} else {
			span.SetStatus(codes.Ok, "")
		}
		if callback != nil {
			callback(result, err)
		}
	}, opts...)
}

func (t *Traced) Subscribe(ctx context.Context, topic string, handler Handler, opts ...SubscribeOption) error {
	// 包装 handler 以提取 trace context
	tracedHandler := func(ctx context.Context, msg *Message) error {
		// 从消息头提取 trace context
		ctx = t.extractTraceContext(ctx, msg.Headers)

		ctx, span := t.tracer.Start(ctx, "mq.Consume",
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(
				attribute.String("mq.name", t.client.Name()),
				attribute.String("mq.type", string(t.client.Type())),
				attribute.String("mq.topic", topic),
				attribute.String("mq.message_id", msg.ID),
			),
		)
		defer span.End()

		err := handler(ctx, msg)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
		return err
	}

	return t.client.Subscribe(ctx, topic, tracedHandler, opts...)
}

func (t *Traced) Unsubscribe(topic string) error {
	return t.client.Unsubscribe(topic)
}

func (t *Traced) Close() error {
	return t.client.Close()
}

func (t *Traced) Name() string  { return t.client.Name() }
func (t *Traced) Type() Type    { return t.client.Type() }
func (t *Traced) State() State  { return t.client.State() }
func (t *Traced) Stats() Stats  { return t.client.Stats() }

// Unwrap 获取底层客户端
func (t *Traced) Unwrap() Client { return t.client }

func (t *Traced) startSpan(ctx context.Context, op, topic string) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String("mq.name", t.client.Name()),
		attribute.String("mq.type", string(t.client.Type())),
	}
	if topic != "" {
		attrs = append(attrs, attribute.String("mq.topic", topic))
	}

	return t.tracer.Start(ctx, "mq."+op,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attrs...),
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

func (t *Traced) injectTraceContext(ctx context.Context) map[string]string {
	carrier := make(propagation.MapCarrier)
	t.propagator.Inject(ctx, carrier)
	result := make(map[string]string)
	for k, v := range carrier {
		result[k] = v
	}
	return result
}

func (t *Traced) extractTraceContext(ctx context.Context, headers map[string]string) context.Context {
	if headers == nil {
		return ctx
	}
	carrier := make(propagation.MapCarrier)
	for k, v := range headers {
		carrier[k] = v
	}
	return t.propagator.Extract(ctx, carrier)
}

var _ Client = (*Traced)(nil)
