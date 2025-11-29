// Package security 安全相关功能
package security

import (
	"sync"
	"time"
)

// LoginRateLimiter 登录限流器（防暴力破解）
type LoginRateLimiter struct {
	mu       sync.RWMutex
	attempts map[string]*attemptInfo
	config   LoginRateLimitConfig
}

type attemptInfo struct {
	count     int
	firstAt   time.Time
	lockedAt  *time.Time
}

// LoginRateLimitConfig 配置
type LoginRateLimitConfig struct {
	MaxAttempts   int           // 最大尝试次数
	Window        time.Duration // 时间窗口
	LockDuration  time.Duration // 锁定时长
	CleanInterval time.Duration // 清理间隔
}

// DefaultLoginRateLimitConfig 默认配置
func DefaultLoginRateLimitConfig() LoginRateLimitConfig {
	return LoginRateLimitConfig{
		MaxAttempts:   5,
		Window:        15 * time.Minute,
		LockDuration:  30 * time.Minute,
		CleanInterval: 5 * time.Minute,
	}
}

// NewLoginRateLimiter 创建登录限流器
func NewLoginRateLimiter(cfg LoginRateLimitConfig) *LoginRateLimiter {
	rl := &LoginRateLimiter{
		attempts: make(map[string]*attemptInfo),
		config:   cfg,
	}
	go rl.cleanup()
	return rl
}

// Allow 检查是否允许登录尝试
func (rl *LoginRateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	info, exists := rl.attempts[key]

	if !exists {
		rl.attempts[key] = &attemptInfo{count: 0, firstAt: now}
		return true
	}

	// 检查是否被锁定
	if info.lockedAt != nil {
		if now.Sub(*info.lockedAt) < rl.config.LockDuration {
			return false
		}
		// 锁定已过期，重置
		info.count = 0
		info.lockedAt = nil
		info.firstAt = now
	}

	// 检查时间窗口
	if now.Sub(info.firstAt) > rl.config.Window {
		info.count = 0
		info.firstAt = now
	}

	return info.count < rl.config.MaxAttempts
}

// RecordFailure 记录失败尝试
func (rl *LoginRateLimiter) RecordFailure(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	info, exists := rl.attempts[key]
	if !exists {
		info = &attemptInfo{firstAt: time.Now()}
		rl.attempts[key] = info
	}

	info.count++
	if info.count >= rl.config.MaxAttempts {
		now := time.Now()
		info.lockedAt = &now
	}
}

// RecordSuccess 记录成功（清除记录）
func (rl *LoginRateLimiter) RecordSuccess(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, key)
}

// IsLocked 检查是否被锁定
func (rl *LoginRateLimiter) IsLocked(key string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	info, exists := rl.attempts[key]
	if !exists {
		return false
	}
	if info.lockedAt == nil {
		return false
	}
	return time.Since(*info.lockedAt) < rl.config.LockDuration
}

// RemainingAttempts 剩余尝试次数
func (rl *LoginRateLimiter) RemainingAttempts(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	info, exists := rl.attempts[key]
	if !exists {
		return rl.config.MaxAttempts
	}
	remaining := rl.config.MaxAttempts - info.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (rl *LoginRateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.CleanInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, info := range rl.attempts {
			// 清理过期的锁定记录
			if info.lockedAt != nil && now.Sub(*info.lockedAt) > rl.config.LockDuration {
				delete(rl.attempts, key)
				continue
			}
			// 清理过期的窗口记录
			if info.lockedAt == nil && now.Sub(info.firstAt) > rl.config.Window {
				delete(rl.attempts, key)
			}
		}
		rl.mu.Unlock()
	}
}
