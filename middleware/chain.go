package middleware

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// GinChain Gin 中间件链构建器
type GinChain struct {
	middlewares []gin.HandlerFunc
}

// NewGinChain 创建 Gin 中间件链
func NewGinChain(middlewares ...gin.HandlerFunc) *GinChain {
	return &GinChain{middlewares: middlewares}
}

// Use 添加中间件
func (c *GinChain) Use(middlewares ...gin.HandlerFunc) *GinChain {
	c.middlewares = append(c.middlewares, middlewares...)
	return c
}

// Then 返回最终的中间件列表
func (c *GinChain) Then() []gin.HandlerFunc {
	return c.middlewares
}

// Apply 应用到 gin.Engine
func (c *GinChain) Apply(r *gin.Engine) {
	for _, m := range c.middlewares {
		r.Use(m)
	}
}

// GRPCChain gRPC 拦截器链构建器
type GRPCChain struct {
	unary  []grpc.UnaryServerInterceptor
	stream []grpc.StreamServerInterceptor
}

// NewGRPCChain 创建 gRPC 拦截器链
func NewGRPCChain() *GRPCChain {
	return &GRPCChain{}
}

// UseUnary 添加一元拦截器
func (c *GRPCChain) UseUnary(interceptors ...grpc.UnaryServerInterceptor) *GRPCChain {
	c.unary = append(c.unary, interceptors...)
	return c
}

// UseStream 添加流式拦截器
func (c *GRPCChain) UseStream(interceptors ...grpc.StreamServerInterceptor) *GRPCChain {
	c.stream = append(c.stream, interceptors...)
	return c
}

// UnaryInterceptors 返回一元拦截器列表
func (c *GRPCChain) UnaryInterceptors() []grpc.UnaryServerInterceptor {
	return c.unary
}

// StreamInterceptors 返回流式拦截器列表
func (c *GRPCChain) StreamInterceptors() []grpc.StreamServerInterceptor {
	return c.stream
}

// ServerOptions 返回 gRPC ServerOptions
func (c *GRPCChain) ServerOptions() []grpc.ServerOption {
	var opts []grpc.ServerOption
	if len(c.unary) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(c.unary...))
	}
	if len(c.stream) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(c.stream...))
	}
	return opts
}
