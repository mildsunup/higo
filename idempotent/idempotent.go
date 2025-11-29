// Package idempotent 提供幂等性支持
package idempotent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"higo/cache"
)

var (
	ErrDuplicateRequest = errors.New("duplicate request")
	ErrKeyRequired      = errors.New("idempotency key required")
)

// Status 请求状态
type Status int

const (
	StatusPending Status = iota
	StatusCompleted
)

// Record 幂等记录
type Record struct {
	Status      Status `json:"status"`
	Response    []byte `json:"response,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

// Config 配置
type Config struct {
	TTL              time.Duration
	CheckFingerprint bool
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		TTL:              24 * time.Hour,
		CheckFingerprint: true,
	}
}

// Handler 幂等处理器
type Handler struct {
	cache  cache.Cache
	config Config
}

// New 创建处理器
func New(c cache.Cache, cfg Config) *Handler {
	return &Handler{cache: c, config: cfg}
}

// Result 执行结果
type Result struct {
	Cached   bool
	Response []byte
}

// Execute 执行幂等操作
func (h *Handler) Execute(ctx context.Context, key string, fingerprint string, fn func() ([]byte, error)) (*Result, error) {
	if key == "" {
		return nil, ErrKeyRequired
	}

	// 检查是否已存在
	var record Record
	err := h.cache.Get(ctx, key, &record)
	if err == nil {
		// 校验指纹
		if h.config.CheckFingerprint && fingerprint != "" && record.Fingerprint != fingerprint {
			return nil, ErrDuplicateRequest
		}
		if record.Status == StatusCompleted {
			return &Result{Cached: true, Response: record.Response}, nil
		}
		return nil, ErrDuplicateRequest
	}
	if !errors.Is(err, cache.ErrNotFound) {
		return nil, err
	}

	// 设置为处理中
	pending := Record{Status: StatusPending, Fingerprint: fingerprint}
	if err := h.cache.Set(ctx, key, pending, h.config.TTL); err != nil {
		return nil, err
	}

	// 执行业务逻辑
	response, err := fn()
	if err != nil {
		_ = h.cache.Delete(ctx, key)
		return nil, err
	}

	// 设置为已完成
	completed := Record{Status: StatusCompleted, Response: response, Fingerprint: fingerprint}
	if err := h.cache.Set(ctx, key, completed, h.config.TTL); err != nil {
		return nil, err
	}

	return &Result{Cached: false, Response: response}, nil
}

// Fingerprint 生成请求指纹
func Fingerprint(method, path string, body []byte) string {
	h := sha256.New()
	h.Write([]byte(method))
	h.Write([]byte(path))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// MustMarshal JSON 序列化
func MustMarshal(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
