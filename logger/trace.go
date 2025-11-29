package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	"higo/middleware"
)

func init() {
	RegisterExtractor(TraceExtractor)
	RegisterExtractor(RequestIDExtractor)
}

// TraceExtractor 从 context 提取 trace 信息
func TraceExtractor(ctx context.Context) []Field {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return nil
	}

	sc := span.SpanContext()
	return []Field{
		String("trace_id", sc.TraceID().String()),
		String("span_id", sc.SpanID().String()),
	}
}

// RequestIDExtractor 从 context 提取 request_id
func RequestIDExtractor(ctx context.Context) []Field {
	if reqID := middleware.GetRequestID(ctx); reqID != "" {
		return []Field{String("request_id", reqID)}
	}
	return nil
}
