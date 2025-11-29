package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/mildsunup/higo/observability"
)

// Metrics gRPC 指标
type Metrics struct {
	RequestsTotal   observability.Counter
	RequestDuration observability.Histogram
	ActiveRequests  observability.Gauge
}

// NewMetrics 创建 gRPC 指标
func NewMetrics(p observability.MetricsProvider) *Metrics {
	return &Metrics{
		RequestsTotal:   p.Counter("grpc_requests_total", "Total gRPC requests", "method", "code"),
		RequestDuration: p.Histogram("grpc_request_duration_seconds", "gRPC request duration", observability.DurationBuckets, "method"),
		ActiveRequests:  p.Gauge("grpc_active_requests", "Active gRPC requests"),
	}
}

// UnaryInterceptor 一元调用指标拦截器
func (m *Metrics) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		m.ActiveRequests.Inc()
		defer m.ActiveRequests.Dec()

		resp, err := handler(ctx, req)

		m.RequestsTotal.Inc(info.FullMethod, status.Code(err).String())
		m.RequestDuration.Since(start, info.FullMethod)
		return resp, err
	}
}

// StreamInterceptor 流式调用指标拦截器
func (m *Metrics) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		m.ActiveRequests.Inc()
		defer m.ActiveRequests.Dec()

		err := handler(srv, ss)

		m.RequestsTotal.Inc(info.FullMethod, status.Code(err).String())
		m.RequestDuration.Since(start, info.FullMethod)
		return err
	}
}
