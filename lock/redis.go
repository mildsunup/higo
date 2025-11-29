package lock

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Redis Lua 脚本
var (
	// 释放锁：只有持有者才能释放
	unlockScript = redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		end
		return 0
	`)

	// 刷新锁：只有持有者才能刷新
	refreshScript = redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("pexpire", KEYS[1], ARGV[2])
		end
		return 0
	`)
)

// RedisLocker Redis 锁工厂
type RedisLocker struct {
	client redis.Cmdable
	prefix string
}

// NewRedisLocker 创建 Redis 锁工厂
func NewRedisLocker(client redis.Cmdable, prefix string) *RedisLocker {
	if prefix == "" {
		prefix = "lock:"
	}
	return &RedisLocker{client: client, prefix: prefix}
}

func (l *RedisLocker) NewLock(key string, opts ...Option) Lock {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	if options.Token == "" {
		options.Token = uuid.New().String()
	}
	return &redisLock{
		client: l.client,
		key:    l.prefix + key,
		opts:   options,
	}
}

type redisLock struct {
	client redis.Cmdable
	key    string
	opts   Options
}

func (l *redisLock) Lock(ctx context.Context) error {
	retries := 0
	for {
		ok, err := l.TryLock(ctx)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

		retries++
		if l.opts.RetryCount >= 0 && retries > l.opts.RetryCount {
			return ErrLockFailed
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(l.opts.RetryDelay):
		}
	}
}

func (l *redisLock) TryLock(ctx context.Context) (bool, error) {
	return l.client.SetNX(ctx, l.key, l.opts.Token, l.opts.TTL).Result()
}

func (l *redisLock) Unlock(ctx context.Context) error {
	result, err := unlockScript.Run(ctx, l.client, []string{l.key}, l.opts.Token).Int64()
	if err != nil {
		return err
	}
	if result == 0 {
		return ErrNotHeld
	}
	return nil
}

func (l *redisLock) Refresh(ctx context.Context) error {
	result, err := refreshScript.Run(ctx, l.client, []string{l.key}, l.opts.Token, l.opts.TTL.Milliseconds()).Int64()
	if err != nil {
		return err
	}
	if result == 0 {
		return ErrLockExpired
	}
	return nil
}

var _ Locker = (*RedisLocker)(nil)
var _ Lock = (*redisLock)(nil)
