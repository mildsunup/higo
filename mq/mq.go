package mq

import (
	"context"
	"sync/atomic"
)

// Producer 生产者接口
type Producer interface {
	// Publish 发布消息
	Publish(ctx context.Context, topic string, value []byte, opts ...PublishOption) (*PublishResult, error)
	// PublishAsync 异步发布
	PublishAsync(ctx context.Context, topic string, value []byte, callback func(*PublishResult, error), opts ...PublishOption)
	// Close 关闭生产者
	Close() error
}

// Consumer 消费者接口
type Consumer interface {
	// Subscribe 订阅主题
	Subscribe(ctx context.Context, topic string, handler Handler, opts ...SubscribeOption) error
	// Unsubscribe 取消订阅
	Unsubscribe(topic string) error
	// Close 关闭消费者
	Close() error
}

// Client MQ 客户端接口（同时支持生产和消费）
type Client interface {
	Producer
	Consumer
	// Connect 建立连接
	Connect(ctx context.Context) error
	// Ping 健康检查
	Ping(ctx context.Context) error
	// Name 客户端名称
	Name() string
	// Type 队列类型
	Type() Type
	// State 连接状态
	State() State
	// Stats 统计信息
	Stats() Stats
}

// Base 客户端基类
type Base struct {
	name  string
	typ   Type
	state atomic.Int32
	stats Stats
}

// NewBase 创建基类
func NewBase(name string, typ Type) *Base {
	b := &Base{name: name, typ: typ}
	b.state.Store(int32(StateDisconnected))
	return b
}

func (b *Base) Name() string { return b.name }
func (b *Base) Type() Type   { return b.typ }
func (b *Base) State() State { return State(b.state.Load()) }
func (b *Base) Stats() Stats { return b.stats }

func (b *Base) SetState(s State)                           { b.state.Store(int32(s)) }
func (b *Base) CompareAndSwapState(old, new State) bool    { return b.state.CompareAndSwap(int32(old), int32(new)) }
func (b *Base) IncPublished()                              { atomic.AddInt64(&b.stats.Published, 1) }
func (b *Base) IncConsumed()                               { atomic.AddInt64(&b.stats.Consumed, 1) }
func (b *Base) IncErrors()                                 { atomic.AddInt64(&b.stats.Errors, 1) }
func (b *Base) IncRetries()                                { atomic.AddInt64(&b.stats.Retries, 1) }

// Manager MQ 管理器接口
type Manager interface {
	// Register 注册客户端
	Register(client Client) error
	// Get 获取客户端
	Get(name string) (Client, bool)
	// GetByType 按类型获取
	GetByType(typ Type) []Client
	// ConnectAll 连接所有客户端
	ConnectAll(ctx context.Context) error
	// CloseAll 关闭所有客户端
	CloseAll(ctx context.Context) error
	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) []HealthStatus
	// List 列出所有客户端
	List() []string
}
