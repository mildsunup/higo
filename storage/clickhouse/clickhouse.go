package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/mildsunup/higo/storage"
)

// Config ClickHouse 配置
type Config struct {
	Name            string        `json:"name" yaml:"name"`
	Addr            string        `json:"addr" yaml:"addr"`
	Database        string        `json:"database" yaml:"database"`
	Username        string        `json:"username" yaml:"username"`
	Password        string        `json:"password" yaml:"password"`
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	DialTimeout     time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
}

// Storage ClickHouse 存储
type Storage struct {
	*storage.Base
	conn   driver.Conn
	config Config
}

// New 创建 ClickHouse 存储
func New(cfg Config) *Storage {
	name := cfg.Name
	if name == "" {
		name = "clickhouse"
	}
	return &Storage{
		Base:   storage.NewBase(name, storage.TypeClickHouse),
		config: cfg,
	}
}

func (s *Storage) Connect(ctx context.Context) error {
	if !s.CompareAndSwapState(storage.StateDisconnected, storage.StateConnecting) {
		return fmt.Errorf("clickhouse: invalid state for connect")
	}

	opts := &clickhouse.Options{
		Addr: []string{s.config.Addr},
		Auth: clickhouse.Auth{
			Database: s.config.Database,
			Username: s.config.Username,
			Password: s.config.Password,
		},
	}

	if s.config.MaxOpenConns > 0 {
		opts.MaxOpenConns = s.config.MaxOpenConns
	}
	if s.config.ConnMaxLifetime > 0 {
		opts.ConnMaxLifetime = s.config.ConnMaxLifetime
	}
	if s.config.DialTimeout > 0 {
		opts.DialTimeout = s.config.DialTimeout
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		s.SetState(storage.StateDisconnected)
		return fmt.Errorf("clickhouse: connect failed: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		s.SetState(storage.StateDisconnected)
		return fmt.Errorf("clickhouse: ping failed: %w", err)
	}

	s.conn = conn
	s.SetState(storage.StateConnected)
	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	if s.conn == nil {
		return fmt.Errorf("clickhouse: not connected")
	}
	return s.conn.Ping(ctx)
}

func (s *Storage) Close(ctx context.Context) error {
	if s.conn == nil {
		return nil
	}
	s.SetState(storage.StateDisconnecting)
	err := s.conn.Close()
	s.SetState(storage.StateDisconnected)
	s.conn = nil
	return err
}

// Conn 返回 ClickHouse 连接
func (s *Storage) Conn() driver.Conn { return s.conn }

var _ storage.Storage = (*Storage)(nil)
