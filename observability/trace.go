package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "higo"

// StartSpan 开始一个新的 Span
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, name, opts...)
}

// SpanFromContext 从 context 获取当前 Span
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// TraceID 从 context 获取 TraceID
func TraceID(ctx context.Context) string {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if sc.IsValid() {
		return sc.TraceID().String()
	}
	return ""
}

// SpanID 从 context 获取 SpanID
func SpanID(ctx context.Context) string {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if sc.IsValid() {
		return sc.SpanID().String()
	}
	return ""
}

// RecordError 记录错误到当前 Span
func RecordError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// SetAttributes 设置属性到当前 Span
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	trace.SpanFromContext(ctx).SetAttributes(attrs...)
}

// AddEvent 添加事件到当前 Span
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	trace.SpanFromContext(ctx).AddEvent(name, trace.WithAttributes(attrs...))
}

// Inject 注入追踪上下文到 carrier
func Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	otel.GetTextMapPropagator().Inject(ctx, carrier)
}

// Extract 从 carrier 提取追踪上下文
func Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}

// SetBaggage 设置 baggage
func SetBaggage(ctx context.Context, key, value string) context.Context {
	member, _ := baggage.NewMember(key, value)
	bag, _ := baggage.New(member)
	return baggage.ContextWithBaggage(ctx, bag)
}

// GetBaggage 获取 baggage
func GetBaggage(ctx context.Context, key string) string {
	return baggage.FromContext(ctx).Member(key).Value()
}
