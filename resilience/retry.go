package resilience

import (
	"context"
	"time"
)

// Retry 重试执行器
type Retry struct {
	cfg RetryConfig
}

// RetryOption 重试选项
type RetryOption func(*Retry)

// WithMaxAttempts 设置最大重试次数
func WithMaxAttempts(attempts int) RetryOption {
	return func(r *Retry) {
		r.cfg.MaxAttempts = attempts
	}
}

// WithDelay 设置初始延迟
func WithDelay(delay time.Duration) RetryOption {
	return func(r *Retry) {
		r.cfg.Delay = delay
	}
}

// WithMaxDelay 设置最大延迟
func WithMaxDelay(maxDelay time.Duration) RetryOption {
	return func(r *Retry) {
		r.cfg.MaxDelay = maxDelay
	}
}

// WithMultiplier 设置延迟倍数
func WithMultiplier(multiplier float64) RetryOption {
	return func(r *Retry) {
		r.cfg.Multiplier = multiplier
	}
}

// WithRetryIf 设置重试条件
func WithRetryIf(fn func(error) bool) RetryOption {
	return func(r *Retry) {
		r.cfg.RetryIf = fn
	}
}

// NewRetry 创建重试执行器
func NewRetry(opts ...RetryOption) *Retry {
	r := &Retry{
		cfg: RetryConfig{
			MaxAttempts: 3,
			Delay:       100 * time.Millisecond,
			MaxDelay:    5 * time.Second,
			Multiplier:  2.0,
			RetryIf:     func(err error) bool { return err != nil },
		},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *Retry) Execute(ctx context.Context, fn func(context.Context) error) error {
	var lastErr error
	delay := r.cfg.Delay

	for attempt := 0; attempt < r.cfg.MaxAttempts; attempt++ {
		if err := fn(ctx); err == nil {
			return nil
		} else {
			lastErr = err
			if !r.cfg.RetryIf(err) {
				return err
			}
		}

		if attempt < r.cfg.MaxAttempts-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}

			delay = time.Duration(float64(delay) * r.cfg.Multiplier)
			if delay > r.cfg.MaxDelay {
				delay = r.cfg.MaxDelay
			}
		}
	}

	return lastErr
}

var _ Executor = (*Retry)(nil)
