package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/mildsunup/higo/http"

// TracingConfig 追踪配置
type TracingConfig struct {
	ServiceName    string
	SkipPaths      []string          // 跳过追踪的路径
	SpanNameFunc   func(*gin.Context) string
	TracerProvider trace.TracerProvider
}

// DefaultTracingConfig 默认配置
func DefaultTracingConfig() TracingConfig {
	return TracingConfig{
		ServiceName: "http-server",
		SkipPaths:   []string{"/health", "/metrics", "/ready"},
		SpanNameFunc: func(c *gin.Context) string {
			return fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		},
	}
}

// Tracing 追踪中间件
func Tracing(cfg TracingConfig) gin.HandlerFunc {
	if cfg.SpanNameFunc == nil {
		cfg.SpanNameFunc = DefaultTracingConfig().SpanNameFunc
	}

	skipPaths := make(map[string]bool)
	for _, p := range cfg.SkipPaths {
		skipPaths[p] = true
	}

	tp := cfg.TracerProvider
	if tp == nil {
		tp = otel.GetTracerProvider()
	}
	tracer := tp.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		// 跳过指定路径
		if skipPaths[c.FullPath()] {
			c.Next()
			return
		}

		// 从请求头提取追踪上下文
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// 创建 Span
		spanName := cfg.SpanNameFunc(c)
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethod(c.Request.Method),
				semconv.HTTPTarget(c.Request.URL.Path),
				semconv.HTTPScheme(scheme(c)),
				semconv.NetHostName(c.Request.Host),
				semconv.UserAgentOriginal(c.Request.UserAgent()),
				semconv.HTTPRequestContentLength(int(c.Request.ContentLength)),
				attribute.String("http.client_ip", c.ClientIP()),
			),
		)
		defer span.End()

		// 注入 trace_id 到响应头（便于调试）
		if span.SpanContext().HasTraceID() {
			c.Header("X-Trace-ID", span.SpanContext().TraceID().String())
		}

		// 更新请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 执行后续处理
		c.Next()

		// 记录响应信息
		status := c.Writer.Status()
		span.SetAttributes(
			semconv.HTTPStatusCode(status),
			semconv.HTTPResponseContentLength(c.Writer.Size()),
		)

		// 设置状态
		if status >= 500 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", status))
		} else if status >= 400 {
			span.SetStatus(codes.Unset, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}

		// 记录错误
		if len(c.Errors) > 0 {
			span.RecordError(c.Errors.Last().Err)
		}
	}
}

func scheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	if s := c.GetHeader("X-Forwarded-Proto"); s != "" {
		return s
	}
	return "http"
}
