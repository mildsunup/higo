package resilience

import (
	"context"
	"sync"
	"time"
)

// RateLimiter 限流器接口
type RateLimiter interface {
	Executor
	Allow() bool
	AllowN(n int) bool
}

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	mu         sync.Mutex
	rate       float64   // 每秒补充令牌数
	burst      float64   // 桶容量
	tokens     float64   // 当前令牌数
	lastUpdate time.Time // 上次更新时间
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(rate float64, burst int) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		burst:      float64(burst),
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastUpdate).Seconds()
	tb.tokens = min(tb.tokens+elapsed*tb.rate, tb.burst)
	tb.lastUpdate = now
}

func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

func (tb *TokenBucket) AllowN(n int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	need := float64(n)
	if tb.tokens < need {
		return false
	}
	tb.tokens -= need
	return true
}

func (tb *TokenBucket) Execute(ctx context.Context, fn func(context.Context) error) error {
	if !tb.Allow() {
		return ErrRateLimited
	}
	return fn(ctx)
}

// Tokens 返回当前令牌数（用于监控）
func (tb *TokenBucket) Tokens() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	return tb.tokens
}

var _ RateLimiter = (*TokenBucket)(nil)
