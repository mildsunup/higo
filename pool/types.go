package pool

import (
	"context"
	"time"
)

// Pool 对象池接口
type Pool[T any] interface {
	// Get 获取对象
	Get() T
	// Put 归还对象
	Put(obj T)
	// Stats 统计信息
	Stats() Stats
}

// Stats 池统计
type Stats struct {
	Gets      int64 `json:"gets"`
	Puts      int64 `json:"puts"`
	Hits      int64 `json:"hits"`
	Misses    int64 `json:"misses"`
	InUse     int64 `json:"in_use"`
	Idle      int64 `json:"idle"`
	MaxSize   int64 `json:"max_size"`
}

// Task 任务
type Task func()

// TaskWithContext 带 context 的任务
type TaskWithContext func(ctx context.Context)

// TaskPool 任务池接口
type TaskPool interface {
	// Submit 提交任务（非阻塞）
	Submit(task Task) error
	// SubmitWait 提交任务（阻塞等待）
	SubmitWait(ctx context.Context, task Task) error
	// Running 运行中的 worker 数
	Running() int
	// Pending 等待中的任务数
	Pending() int
	// Stop 停止池
	Stop()
	// StopWait 停止并等待所有任务完成
	StopWait(ctx context.Context) error
}

// Resetter 可重置对象接口
type Resetter interface {
	Reset()
}

// Closer 可关闭对象接口
type Closer interface {
	Close() error
}

// Option 池选项
type Option func(*Options)

// Options 池配置
type Options struct {
	MaxSize     int
	MinIdle     int
	MaxIdle     int
	IdleTimeout time.Duration
	MaxLifetime time.Duration
}

// WithMaxSize 设置最大容量
func WithMaxSize(n int) Option {
	return func(o *Options) { o.MaxSize = n }
}

// WithMinIdle 设置最小空闲数
func WithMinIdle(n int) Option {
	return func(o *Options) { o.MinIdle = n }
}

// WithMaxIdle 设置最大空闲数
func WithMaxIdle(n int) Option {
	return func(o *Options) { o.MaxIdle = n }
}

// WithIdleTimeout 设置空闲超时
func WithIdleTimeout(d time.Duration) Option {
	return func(o *Options) { o.IdleTimeout = d }
}

// WithMaxLifetime 设置最大生命周期
func WithMaxLifetime(d time.Duration) Option {
	return func(o *Options) { o.MaxLifetime = d }
}
