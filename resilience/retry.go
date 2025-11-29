package resilience

import (
	"context"
	"time"
)

// Retry 重试执行器
type Retry struct {
	cfg RetryConfig
}

// NewRetry 创建重试执行器
func NewRetry(cfg RetryConfig) *Retry {
	if cfg.RetryIf == nil {
		cfg.RetryIf = func(err error) bool { return err != nil }
	}
	return &Retry{cfg: cfg}
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
