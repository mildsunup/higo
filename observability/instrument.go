package observability

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// GinMiddleware 返回 Gin 追踪中间件
func GinMiddleware(serviceName string, opts ...otelgin.Option) gin.HandlerFunc {
	return otelgin.Middleware(serviceName, opts...)
}

// GRPCServerOption 返回 gRPC 服务端追踪选项
func GRPCServerOption(opts ...otelgrpc.Option) grpc.ServerOption {
	return grpc.StatsHandler(otelgrpc.NewServerHandler(opts...))
}

// GRPCDialOption 返回 gRPC 客户端追踪选项
func GRPCDialOption(opts ...otelgrpc.Option) grpc.DialOption {
	return grpc.WithStatsHandler(otelgrpc.NewClientHandler(opts...))
}
