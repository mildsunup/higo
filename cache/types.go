package cache

import (
	"context"
	"errors"
	"time"
)

// 预定义错误
var (
	ErrNotFound     = errors.New("cache: key not found")
	ErrNil          = errors.New("cache: nil value")
	ErrTypeMismatch = errors.New("cache: type mismatch")
)

// Cache 缓存接口
type Cache interface {
	// Get 获取缓存
	Get(ctx context.Context, key string, dest any) error
	// Set 设置缓存
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	// Delete 删除缓存
	Delete(ctx context.Context, keys ...string) error
	// Exists 检查是否存在
	Exists(ctx context.Context, key string) (bool, error)
	// Close 关闭缓存
	Close() error
}

// BatchCache 批量操作接口
type BatchCache interface {
	Cache
	// MGet 批量获取
	MGet(ctx context.Context, keys []string) (map[string][]byte, error)
	// MSet 批量设置
	MSet(ctx context.Context, items map[string]any, ttl time.Duration) error
	// MDelete 批量删除
	MDelete(ctx context.Context, keys []string) error
}

// LoadFunc 加载函数
type LoadFunc func(ctx context.Context) (any, error)

// Loader 带加载功能的缓存
type Loader interface {
	Cache
	// GetOrLoad 获取或加载
	GetOrLoad(ctx context.Context, key string, dest any, loader LoadFunc, ttl time.Duration) error
}

// Stats 缓存统计
type Stats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	HitRate    float64 `json:"hit_rate"`
	Keys       int64 `json:"keys"`
	Size       int64 `json:"size"`
}

// StatsProvider 统计提供者
type StatsProvider interface {
	Stats() Stats
}

// Serializer 序列化器接口
type Serializer interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

// Option 缓存选项
type Option func(*Options)

// Options 缓存配置
type Options struct {
	Prefix       string
	Serializer   Serializer
	DefaultTTL   time.Duration
	NullTTL      time.Duration // 空值缓存时间（防穿透）
	EnableStats  bool
}

// WithPrefix 设置键前缀
func WithPrefix(prefix string) Option {
	return func(o *Options) { o.Prefix = prefix }
}

// WithSerializer 设置序列化器
func WithSerializer(s Serializer) Option {
	return func(o *Options) { o.Serializer = s }
}

// WithDefaultTTL 设置默认 TTL
func WithDefaultTTL(ttl time.Duration) Option {
	return func(o *Options) { o.DefaultTTL = ttl }
}

// WithNullTTL 设置空值缓存时间
func WithNullTTL(ttl time.Duration) Option {
	return func(o *Options) { o.NullTTL = ttl }
}

// WithStats 启用统计
func WithStats() Option {
	return func(o *Options) { o.EnableStats = true }
}

func defaultOptions() Options {
	return Options{
		Serializer: &JSONSerializer{},
		DefaultTTL: 5 * time.Minute,
		NullTTL:    time.Minute,
	}
}
