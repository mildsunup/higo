// Package eventbus 提供领域事件发布能力
package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"higo/mq"
)

// Event 领域事件接口
type Event interface {
	EventName() string
}

// Publisher 事件发布者接口
type Publisher interface {
	Publish(ctx context.Context, event any) error
	PublishAsync(ctx context.Context, event any) // 异步发布，不等待结果
}

// Envelope 事件信封
type Envelope struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Payload   any       `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

// Config 配置
type Config struct {
	Topic     string             // 事件主题，默认 "domain-events"
	IDFunc    func() string      // ID 生成函数
	TopicFunc func(Event) string // 按事件类型路由主题
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		Topic: "domain-events",
	}
}

// Bus 事件总线
type Bus struct {
	producer mq.Producer
	config   Config
}

// New 创建事件总线
func New(producer mq.Producer, cfg Config) *Bus {
	if cfg.Topic == "" {
		cfg.Topic = "domain-events"
	}
	return &Bus{producer: producer, config: cfg}
}

// Publish 同步发布事件
func (b *Bus) Publish(ctx context.Context, event any) error {
	data, topic, err := b.prepare(event)
	if err != nil {
		return err
	}
	_, err = b.producer.Publish(ctx, topic, data)
	return err
}

// PublishAsync 异步发布事件（不阻塞调用方）
func (b *Bus) PublishAsync(ctx context.Context, event any) {
	data, topic, err := b.prepare(event)
	if err != nil {
		return
	}
	b.producer.PublishAsync(ctx, topic, data, nil)
}

func (b *Bus) prepare(event any) ([]byte, string, error) {
	envelope := Envelope{
		Payload:   event,
		Timestamp: time.Now(),
	}

	if e, ok := event.(Event); ok {
		envelope.Name = e.EventName()
	} else {
		envelope.Name = fmt.Sprintf("%T", event)
	}

	if b.config.IDFunc != nil {
		envelope.ID = b.config.IDFunc()
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return nil, "", fmt.Errorf("eventbus: marshal failed: %w", err)
	}

	topic := b.config.Topic
	if b.config.TopicFunc != nil {
		if e, ok := event.(Event); ok {
			topic = b.config.TopicFunc(e)
		}
	}

	return data, topic, nil
}

// Noop 空实现
type Noop struct{}

func (Noop) Publish(ctx context.Context, event any) error { return nil }
func (Noop) PublishAsync(ctx context.Context, event any)  {}

var _ Publisher = (*Bus)(nil)
var _ Publisher = (*Noop)(nil)
