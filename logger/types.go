package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"
)

// Level 日志级别
type Level int8

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	default:
		return "info"
	}
}

// ParseLevel 解析日志级别
func ParseLevel(s string) Level {
	switch s {
	case "debug":
		return DebugLevel
	case "info", "":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// Field 日志字段
type Field struct {
	Key   string
	Value any
}

// 字段构造函数
func String(key, val string) Field        { return Field{Key: key, Value: val} }
func Int(key string, val int) Field       { return Field{Key: key, Value: val} }
func Int64(key string, val int64) Field   { return Field{Key: key, Value: val} }
func Float64(key string, val float64) Field { return Field{Key: key, Value: val} }
func Bool(key string, val bool) Field     { return Field{Key: key, Value: val} }
func Duration(key string, val time.Duration) Field { return Field{Key: key, Value: val} }
func Time(key string, val time.Time) Field { return Field{Key: key, Value: val} }
func Any(key string, val any) Field       { return Field{Key: key, Value: val} }
func Err(err error) Field                 { return Field{Key: "error", Value: err} }

// Logger 日志接口
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)

	// With 返回带有预设字段的 Logger
	With(fields ...Field) Logger

	// WithLevel 返回指定级别的 Logger
	WithLevel(level Level) Logger

	// Sync 刷新缓冲区
	Sync() error
}

// Config 日志配置
type Config struct {
	Level      string `yaml:"level" mapstructure:"level"`             // debug, info, warn, error
	Format     string `yaml:"format" mapstructure:"format"`           // json, console
	Output     string `yaml:"output" mapstructure:"output"`           // stdout, stderr, filepath
	AddCaller  bool   `yaml:"add_caller" mapstructure:"add_caller"`   // 是否添加调用者信息
	CallerSkip int    `yaml:"caller_skip" mapstructure:"caller_skip"` // 调用栈跳过层数

	// 日志分割配置 (仅当 Output 为文件路径时生效)
	Rotation RotationConfig `yaml:"rotation" mapstructure:"rotation"`
}

// RotationConfig 日志分割配置
type RotationConfig struct {
	MaxSize    int  `yaml:"max_size" mapstructure:"max_size"`       // 单文件最大大小 (MB)，默认 100
	MaxBackups int  `yaml:"max_backups" mapstructure:"max_backups"` // 保留旧文件数量，默认 3
	MaxAge     int  `yaml:"max_age" mapstructure:"max_age"`         // 保留天数，默认 7
	Compress   bool `yaml:"compress" mapstructure:"compress"`       // 是否压缩旧文件，默认 true
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		AddCaller:  true,
		CallerSkip: 2,
		Rotation: RotationConfig{
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
	}
}

// ContextExtractor 从 context 提取字段的函数
type ContextExtractor func(ctx context.Context) []Field

// NewRequestID 生成请求 ID
func NewRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b)
}
