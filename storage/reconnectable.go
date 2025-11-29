package storage

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ReconnectConfig 重连配置
type ReconnectConfig struct {
	MaxRetries      int           // 最大重试次数
	InitialInterval time.Duration // 初始重试间隔
	MaxInterval     time.Duration // 最大重试间隔
	Multiplier      float64       // 退避乘数
}

// DefaultReconnectConfig 默认重连配置
func DefaultReconnectConfig() ReconnectConfig {
	return ReconnectConfig{
		MaxRetries:      5,
		InitialInterval: time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
	}
}

// Reconnectable 可重连存储装饰器
type Reconnectable struct {
	storage Storage
	config  ReconnectConfig
	mu      sync.RWMutex
}

// NewReconnectable 创建可重连存储
func NewReconnectable(s Storage, cfg ReconnectConfig) *Reconnectable {
	if cfg.Multiplier <= 0 {
		cfg.Multiplier = 2.0
	}
	return &Reconnectable{storage: s, config: cfg}
}

func (r *Reconnectable) Connect(ctx context.Context) error {
	return r.connectWithRetry(ctx)
}

func (r *Reconnectable) connectWithRetry(ctx context.Context) error {
	var lastErr error
	interval := r.config.InitialInterval

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(interval):
			}
			// 指数退避
			interval = time.Duration(float64(interval) * r.config.Multiplier)
			if interval > r.config.MaxInterval {
				interval = r.config.MaxInterval
			}
		}

		if err := r.storage.Connect(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}

	return fmt.Errorf("connect failed after %d attempts: %w", r.config.MaxRetries+1, lastErr)
}

func (r *Reconnectable) Ping(ctx context.Context) error {
	err := r.storage.Ping(ctx)
	if err != nil && isConnectionError(err) {
		// 尝试重连
		r.mu.Lock()
		defer r.mu.Unlock()

		_ = r.storage.Close(ctx)
		if reconnErr := r.connectWithRetry(ctx); reconnErr != nil {
			return fmt.Errorf("ping failed, reconnect failed: %w", reconnErr)
		}
		return r.storage.Ping(ctx)
	}
	return err
}

func (r *Reconnectable) Close(ctx context.Context) error {
	return r.storage.Close(ctx)
}

func (r *Reconnectable) Name() string  { return r.storage.Name() }
func (r *Reconnectable) Type() Type    { return r.storage.Type() }
func (r *Reconnectable) State() State  { return r.storage.State() }

// Unwrap 获取底层存储
func (r *Reconnectable) Unwrap() Storage { return r.storage }

// isConnectionError 判断是否为连接错误
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	keywords := []string{
		"connection refused", "connection reset", "broken pipe",
		"no such host", "timeout", "eof", "closed",
	}
	for _, kw := range keywords {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}

var _ Storage = (*Reconnectable)(nil)
