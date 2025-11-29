package observability

import (
	"context"
	"reflect"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Traced 包装函数，自动创建 Span
func Traced[T any](ctx context.Context, name string, fn func(context.Context) (T, error)) (T, error) {
	ctx, span := StartSpan(ctx, name)
	defer span.End()

	result, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

// TracedVoid 包装无返回值函数
func TracedVoid(ctx context.Context, name string, fn func(context.Context) error) error {
	ctx, span := StartSpan(ctx, name)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

// AutoSpan 自动从函数名生成 Span
func AutoSpan(ctx context.Context, skip int) (context.Context, trace.Span) {
	pc, _, _, ok := runtime.Caller(skip + 1)
	name := "unknown"
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			name = shortFuncName(fn.Name())
		}
	}
	return StartSpan(ctx, name)
}

// SpanFunc 装饰器 - 自动为方法添加 Span
type SpanFunc[Req, Resp any] func(ctx context.Context, req Req) (Resp, error)

// WithSpan 包装处理函数，自动创建 Span
func WithSpan[Req, Resp any](name string, fn SpanFunc[Req, Resp]) SpanFunc[Req, Resp] {
	return func(ctx context.Context, req Req) (Resp, error) {
		ctx, span := StartSpan(ctx, name)
		defer span.End()

		// 记录请求类型
		span.SetAttributes(attribute.String("request.type", reflect.TypeOf(req).String()))

		resp, err := fn(ctx, req)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return resp, err
	}
}

// TracedHandler 创建带追踪的处理器包装
type TracedHandler[H any] struct {
	handler H
	tracer  trace.Tracer
}

// NewTracedHandler 创建追踪处理器
func NewTracedHandler[H any](handler H, serviceName string) *TracedHandler[H] {
	return &TracedHandler[H]{
		handler: handler,
		tracer:  otel.Tracer(serviceName),
	}
}

// Handler 返回原始处理器
func (t *TracedHandler[H]) Handler() H {
	return t.handler
}

// Tracer 返回 tracer
func (t *TracedHandler[H]) Tracer() trace.Tracer {
	return t.tracer
}

func shortFuncName(fullName string) string {
	// github.com/xxx/higo/internal/application/user.(*Handler).Register
	// -> user.Handler.Register
	if idx := strings.LastIndex(fullName, "/"); idx >= 0 {
		fullName = fullName[idx+1:]
	}
	// 移除指针标记
	fullName = strings.ReplaceAll(fullName, "(*", "")
	fullName = strings.ReplaceAll(fullName, ")", "")
	return fullName
}
