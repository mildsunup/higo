package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"

	"github.com/soheilhy/cmux"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

// MultiplexServer 多协议复用服务器 (HTTP + gRPC 同端口)
type MultiplexServer struct {
	opts       Options
	listener   net.Listener
	mux        cmux.CMux
	httpServer *http.Server
	grpcServer *grpc.Server
	httpHandler http.Handler
	mu         sync.Mutex
	running    bool
	wg         sync.WaitGroup
}

// NewMultiplexServer 创建多协议复用服务器
func NewMultiplexServer(httpHandler http.Handler, grpcServer *grpc.Server, opts ...Option) *MultiplexServer {
	o := ApplyOptions(opts...)
	o.Name = "multiplex-server"
	return &MultiplexServer{
		opts:        o,
		httpHandler: httpHandler,
		grpcServer:  grpcServer,
	}
}

// Name 服务器名称
func (s *MultiplexServer) Name() string { return s.opts.Name }

// Addr 监听地址
func (s *MultiplexServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.opts.Addr
}

// Start 启动服务器
func (s *MultiplexServer) Start(ctx context.Context) error {
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
	s.mux = cmux.New(listener)

	// HTTP/2 with H2C
	h2s := &http2.Server{}
	s.httpServer = &http.Server{
		Handler:        h2c.NewHandler(s.httpHandler, h2s),
		ReadTimeout:    s.opts.ReadTimeout,
		WriteTimeout:   s.opts.WriteTimeout,
		IdleTimeout:    s.opts.IdleTimeout,
		MaxHeaderBytes: s.opts.MaxHeaderBytes,
	}

	s.running = true
	s.mu.Unlock()

	// gRPC 匹配
	grpcL := s.mux.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
	)
	// HTTP 匹配 (其他所有)
	httpL := s.mux.Match(cmux.Any())

	s.wg.Add(3)

	// 启动 gRPC
	go func() {
		defer s.wg.Done()
		_ = s.grpcServer.Serve(grpcL)
	}()

	// 启动 HTTP
	go func() {
		defer s.wg.Done()
		_ = s.httpServer.Serve(httpL)
	}()

	// 启动 cmux
	go func() {
		defer s.wg.Done()
		_ = s.mux.Serve()
	}()

	return nil
}

// Stop 停止服务器
func (s *MultiplexServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	// 优雅关闭 gRPC
	s.grpcServer.GracefulStop()

	// 优雅关闭 HTTP
	shutdownCtx, cancel := context.WithTimeout(ctx, s.opts.ShutdownTimeout)
	defer cancel()
	_ = s.httpServer.Shutdown(shutdownCtx)

	// 关闭监听器
	_ = s.listener.Close()

	// 等待所有 goroutine 退出
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
