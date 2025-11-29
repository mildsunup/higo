package grpc

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const tracerName = "github.com/mildsunup/higo/grpc"

// TracingConfig 追踪配置
type TracingConfig struct {
	ServiceName    string
	SkipMethods    []string // 跳过追踪的方法
	TracerProvider trace.TracerProvider
}

// DefaultTracingConfig 默认配置
func DefaultTracingConfig() TracingConfig {
	return TracingConfig{
		ServiceName: "grpc-server",
		SkipMethods: []string{
			"/grpc.health.v1.Health/Check",
			"/grpc.health.v1.Health/Watch",
			"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
		},
	}
}

// UnaryServerTracing 一元 RPC 追踪拦截器
func UnaryServerTracing(cfg TracingConfig) grpc.UnaryServerInterceptor {
	skipMethods := make(map[string]bool)
	for _, m := range cfg.SkipMethods {
		skipMethods[m] = true
	}

	tp := cfg.TracerProvider
	if tp == nil {
		tp = otel.GetTracerProvider()
	}
	tracer := tp.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 跳过指定方法
		if skipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// 从 metadata 提取追踪上下文
		ctx = extractContext(ctx, propagator)

		// 创建 Span
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.RPCSystemGRPC,
				semconv.RPCService(serviceFromMethod(info.FullMethod)),
				semconv.RPCMethod(methodFromMethod(info.FullMethod)),
			),
		)
		defer span.End()

		// 执行处理
		resp, err := handler(ctx, req)

		// 记录状态
		if err != nil {
			s, _ := status.FromError(err)
			span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int(int(s.Code())))
			if s.Code() != grpccodes.OK {
				span.SetStatus(codes.Error, s.Message())
				span.RecordError(err)
			}
		} else {
			span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int(int(grpccodes.OK)))
			span.SetStatus(codes.Ok, "")
		}

		return resp, err
	}
}

// StreamServerTracing 流式 RPC 追踪拦截器
func StreamServerTracing(cfg TracingConfig) grpc.StreamServerInterceptor {
	skipMethods := make(map[string]bool)
	for _, m := range cfg.SkipMethods {
		skipMethods[m] = true
	}

	tp := cfg.TracerProvider
	if tp == nil {
		tp = otel.GetTracerProvider()
	}
	tracer := tp.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 跳过指定方法
		if skipMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		ctx := ss.Context()
		ctx = extractContext(ctx, propagator)

		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.RPCSystemGRPC,
				semconv.RPCService(serviceFromMethod(info.FullMethod)),
				semconv.RPCMethod(methodFromMethod(info.FullMethod)),
				attribute.Bool("rpc.grpc.is_client_stream", info.IsClientStream),
				attribute.Bool("rpc.grpc.is_server_stream", info.IsServerStream),
			),
		)
		defer span.End()

		// 包装 stream 以传递新的 context
		wrapped := &wrappedServerStream{ServerStream: ss, ctx: ctx}

		err := handler(srv, wrapped)

		if err != nil {
			s, _ := status.FromError(err)
			span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int(int(s.Code())))
			span.SetStatus(codes.Error, s.Message())
			span.RecordError(err)
		} else {
			span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int(int(grpccodes.OK)))
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

// --- Client Interceptors ---

// UnaryClientTracing 客户端一元 RPC 追踪拦截器
func UnaryClientTracing(cfg TracingConfig) grpc.UnaryClientInterceptor {
	tp := cfg.TracerProvider
	if tp == nil {
		tp = otel.GetTracerProvider()
	}
	tracer := tp.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, span := tracer.Start(ctx, method,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(
				semconv.RPCSystemGRPC,
				semconv.RPCService(serviceFromMethod(method)),
				semconv.RPCMethod(methodFromMethod(method)),
			),
		)
		defer span.End()

		// 注入追踪上下文到 metadata
		ctx = injectContext(ctx, propagator)

		err := invoker(ctx, method, req, reply, cc, opts...)

		if err != nil {
			s, _ := status.FromError(err)
			span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int(int(s.Code())))
			span.SetStatus(codes.Error, s.Message())
			span.RecordError(err)
		} else {
			span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int(int(grpccodes.OK)))
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

// --- Helpers ---

type metadataCarrier metadata.MD

func (m metadataCarrier) Get(key string) string {
	vals := metadata.MD(m).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (m metadataCarrier) Set(key, value string) {
	metadata.MD(m).Set(key, value)
}

func (m metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func extractContext(ctx context.Context, propagator propagation.TextMapPropagator) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	return propagator.Extract(ctx, metadataCarrier(md))
}

func injectContext(ctx context.Context, propagator propagation.TextMapPropagator) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	propagator.Inject(ctx, metadataCarrier(md))
	return metadata.NewOutgoingContext(ctx, md)
}

func serviceFromMethod(fullMethod string) string {
	if len(fullMethod) > 0 && fullMethod[0] == '/' {
		fullMethod = fullMethod[1:]
	}
	for i := len(fullMethod) - 1; i >= 0; i-- {
		if fullMethod[i] == '/' {
			return fullMethod[:i]
		}
	}
	return fullMethod
}

func methodFromMethod(fullMethod string) string {
	for i := len(fullMethod) - 1; i >= 0; i-- {
		if fullMethod[i] == '/' {
			return fullMethod[i+1:]
		}
	}
	return fullMethod
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
