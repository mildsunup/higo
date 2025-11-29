package cache

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/ristretto"
)

// MemoryConfig 内存缓存配置
type MemoryConfig struct {
	MaxSize     int64 // 最大内存（字节）
	NumCounters int64 // 计数器数量
}

// DefaultMemoryConfig 默认配置
func DefaultMemoryConfig() MemoryConfig {
	return MemoryConfig{
		MaxSize:     256 << 20, // 256MB
		NumCounters: 1e7,
	}
}

// Memory 内存缓存
type Memory struct {
	cache  *ristretto.Cache
	opts   Options
	hits   atomic.Int64
	misses atomic.Int64
}

// NewMemory 创建内存缓存
func NewMemory(cfg MemoryConfig, opts ...Option) (*Memory, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: cfg.NumCounters,
		MaxCost:     cfg.MaxSize,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	return &Memory{cache: cache, opts: o}, nil
}

func (m *Memory) key(k string) string {
	if m.opts.Prefix != "" {
		return m.opts.Prefix + ":" + k
	}
	return k
}

func (m *Memory) Get(ctx context.Context, key string, dest any) error {
	val, found := m.cache.Get(m.key(key))
	if !found {
		m.misses.Add(1)
		return ErrNotFound
	}

	m.hits.Add(1)

	data, ok := val.([]byte)
	if !ok {
		return ErrTypeMismatch
	}

	return m.opts.Serializer.Unmarshal(data, dest)
}

func (m *Memory) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if ttl == 0 {
		ttl = m.opts.DefaultTTL
	}

	data, err := m.opts.Serializer.Marshal(value)
	if err != nil {
		return err
	}

	m.cache.SetWithTTL(m.key(key), data, int64(len(data)), ttl)
	return nil
}

func (m *Memory) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		m.cache.Del(m.key(key))
	}
	return nil
}

func (m *Memory) Exists(ctx context.Context, key string) (bool, error) {
	_, found := m.cache.Get(m.key(key))
	return found, nil
}

func (m *Memory) Close() error {
	m.cache.Close()
	return nil
}

func (m *Memory) Stats() Stats {
	hits := m.hits.Load()
	misses := m.misses.Load()
	total := hits + misses

	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	return Stats{
		Hits:    hits,
		Misses:  misses,
		HitRate: hitRate,
	}
}

var (
	_ Cache         = (*Memory)(nil)
	_ StatsProvider = (*Memory)(nil)
)
