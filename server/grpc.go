package server

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// GRPCOptions gRPC 服务器配置
type GRPCOptions struct {
	Options
	MaxRecvMsgSize   int
	MaxSendMsgSize   int
	EnableReflection bool
	Interceptors     []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor
}

// DefaultGRPCOptions 默认 gRPC 配置
func DefaultGRPCOptions() GRPCOptions {
	return GRPCOptions{
		Options:          DefaultOptions(),
		MaxRecvMsgSize:   4 << 20, // 4MB
		MaxSendMsgSize:   4 << 20,
		EnableReflection: true,
	}
}

// GRPCOption gRPC 配置选项
type GRPCOption func(*GRPCOptions)

// WithGRPCAddr 设置地址
func WithGRPCAddr(addr string) GRPCOption {
	return func(o *GRPCOptions) { o.Addr = addr }
}

// WithMaxRecvMsgSize 设置最大接收消息大小
func WithMaxRecvMsgSize(size int) GRPCOption {
	return func(o *GRPCOptions) { o.MaxRecvMsgSize = size }
}

// WithMaxSendMsgSize 设置最大发送消息大小
func WithMaxSendMsgSize(size int) GRPCOption {
	return func(o *GRPCOptions) { o.MaxSendMsgSize = size }
}

// WithReflection 启用 reflection
func WithReflection(enable bool) GRPCOption {
	return func(o *GRPCOptions) { o.EnableReflection = enable }
}

// WithUnaryInterceptor 添加 Unary 拦截器
func WithUnaryInterceptor(i grpc.UnaryServerInterceptor) GRPCOption {
	return func(o *GRPCOptions) { o.Interceptors = append(o.Interceptors, i) }
}

// WithStreamInterceptor 添加 Stream 拦截器
func WithStreamInterceptor(i grpc.StreamServerInterceptor) GRPCOption {
	return func(o *GRPCOptions) { o.StreamInterceptors = append(o.StreamInterceptors, i) }
}

// GRPCServer gRPC 服务器
type GRPCServer struct {
	opts         GRPCOptions
	server       *grpc.Server
	healthServer *health.Server
	listener     net.Listener
	mu           sync.Mutex
	running      bool
}

// NewGRPCServer 创建 gRPC 服务器
func NewGRPCServer(opts ...GRPCOption) *GRPCServer {
	o := DefaultGRPCOptions()
	for _, opt := range opts {
		opt(&o)
	}

	serverOpts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(o.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(o.MaxSendMsgSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              2 * time.Hour,
			Timeout:           20 * time.Second,
		}),
	}

	if len(o.Interceptors) > 0 {
		serverOpts = append(serverOpts, grpc.ChainUnaryInterceptor(o.Interceptors...))
	}
	if len(o.StreamInterceptors) > 0 {
		serverOpts = append(serverOpts, grpc.ChainStreamInterceptor(o.StreamInterceptors...))
	}

	grpcServer := grpc.NewServer(serverOpts...)

	// 健康检查
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	// Reflection
	if o.EnableReflection {
		reflection.Register(grpcServer)
	}

	return &GRPCServer{
		opts:         o,
		server:       grpcServer,
		healthServer: healthServer,
	}
}

// Name 服务器名称
func (s *GRPCServer) Name() string { return s.opts.Name }

// Addr 监听地址
func (s *GRPCServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.opts.Addr
}

// Server 返回底层 gRPC Server
func (s *GRPCServer) Server() *grpc.Server { return s.server }

// HealthServer 返回健康检查服务器
func (s *GRPCServer) HealthServer() *health.Server { return s.healthServer }

// RegisterService 注册服务
func (s *GRPCServer) RegisterService(fn func(*grpc.Server)) {
	fn(s.server)
}

// Start 启动服务器
func (s *GRPCServer) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return errors.New("server already running")
	}

	listener, err := net.Listen("tcp", s.opts.Addr)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	s.listener = listener
	s.running = true
	s.mu.Unlock()

	// 设置健康状态
	s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// 非阻塞启动
	go func() {
		_ = s.server.Serve(listener)
	}()

	return nil
}

// Stop 停止服务器
func (s *GRPCServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}
	s.running = false

	// 设置健康状态
	s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)

	// 优雅关闭
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		return ctx.Err()
	}
}
