package cache

import (
	"context"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

// MultiLevel 多级缓存
type MultiLevel struct {
	l1     Cache // 本地缓存
	l2     Cache // 远程缓存
	opts   Options
	group  singleflight.Group
	hits   atomic.Int64
	misses atomic.Int64
}

// MultiLevelConfig 多级缓存配置
type MultiLevelConfig struct {
	L1 Cache
	L2 Cache
}

// NewMultiLevel 创建多级缓存
func NewMultiLevel(cfg MultiLevelConfig, opts ...Option) *MultiLevel {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	return &MultiLevel{
		l1:   cfg.L1,
		l2:   cfg.L2,
		opts: o,
	}
}

// NewMultiLevelCache 便捷构造器
func NewMultiLevelCache(redisAddr, password string, db int) (*MultiLevel, error) {
	l1, err := NewMemory(DefaultMemoryConfig())
	if err != nil {
		return nil, err
	}

	l2 := NewRedis(RedisConfig{
		Addr:     redisAddr,
		Password: password,
		DB:       db,
	})

	return NewMultiLevel(MultiLevelConfig{L1: l1, L2: l2}), nil
}

func (m *MultiLevel) Get(ctx context.Context, key string, dest any) error {
	// 1. 尝试 L1
	if err := m.l1.Get(ctx, key, dest); err == nil {
		m.hits.Add(1)
		return nil
	}

	// 2. 尝试 L2
	if err := m.l2.Get(ctx, key, dest); err != nil {
		m.misses.Add(1)
		return err
	}

	m.hits.Add(1)

	// 3. 回填 L1
	_ = m.l1.Set(ctx, key, dest, m.opts.DefaultTTL)

	return nil
}

func (m *MultiLevel) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if ttl == 0 {
		ttl = m.opts.DefaultTTL
	}

	// 先写 L2（持久化）
	if err := m.l2.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// 再写 L1
	return m.l1.Set(ctx, key, value, ttl)
}

func (m *MultiLevel) Delete(ctx context.Context, keys ...string) error {
	// 先删 L1
	_ = m.l1.Delete(ctx, keys...)
	// 再删 L2
	return m.l2.Delete(ctx, keys...)
}

func (m *MultiLevel) Exists(ctx context.Context, key string) (bool, error) {
	if exists, _ := m.l1.Exists(ctx, key); exists {
		return true, nil
	}
	return m.l2.Exists(ctx, key)
}

func (m *MultiLevel) Close() error {
	_ = m.l1.Close()
	return m.l2.Close()
}

// GetOrLoad 获取或加载（防击穿）
func (m *MultiLevel) GetOrLoad(ctx context.Context, key string, dest any, loader LoadFunc, ttl time.Duration) error {
	// 先尝试获取
	if err := m.Get(ctx, key, dest); err == nil {
		return nil
	}

	// 使用 singleflight 防止缓存击穿
	val, err, _ := m.group.Do(key, func() (any, error) {
		// 双重检查
		if err := m.Get(ctx, key, dest); err == nil {
			return dest, nil
		}

		// 加载数据
		v, err := loader(ctx)
		if err != nil {
			return nil, err
		}

		// 写入缓存
		if v != nil {
			_ = m.Set(ctx, key, v, ttl)
		} else if m.opts.NullTTL > 0 {
			// 空值缓存（防穿透）
			_ = m.Set(ctx, key, struct{}{}, m.opts.NullTTL)
		}

		return v, nil
	})

	if err != nil {
		return err
	}

	// 复制结果到 dest
	if val != nil {
		data, _ := m.opts.Serializer.Marshal(val)
		return m.opts.Serializer.Unmarshal(data, dest)
	}

	return nil
}

func (m *MultiLevel) Stats() Stats {
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

// L1 返回 L1 缓存
func (m *MultiLevel) L1() Cache { return m.l1 }

// L2 返回 L2 缓存
func (m *MultiLevel) L2() Cache { return m.l2 }

var (
	_ Cache         = (*MultiLevel)(nil)
	_ Loader        = (*MultiLevel)(nil)
	_ StatsProvider = (*MultiLevel)(nil)
)
