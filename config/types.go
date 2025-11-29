package config

import (
	"context"
	"time"
)

// Provider 配置提供者接口
type Provider interface {
	// Name 提供者名称
	Name() string
	// Load 加载配置
	Load(ctx context.Context) (map[string]any, error)
	// Watch 监听配置变更（可选实现）
	Watch(ctx context.Context, onChange func()) error
}

// Decoder 配置解码器接口
type Decoder interface {
	Decode(data []byte, v any) error
}

// Encoder 配置编码器接口
type Encoder interface {
	Encode(v any) ([]byte, error)
}

// Validator 配置校验器接口
type Validator interface {
	Validate() error
}

// SecretResolver 敏感信息解析器
type SecretResolver interface {
	// Resolve 解析敏感信息，如 ${vault:secret/path} 或 ${env:VAR_NAME}
	Resolve(ctx context.Context, value string) (string, error)
}

// ChangeEvent 配置变更事件
type ChangeEvent struct {
	Key       string
	OldValue  any
	NewValue  any
	Timestamp time.Time
}

// ChangeListener 配置变更监听器
type ChangeListener func(event ChangeEvent)

// Option 配置选项
type Option func(*Options)

// Options 配置选项
type Options struct {
	Providers      []Provider
	SecretResolver SecretResolver
	WatchInterval  time.Duration
	OnChange       ChangeListener
}

// WithProvider 添加配置提供者
func WithProvider(p Provider) Option {
	return func(o *Options) {
		o.Providers = append(o.Providers, p)
	}
}

// WithSecretResolver 设置敏感信息解析器
func WithSecretResolver(r SecretResolver) Option {
	return func(o *Options) {
		o.SecretResolver = r
	}
}

// WithWatchInterval 设置监听间隔
func WithWatchInterval(d time.Duration) Option {
	return func(o *Options) {
		o.WatchInterval = d
	}
}

// WithOnChange 设置变更回调
func WithOnChange(fn ChangeListener) Option {
	return func(o *Options) {
		o.OnChange = fn
	}
}
