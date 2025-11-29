package grpc

import (
	"context"
	"runtime/debug"
	"time"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/mildsunup/higo/logger"
	mw "github.com/mildsunup/higo/middleware"
)

// UnaryLogging 一元调用日志拦截器
func UnaryLogging(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		// 提取追踪信息到 context
		span := trace.SpanFromContext(ctx)
		if span.SpanContext().IsValid() {
			ctx = mw.WithValue(ctx, mw.TraceIDKey, span.SpanContext().TraceID().String())
			ctx = mw.WithValue(ctx, mw.SpanIDKey, span.SpanContext().SpanID().String())
		}

		// 提取 request_id
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("x-request-id"); len(vals) > 0 {
				ctx = mw.WithValue(ctx, mw.RequestIDKey, vals[0])
			}
		}

		resp, err := handler(ctx, req)

		code := status.Code(err)
		fields := []logger.Field{
			logger.String("method", info.FullMethod),
			logger.Duration("latency", time.Since(start)),
			logger.String("code", code.String()),
		}
		if err != nil {
			fields = append(fields, logger.Err(err))
		}

		// trace_id/span_id 由 logger 自动从 ctx 提取
		switch {
		case code == codes.OK:
			log.Info(ctx, "gRPC request", fields...)
		case code == codes.Unknown || code == codes.Internal:
			log.Error(ctx, "gRPC request", fields...)
		default:
			log.Warn(ctx, "gRPC request", fields...)
		}

		return resp, err
	}
}

// UnaryRecovery 一元调用 panic 恢复拦截器
func UnaryRecovery(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error(ctx, "gRPC panic recovered",
					logger.Any("panic", r),
					logger.String("stack", string(debug.Stack())),
					logger.String("method", info.FullMethod),
				)
				err = status.Errorf(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}

// StreamLogging 流式调用日志拦截器
func StreamLogging(log logger.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		ctx := ss.Context()

		err := handler(srv, ss)

		code := status.Code(err)
		fields := []logger.Field{
			logger.String("method", info.FullMethod),
			logger.Duration("latency", time.Since(start)),
			logger.String("code", code.String()),
			logger.Bool("client_stream", info.IsClientStream),
			logger.Bool("server_stream", info.IsServerStream),
		}
		if err != nil {
			fields = append(fields, logger.Err(err))
		}

		log.Info(ctx, "gRPC stream", fields...)
		return err
	}
}

// StreamRecovery 流式调用 panic 恢复拦截器
func StreamRecovery(log logger.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error(ss.Context(), "gRPC stream panic recovered",
					logger.Any("panic", r),
					logger.String("stack", string(debug.Stack())),
					logger.String("method", info.FullMethod),
				)
				err = status.Errorf(codes.Internal, "internal error")
			}
		}()
		return handler(srv, ss)
	}
}
