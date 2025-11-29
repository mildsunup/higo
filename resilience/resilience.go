// Package resilience 提供弹性模式：熔断、限流、重试
package resilience

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
	ErrRateLimited = errors.New("rate limited")
)

// Executor 弹性执行器接口
type Executor interface {
	Execute(ctx context.Context, fn func(context.Context) error) error
}

// Chain 链式执行多个 Executor
func Chain(executors ...Executor) Executor {
	return &chainExecutor{executors: executors}
}

type chainExecutor struct {
	executors []Executor
}

func (c *chainExecutor) Execute(ctx context.Context, fn func(context.Context) error) error {
	if len(c.executors) == 0 {
		return fn(ctx)
	}

	wrapped := fn
	for i := len(c.executors) - 1; i >= 0; i-- {
		exec := c.executors[i]
		inner := wrapped
		wrapped = func(ctx context.Context) error {
			return exec.Execute(ctx, inner)
		}
	}
	return wrapped(ctx)
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
	RetryIf     func(error) bool
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Delay:       100 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
	}
}
