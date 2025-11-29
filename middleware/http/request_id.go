package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	mw "higo/middleware"
)

const (
	HeaderRequestID = "X-Request-ID"
	HeaderTraceID   = "X-Trace-ID"
)

// RequestID 请求 ID 中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从 header 获取
		requestID := c.GetHeader(HeaderRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置到 context
		ctx := mw.WithValue(c.Request.Context(), mw.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// 设置响应 header
		c.Header(HeaderRequestID, requestID)

		// 从 trace 获取 trace_id 和 span_id
		span := trace.SpanFromContext(ctx)
		if span.SpanContext().IsValid() {
			traceID := span.SpanContext().TraceID().String()
			spanID := span.SpanContext().SpanID().String()

			ctx = mw.WithValue(ctx, mw.TraceIDKey, traceID)
			ctx = mw.WithValue(ctx, mw.SpanIDKey, spanID)
			c.Request = c.Request.WithContext(ctx)

			c.Header(HeaderTraceID, traceID)
		}

		c.Next()
	}
}
