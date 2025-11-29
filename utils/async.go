package utils

import (
	"context"
	"errors"
	"sync"
	"time"
)

// --- 重试 ---

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
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

// Retry 重试执行函数
func Retry(ctx context.Context, fn func() error, cfg RetryConfig) error {
	var lastErr error
	delay := cfg.Delay

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			delay = time.Duration(float64(delay) * cfg.Multiplier)
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}
		}

		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}

	return lastErr
}

// RetryWithResult 重试执行带返回值的函数
func RetryWithResult[T any](ctx context.Context, fn func() (T, error), cfg RetryConfig) (T, error) {
	var result T
	var lastErr error
	delay := cfg.Delay

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(delay):
			}
			delay = time.Duration(float64(delay) * cfg.Multiplier)
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}
		}

		var err error
		result, err = fn()
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	return result, lastErr
}

// --- 并发执行 ---

// Parallel 并行执行多个函数
func Parallel(fns ...func() error) error {
	var wg sync.WaitGroup
	errs := make([]error, len(fns))

	for i, fn := range fns {
		wg.Add(1)
		go func(idx int, f func() error) {
			defer wg.Done()
			errs[idx] = f()
		}(i, fn)
	}

	wg.Wait()
	return errors.Join(errs...)
}

// ParallelWithContext 带 context 的并行执行
func ParallelWithContext(ctx context.Context, fns ...func(context.Context) error) error {
	var wg sync.WaitGroup
	errs := make([]error, len(fns))

	for i, fn := range fns {
		wg.Add(1)
		go func(idx int, f func(context.Context) error) {
			defer wg.Done()
			errs[idx] = f(ctx)
		}(i, fn)
	}

	wg.Wait()
	return errors.Join(errs...)
}

// --- 超时执行 ---

// WithTimeout 带超时执行
func WithTimeout[T any](ctx context.Context, timeout time.Duration, fn func(context.Context) (T, error)) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultCh := make(chan T, 1)
	errCh := make(chan error, 1)

	go func() {
		result, err := fn(ctx)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- result
	}()

	select {
	case result := <-resultCh:
		return result, nil
	case err := <-errCh:
		var zero T
		return zero, err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

// --- 防抖/节流 ---

// Debounce 防抖
func Debounce(fn func(), delay time.Duration) func() {
	var timer *time.Timer
	var mu sync.Mutex

	return func() {
		mu.Lock()
		defer mu.Unlock()

		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(delay, fn)
	}
}

// Throttle 节流
func Throttle(fn func(), interval time.Duration) func() {
	var lastTime time.Time
	var mu sync.Mutex

	return func() {
		mu.Lock()
		defer mu.Unlock()

		now := time.Now()
		if now.Sub(lastTime) >= interval {
			lastTime = now
			fn()
		}
	}
}
