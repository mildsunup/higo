// Package lock 提供分布式锁支持
package lock

import (
	"context"
	"errors"
	"time"
)

var (
	ErrLockFailed  = errors.New("lock: failed to acquire lock")
	ErrLockExpired = errors.New("lock: lock expired")
	ErrNotHeld     = errors.New("lock: lock not held")
)

// Lock 分布式锁接口
type Lock interface {
	// Lock 获取锁，阻塞直到获取成功或 ctx 取消
	Lock(ctx context.Context) error
	// TryLock 尝试获取锁，立即返回
	TryLock(ctx context.Context) (bool, error)
	// Unlock 释放锁
	Unlock(ctx context.Context) error
	// Refresh 刷新锁的过期时间
	Refresh(ctx context.Context) error
}

// Locker 锁工厂接口
type Locker interface {
	// NewLock 创建锁
	NewLock(key string, opts ...Option) Lock
}

// Options 锁配置
type Options struct {
	TTL         time.Duration // 锁过期时间
	RetryDelay  time.Duration // 重试间隔
	RetryCount  int           // 重试次数，-1 表示无限重试
	Token       string        // 锁持有者标识
}

// Option 配置函数
type Option func(*Options)

// DefaultOptions 默认配置
func DefaultOptions() Options {
	return Options{
		TTL:        30 * time.Second,
		RetryDelay: 100 * time.Millisecond,
		RetryCount: -1,
	}
}

func WithTTL(ttl time.Duration) Option {
	return func(o *Options) { o.TTL = ttl }
}

func WithRetryDelay(d time.Duration) Option {
	return func(o *Options) { o.RetryDelay = d }
}

func WithRetryCount(n int) Option {
	return func(o *Options) { o.RetryCount = n }
}

func WithToken(token string) Option {
	return func(o *Options) { o.Token = token }
}

// Do 在锁保护下执行操作
func Do(ctx context.Context, l Lock, fn func(ctx context.Context) error) error {
	if err := l.Lock(ctx); err != nil {
		return err
	}
	defer l.Unlock(ctx)
	return fn(ctx)
}

// TryDo 尝试获取锁并执行，获取失败立即返回
func TryDo(ctx context.Context, l Lock, fn func(ctx context.Context) error) (bool, error) {
	ok, err := l.TryLock(ctx)
	if err != nil || !ok {
		return false, err
	}
	defer l.Unlock(ctx)
	return true, fn(ctx)
}
