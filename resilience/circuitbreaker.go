package resilience

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// CircuitState 熔断器状态
type CircuitState int32

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
	OnStateChange    func(from, to CircuitState)
}

// DefaultCircuitBreakerConfig 默认配置
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          30 * time.Second,
	}
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	cfg         CircuitBreakerConfig
	state       atomic.Int32
	failures    atomic.Int32
	successes   atomic.Int32
	lastFailure atomic.Int64
	mu          sync.Mutex
}

// CircuitBreakerOption 熔断器选项
type CircuitBreakerOption func(*CircuitBreaker)

// WithFailureThreshold 设置失败阈值
func WithFailureThreshold(threshold int) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.cfg.FailureThreshold = threshold
	}
}

// WithSuccessThreshold 设置成功阈值
func WithSuccessThreshold(threshold int) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.cfg.SuccessThreshold = threshold
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.cfg.Timeout = timeout
	}
}

// WithStateChangeHandler 设置状态变更回调
func WithStateChangeHandler(handler func(from, to CircuitState)) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.cfg.OnStateChange = handler
	}
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(opts ...CircuitBreakerOption) *CircuitBreaker {
	cb := &CircuitBreaker{
		cfg: CircuitBreakerConfig{
			FailureThreshold: 5,
			SuccessThreshold: 2,
			Timeout:          30 * time.Second,
		},
	}
	cb.state.Store(int32(CircuitClosed))

	for _, opt := range opts {
		opt(cb)
	}

	return cb
}

// State 获取当前状态
func (cb *CircuitBreaker) State() CircuitState {
	return CircuitState(cb.state.Load())
}

// Execute 执行函数
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	if err := cb.allow(); err != nil {
		return err
	}

	err := fn(ctx)

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}

	return err
}

func (cb *CircuitBreaker) allow() error {
	state := cb.State()

	switch state {
	case CircuitClosed:
		return nil
	case CircuitOpen:
		if time.Now().UnixNano()-cb.lastFailure.Load() > int64(cb.cfg.Timeout) {
			cb.transition(CircuitHalfOpen)
			return nil
		}
		return ErrCircuitOpen
	case CircuitHalfOpen:
		return nil
	}
	return nil
}

func (cb *CircuitBreaker) recordFailure() {
	state := cb.State()
	switch state {
	case CircuitClosed:
		if cb.failures.Add(1) >= int32(cb.cfg.FailureThreshold) {
			cb.transition(CircuitOpen)
		}
	case CircuitHalfOpen:
		cb.transition(CircuitOpen)
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	state := cb.State()
	switch state {
	case CircuitClosed:
		cb.failures.Store(0)
	case CircuitHalfOpen:
		if cb.successes.Add(1) >= int32(cb.cfg.SuccessThreshold) {
			cb.transition(CircuitClosed)
		}
	}
}

func (cb *CircuitBreaker) transition(to CircuitState) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	from := cb.State()
	if from == to {
		return
	}

	cb.state.Store(int32(to))
	cb.failures.Store(0)
	cb.successes.Store(0)

	if to == CircuitOpen {
		cb.lastFailure.Store(time.Now().UnixNano())
	}

	if cb.cfg.OnStateChange != nil {
		cb.cfg.OnStateChange(from, to)
	}
}

var _ Executor = (*CircuitBreaker)(nil)
