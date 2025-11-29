package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter 限流器配置
type RateLimiterConfig struct {
	Rate      rate.Limit    // 每秒请求数
	Burst     int           // 突发容量
	KeyFunc   func(*gin.Context) string // 限流键函数
	ExcludeFunc func(*gin.Context) bool // 排除函数
}

// DefaultRateLimiterConfig 默认配置
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Rate:    100,
		Burst:   200,
		KeyFunc: func(c *gin.Context) string { return c.ClientIP() },
	}
}

// RateLimiter 限流中间件
func RateLimiter(cfg RateLimiterConfig) gin.HandlerFunc {
	var (
		limiters = make(map[string]*rate.Limiter)
		mu       sync.RWMutex
	)

	getLimiter := func(key string) *rate.Limiter {
		mu.RLock()
		limiter, exists := limiters[key]
		mu.RUnlock()

		if exists {
			return limiter
		}

		mu.Lock()
		defer mu.Unlock()

		// 双重检查
		if limiter, exists = limiters[key]; exists {
			return limiter
		}

		limiter = rate.NewLimiter(cfg.Rate, cfg.Burst)
		limiters[key] = limiter
		return limiter
	}

	// 定期清理
	go func() {
		ticker := time.NewTicker(time.Hour)
		for range ticker.C {
			mu.Lock()
			limiters = make(map[string]*rate.Limiter)
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		// 检查是否排除
		if cfg.ExcludeFunc != nil && cfg.ExcludeFunc(c) {
			c.Next()
			return
		}

		key := cfg.KeyFunc(c)
		limiter := getLimiter(key)

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "Too Many Requests",
			})
			return
		}

		c.Next()
	}
}
