package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis Redis 缓存
type Redis struct {
	client *redis.Client
	opts   Options
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
}

// NewRedis 创建 Redis 缓存
func NewRedis(cfg RedisConfig, opts ...Option) *Redis {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	return &Redis{client: client, opts: o}
}

// NewRedisFromClient 从已有客户端创建
func NewRedisFromClient(client *redis.Client, opts ...Option) *Redis {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return &Redis{client: client, opts: o}
}

func (r *Redis) key(k string) string {
	if r.opts.Prefix != "" {
		return r.opts.Prefix + ":" + k
	}
	return k
}

func (r *Redis) Get(ctx context.Context, key string, dest any) error {
	data, err := r.client.Get(ctx, r.key(key)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrNotFound
		}
		return err
	}
	return r.opts.Serializer.Unmarshal(data, dest)
}

func (r *Redis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if ttl == 0 {
		ttl = r.opts.DefaultTTL
	}

	data, err := r.opts.Serializer.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, r.key(key), data, ttl).Err()
}

func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = r.key(k)
	}
	return r.client.Del(ctx, fullKeys...).Err()
}

func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, r.key(key)).Result()
	return n > 0, err
}

func (r *Redis) Close() error {
	return r.client.Close()
}

// MGet 批量获取
func (r *Redis) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = r.key(k)
	}

	vals, err := r.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte)
	for i, val := range vals {
		if val != nil {
			if s, ok := val.(string); ok {
				result[keys[i]] = []byte(s)
			}
		}
	}
	return result, nil
}

// MSet 批量设置
func (r *Redis) MSet(ctx context.Context, items map[string]any, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	for k, v := range items {
		data, err := r.opts.Serializer.Marshal(v)
		if err != nil {
			return err
		}
		pipe.Set(ctx, r.key(k), data, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// MDelete 批量删除
func (r *Redis) MDelete(ctx context.Context, keys []string) error {
	return r.Delete(ctx, keys...)
}

// Client 返回底层 Redis 客户端
func (r *Redis) Client() *redis.Client {
	return r.client
}

var (
	_ Cache      = (*Redis)(nil)
	_ BatchCache = (*Redis)(nil)
)
