package mq

import (
	"context"
	"time"
)

// Type 消息队列类型
type Type string

const (
	TypeKafka    Type = "kafka"
	TypeRabbitMQ Type = "rabbitmq"
	TypeRocketMQ Type = "rocketmq"
	TypePulsar   Type = "pulsar"
	TypeNSQ      Type = "nsq"
	TypeRedis    Type = "redis"
	TypeMemory   Type = "memory"
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

// Message 消息
type Message struct {
	ID        string            `json:"id"`
	Topic     string            `json:"topic"`
	Key       string            `json:"key,omitempty"`
	Value     []byte            `json:"value"`
	Headers   map[string]string `json:"headers,omitempty"`
	Timestamp time.Time         `json:"timestamp"`

	// 追踪信息
	TraceID string `json:"trace_id,omitempty"`
	SpanID  string `json:"span_id,omitempty"`

	// Raw 原始消息（用于 Ack/Nack）
	Raw any `json:"-"`
}

// PublishResult 发布结果
type PublishResult struct {
	MessageID string
	Partition int32
	Offset    int64
}

// ConsumeResult 消费结果
type ConsumeResult struct {
	Ack   func() error
	Nack  func() error
	Retry func(delay time.Duration) error
}

// Handler 消息处理器
type Handler func(ctx context.Context, msg *Message) error

// HealthStatus 健康状态
type HealthStatus struct {
	Name      string        `json:"name"`
	Type      Type          `json:"type"`
	State     string        `json:"state"`
	Healthy   bool          `json:"healthy"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
	CheckedAt time.Time     `json:"checked_at"`
}

// Stats 统计信息
type Stats struct {
	Published     int64 `json:"published"`
	Consumed      int64 `json:"consumed"`
	Errors        int64 `json:"errors"`
	Retries       int64 `json:"retries"`
	PendingCount  int64 `json:"pending_count"`
	ConsumerCount int   `json:"consumer_count"`
}

// PublishOption 发布选项
type PublishOption func(*PublishOptions)

// PublishOptions 发布选项结构
type PublishOptions struct {
	Key       string
	Headers   map[string]string
	Partition *int32
	Delay     time.Duration
}

// WithKey 设置消息 Key
func WithKey(key string) PublishOption {
	return func(o *PublishOptions) { o.Key = key }
}

// WithHeaders 设置消息头
func WithHeaders(headers map[string]string) PublishOption {
	return func(o *PublishOptions) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		for k, v := range headers {
			o.Headers[k] = v
		}
	}
}

// WithPartition 指定分区
func WithPartition(p int32) PublishOption {
	return func(o *PublishOptions) { o.Partition = &p }
}

// WithDelay 延迟发送
func WithDelay(d time.Duration) PublishOption {
	return func(o *PublishOptions) { o.Delay = d }
}

// SubscribeOption 订阅选项
type SubscribeOption func(*SubscribeOptions)

// SubscribeOptions 订阅选项结构
type SubscribeOptions struct {
	Group       string
	Concurrency int
	AutoAck     bool
	MaxRetries  int
	RetryDelay  time.Duration
}

// WithGroup 设置消费组
func WithGroup(group string) SubscribeOption {
	return func(o *SubscribeOptions) { o.Group = group }
}

// WithConcurrency 设置并发数
func WithConcurrency(n int) SubscribeOption {
	return func(o *SubscribeOptions) { o.Concurrency = n }
}

// WithAutoAck 自动确认
func WithAutoAck(auto bool) SubscribeOption {
	return func(o *SubscribeOptions) { o.AutoAck = auto }
}

// WithMaxRetries 最大重试次数
func WithMaxRetries(n int) SubscribeOption {
	return func(o *SubscribeOptions) { o.MaxRetries = n }
}

// WithRetryDelay 重试延迟
func WithRetryDelay(d time.Duration) SubscribeOption {
	return func(o *SubscribeOptions) { o.RetryDelay = d }
}

// DefaultSubscribeOptions 默认订阅选项
func DefaultSubscribeOptions() SubscribeOptions {
	return SubscribeOptions{
		Concurrency: 1,
		AutoAck:     true,
		MaxRetries:  3,
		RetryDelay:  time.Second,
	}
}
