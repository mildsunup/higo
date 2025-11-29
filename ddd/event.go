package ddd

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DomainEvent 领域事件接口
type DomainEvent interface {
	EventID() string
	EventName() string
	OccurredAt() time.Time
	AggregateID() string
	AggregateType() string
}

// EventBase 事件基类
type EventBase struct {
	id            string
	name          string
	occurredAt    time.Time
	aggregateID   string
	aggregateType string
}

// NewEventBase 创建事件基类
func NewEventBase(name, aggregateID, aggregateType string) EventBase {
	return EventBase{
		id:            uuid.New().String(),
		name:          name,
		occurredAt:    time.Now(),
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
	}
}

func (e EventBase) EventID() string       { return e.id }
func (e EventBase) EventName() string     { return e.name }
func (e EventBase) OccurredAt() time.Time { return e.occurredAt }
func (e EventBase) AggregateID() string   { return e.aggregateID }
func (e EventBase) AggregateType() string { return e.aggregateType }

// EventHandler 事件处理器
type EventHandler func(ctx context.Context, event DomainEvent) error

// EventPublisher 事件发布者接口
type EventPublisher interface {
	Publish(ctx context.Context, events ...DomainEvent) error
}

// EventSubscriber 事件订阅者接口
type EventSubscriber interface {
	Subscribe(eventName string, handler EventHandler)
}

// EventBus 事件总线接口（组合发布和订阅）
type EventBus interface {
	EventPublisher
	EventSubscriber
	Close() error
}
