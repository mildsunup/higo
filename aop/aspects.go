package aop

import (
	"errors"
	"time"

	"go.uber.org/zap"
)

// Logging 日志切面
func Logging(logger *zap.Logger) Interceptor {
	return InterceptorFunc(func(inv *Invocation) error {
		start := time.Now()
		err := inv.Proceed()
		logger.Debug("method executed",
			zap.String("method", inv.Method()),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)
		return err
	})
}

// Metrics 指标切面
func Metrics(recorder func(method string, duration time.Duration, success bool)) Interceptor {
	return InterceptorFunc(func(inv *Invocation) error {
		start := time.Now()
		err := inv.Proceed()
		recorder(inv.Method(), time.Since(start), err == nil)
		return err
	})
}

// Recovery 恢复切面
func Recovery(onPanic func(r any)) Interceptor {
	return InterceptorFunc(func(inv *Invocation) (err error) {
		defer func() {
			if r := recover(); r != nil {
				if onPanic != nil {
					onPanic(r)
				}
				err = errors.New("panic recovered")
				inv.SetError(err)
			}
		}()
		return inv.Proceed()
	})
}

// Timeout 超时切面
func Timeout(d time.Duration) Interceptor {
	return InterceptorFunc(func(inv *Invocation) error {
		done := make(chan error, 1)
		go func() { done <- inv.Proceed() }()
		select {
		case err := <-done:
			return err
		case <-time.After(d):
			inv.SetError(ErrTimeout)
			return ErrTimeout
		}
	})
}

// ErrTimeout 超时错误
var ErrTimeout = errors.New("operation timeout")
