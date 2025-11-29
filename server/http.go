package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// HTTPServer HTTP 服务器
type HTTPServer struct {
	opts     Options
	server   *http.Server
	handler  http.Handler
	listener net.Listener
	mu       sync.Mutex
	running  bool
}

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(handler http.Handler, opts ...Option) *HTTPServer {
	o := ApplyOptions(opts...)
	return &HTTPServer{
		opts:    o,
		handler: handler,
	}
}

// Name 服务器名称
func (s *HTTPServer) Name() string { return s.opts.Name }

// Addr 监听地址
func (s *HTTPServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.opts.Addr
}

// Start 启动服务器
func (s *HTTPServer) Start(ctx context.Context) error {
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

	// 支持 H2C (HTTP/2 Cleartext)
	h2s := &http2.Server{}
	s.server = &http.Server{
		Handler:        h2c.NewHandler(s.handler, h2s),
		ReadTimeout:    s.opts.ReadTimeout,
		WriteTimeout:   s.opts.WriteTimeout,
		IdleTimeout:    s.opts.IdleTimeout,
		MaxHeaderBytes: s.opts.MaxHeaderBytes,
	}

	s.running = true
	s.mu.Unlock()

	// 非阻塞启动
	go func() {
		if err := s.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// 记录错误但不阻塞
		}
	}()

	return nil
}

// Stop 停止服务器
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.server == nil {
		return nil
	}

	s.running = false

	shutdownCtx, cancel := context.WithTimeout(ctx, s.opts.ShutdownTimeout)
	defer cancel()

	return s.server.Shutdown(shutdownCtx)
}

// Handler 返回 HTTP Handler
func (s *HTTPServer) Handler() http.Handler {
	return s.handler
}
