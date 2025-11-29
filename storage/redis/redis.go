package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"

	"higo/storage"
)

// Config Redis 配置
type Config struct {
	Name         string `json:"name" yaml:"name"`
	Addr         string `json:"addr" yaml:"addr"`
	Password     string `json:"password" yaml:"password"`
	DB           int    `json:"db" yaml:"db"`
	MaxRetries   int    `json:"max_retries" yaml:"max_retries"`
	PoolSize     int    `json:"pool_size" yaml:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns" yaml:"min_idle_conns"`
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	Tracer       trace.TracerProvider
}

// Storage Redis 存储
type Storage struct {
	*storage.Base
	client *redis.Client
	config Config
}

// New 创建 Redis 存储
func New(cfg Config) *Storage {
	name := cfg.Name
	if name == "" {
		name = "redis"
	}
	return &Storage{
		Base:   storage.NewBase(name, storage.TypeRedis),
		config: cfg,
	}
}

func (s *Storage) Connect(ctx context.Context) error {
	if !s.CompareAndSwapState(storage.StateDisconnected, storage.StateConnecting) {
		return fmt.Errorf("redis: invalid state for connect")
	}

	opts := &redis.Options{
		Addr:         s.config.Addr,
		Password:     s.config.Password,
		DB:           s.config.DB,
		MaxRetries:   s.config.MaxRetries,
		PoolSize:     s.config.PoolSize,
		MinIdleConns: s.config.MinIdleConns,
		DialTimeout:  s.config.DialTimeout,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	client := redis.NewClient(opts)

	// OpenTelemetry 追踪
	if s.config.Tracer != nil {
		if err := redisotel.InstrumentTracing(client,
			redisotel.WithTracerProvider(s.config.Tracer),
		); err != nil {
			s.SetState(storage.StateDisconnected)
			return fmt.Errorf("redis: setup tracing failed: %w", err)
		}
	}

	if err := client.Ping(ctx).Err(); err != nil {
		s.SetState(storage.StateDisconnected)
		return fmt.Errorf("redis: ping failed: %w", err)
	}

	s.client = client
	s.SetState(storage.StateConnected)
	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	if s.client == nil {
		return fmt.Errorf("redis: not connected")
	}
	return s.client.Ping(ctx).Err()
}

func (s *Storage) Close(ctx context.Context) error {
	if s.client == nil {
		return nil
	}
	s.SetState(storage.StateDisconnecting)
	err := s.client.Close()
	s.SetState(storage.StateDisconnected)
	s.client = nil
	return err
}

// Client 返回 Redis 客户端
func (s *Storage) Client() *redis.Client { return s.client }

// Stats 返回连接池统计
func (s *Storage) Stats() storage.Stats {
	if s.client == nil {
		return storage.Stats{}
	}
	stats := s.client.PoolStats()
	return storage.Stats{
		MaxOpenConnections: s.config.PoolSize,
		OpenConnections:    int(stats.TotalConns),
		Idle:               int(stats.IdleConns),
		InUse:              int(stats.TotalConns - stats.IdleConns),
		WaitCount:          int64(stats.Timeouts),
	}
}

var (
	_ storage.Storage       = (*Storage)(nil)
	_ storage.StatsProvider = (*Storage)(nil)
)
