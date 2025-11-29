package middleware

import (
	"context"
)

// 上下文键类型
type ctxKey string

const (
	RequestIDKey ctxKey = "request_id"
	TraceIDKey   ctxKey = "trace_id"
	SpanIDKey    ctxKey = "span_id"
	UserIDKey    ctxKey = "user_id"
	TenantIDKey  ctxKey = "tenant_id"
	ClientIPKey  ctxKey = "client_ip"
	UserAgentKey ctxKey = "user_agent"
)

// WithValue 设置上下文值
func WithValue(ctx context.Context, key ctxKey, value any) context.Context {
	return context.WithValue(ctx, key, value)
}

// GetValue 获取上下文值
func GetValue[T any](ctx context.Context, key ctxKey) (T, bool) {
	val := ctx.Value(key)
	if val == nil {
		var zero T
		return zero, false
	}
	v, ok := val.(T)
	return v, ok
}

// GetRequestID 获取请求 ID
func GetRequestID(ctx context.Context) string {
	v, _ := GetValue[string](ctx, RequestIDKey)
	return v
}

// GetTraceID 获取追踪 ID
func GetTraceID(ctx context.Context) string {
	v, _ := GetValue[string](ctx, TraceIDKey)
	return v
}

// GetSpanID 获取 Span ID
func GetSpanID(ctx context.Context) string {
	v, _ := GetValue[string](ctx, SpanIDKey)
	return v
}

// GetUserID 获取用户 ID
func GetUserID(ctx context.Context) (uint64, bool) {
	return GetValue[uint64](ctx, UserIDKey)
}

// GetTenantID 获取租户 ID
func GetTenantID(ctx context.Context) (string, bool) {
	return GetValue[string](ctx, TenantIDKey)
}

// GetClientIP 获取客户端 IP
func GetClientIP(ctx context.Context) string {
	v, _ := GetValue[string](ctx, ClientIPKey)
	return v
}
