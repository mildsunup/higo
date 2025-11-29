package storage

import (
	"context"
	"time"
)

// Type 存储类型
type Type string

const (
	TypeMySQL         Type = "mysql"
	TypeRedis         Type = "redis"
	TypeMongoDB       Type = "mongodb"
	TypeClickHouse    Type = "clickhouse"
	TypeElasticsearch Type = "elasticsearch"
	TypePostgreSQL    Type = "postgresql"
	TypeUnknown       Type = "unknown"
)

// State 连接状态
type State int32

const (
	StateDisconnected State = iota
	StateConnecting
	StateConnected
	StateDisconnecting
)

func (s State) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateDisconnecting:
		return "disconnecting"
	}
	return "unknown"
}

// Storage 存储接口
type Storage interface {
	// Connect 建立连接
	Connect(ctx context.Context) error
	// Ping 健康检查
	Ping(ctx context.Context) error
	// Close 关闭连接
	Close(ctx context.Context) error
	// Name 存储标识
	Name() string
	// Type 存储类型
	Type() Type
	// State 连接状态
	State() State
}

// Stats 连接池统计
type Stats struct {
	MaxOpenConnections int           `json:"max_open_connections"`
	OpenConnections    int           `json:"open_connections"`
	InUse              int           `json:"in_use"`
	Idle               int           `json:"idle"`
	WaitCount          int64         `json:"wait_count"`
	WaitDuration       time.Duration `json:"wait_duration"`
}

// StatsProvider 提供连接池统计的存储
type StatsProvider interface {
	Stats() Stats
}

// HealthStatus 健康状态
type HealthStatus struct {
	Name      string        `json:"name"`
	Type      Type          `json:"type"`
	State     string        `json:"state"`
	Healthy   bool          `json:"healthy"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
	Stats     *Stats        `json:"stats,omitempty"`
	CheckedAt time.Time     `json:"checked_at"`
}

// Config 基础配置
type Config struct {
	Name            string        `json:"name" yaml:"name"`
	ConnectTimeout  time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
	ReadTimeout     time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout" yaml:"write_timeout"`
	MaxRetries      int           `json:"max_retries" yaml:"max_retries"`
	EnableTracing   bool          `json:"enable_tracing" yaml:"enable_tracing"`
	EnableMetrics   bool          `json:"enable_metrics" yaml:"enable_metrics"`
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		ConnectTimeout: 10 * time.Second,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxRetries:     3,
	}
}
