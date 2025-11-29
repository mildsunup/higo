// Package server 提供统一的服务器抽象
package server

import (
	"context"
	"time"
)

// Server 服务器接口
type Server interface {
	// Start 启动服务器（非阻塞）
	Start(ctx context.Context) error
	// Stop 停止服务器
	Stop(ctx context.Context) error
	// Name 服务器名称
	Name() string
	// Addr 监听地址
	Addr() string
}

// Runnable 可运行组件接口（兼容 runtime）
type Runnable interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Options 通用服务器配置
type Options struct {
	Name            string
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	MaxHeaderBytes  int
}

// DefaultOptions 默认配置
func DefaultOptions() Options {
	return Options{
		Name:            "server",
		Addr:            ":8080",
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		MaxHeaderBytes:  1 << 20, // 1MB
	}
}

// Option 配置选项函数
type Option func(*Options)

// WithName 设置服务器名称
func WithName(name string) Option {
	return func(o *Options) { o.Name = name }
}

// WithAddr 设置监听地址
func WithAddr(addr string) Option {
	return func(o *Options) { o.Addr = addr }
}

// WithReadTimeout 设置读超时
func WithReadTimeout(d time.Duration) Option {
	return func(o *Options) { o.ReadTimeout = d }
}

// WithWriteTimeout 设置写超时
func WithWriteTimeout(d time.Duration) Option {
	return func(o *Options) { o.WriteTimeout = d }
}

// WithShutdownTimeout 设置关闭超时
func WithShutdownTimeout(d time.Duration) Option {
	return func(o *Options) { o.ShutdownTimeout = d }
}

// ApplyOptions 应用配置选项
func ApplyOptions(opts ...Option) Options {
	o := DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
