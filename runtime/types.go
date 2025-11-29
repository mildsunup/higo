package runtime

import (
	"context"
	"time"

	"github.com/mildsunup/higo/logger"
)

// Component 组件接口
type Component interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// HealthChecker 健康检查接口
type HealthChecker interface {
	Health(ctx context.Context) error
}

// Hook 生命周期钩子
type Hook func(ctx context.Context) error

// State 应用状态
type State int

const (
	StateCreated State = iota
	StateStarting
	StateRunning
	StateStopping
	StateStopped
	StateFailed
)

func (s State) String() string {
	switch s {
	case StateCreated:
		return "created"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateStopped:
		return "stopped"
	case StateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// Config 应用配置
type Config struct {
	Name            string        `yaml:"name" mapstructure:"name"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" mapstructure:"shutdown_timeout"`
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		Name:            "app",
		ShutdownTimeout: 30 * time.Second,
	}
}

// Logger 日志接口
type Logger = logger.Logger

// Application 应用接口
type Application interface {
	// Run 启动应用并阻塞直到收到关闭信号
	Run(ctx context.Context) error
	// Start 启动应用
	Start(ctx context.Context) error
	// Stop 停止应用
	Stop(ctx context.Context) error
	// State 获取当前状态
	State() State
	// Health 健康检查
	Health(ctx context.Context) error
}
