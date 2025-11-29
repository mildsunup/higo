package memory

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"higo/mq"
)

// Client 内存 MQ 客户端（用于测试和开发）
type Client struct {
	*mq.Base
	mu          sync.RWMutex
	topics      map[string][]chan *mq.Message
	handlers    map[string]mq.Handler
	subscribers map[string]context.CancelFunc
	closed      atomic.Bool
}

// New 创建内存 MQ 客户端
func New(name string) *Client {
	if name == "" {
		name = "memory"
	}
	return &Client{
		Base:        mq.NewBase(name, mq.TypeMemory),
		topics:      make(map[string][]chan *mq.Message),
		handlers:    make(map[string]mq.Handler),
		subscribers: make(map[string]context.CancelFunc),
	}
}

func (c *Client) Connect(ctx context.Context) error {
	if !c.CompareAndSwapState(mq.StateDisconnected, mq.StateConnecting) {
		return fmt.Errorf("memory mq: invalid state for connect")
	}
	c.SetState(mq.StateConnected)
	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	if c.State() != mq.StateConnected {
		return fmt.Errorf("memory mq: not connected")
	}
	return nil
}

func (c *Client) Publish(ctx context.Context, topic string, value []byte, opts ...mq.PublishOption) (*mq.PublishResult, error) {
	if c.closed.Load() {
		return nil, fmt.Errorf("memory mq: client closed")
	}

	var options mq.PublishOptions
	for _, opt := range opts {
		opt(&options)
	}

	msg := &mq.Message{
		ID:        uuid.New().String(),
		Topic:     topic,
		Key:       options.Key,
		Value:     value,
		Headers:   options.Headers,
		Timestamp: time.Now(),
	}

	c.mu.RLock()
	channels := c.topics[topic]
	c.mu.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- msg:
		default:
			// 通道满了，跳过
		}
	}

	c.IncPublished()
	return &mq.PublishResult{
		MessageID: msg.ID,
		Partition: 0,
		Offset:    c.Stats().Published,
	}, nil
}

func (c *Client) PublishAsync(ctx context.Context, topic string, value []byte, callback func(*mq.PublishResult, error), opts ...mq.PublishOption) {
	go func() {
		result, err := c.Publish(ctx, topic, value, opts...)
		if callback != nil {
			callback(result, err)
		}
	}()
}

func (c *Client) Subscribe(ctx context.Context, topic string, handler mq.Handler, opts ...mq.SubscribeOption) error {
	if c.closed.Load() {
		return fmt.Errorf("memory mq: client closed")
	}

	options := mq.DefaultSubscribeOptions()
	for _, opt := range opts {
		opt(&options)
	}

	ch := make(chan *mq.Message, 100)

	c.mu.Lock()
	c.topics[topic] = append(c.topics[topic], ch)
	c.handlers[topic] = handler
	c.mu.Unlock()

	subCtx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	c.subscribers[topic] = cancel
	c.mu.Unlock()

	// 启动消费协程
	for i := 0; i < options.Concurrency; i++ {
		go c.consume(subCtx, topic, ch, handler, options)
	}

	return nil
}

func (c *Client) consume(ctx context.Context, topic string, ch chan *mq.Message, handler mq.Handler, opts mq.SubscribeOptions) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			if msg == nil {
				return
			}

			var err error
			for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
				if attempt > 0 {
					c.IncRetries()
					time.Sleep(opts.RetryDelay)
				}

				err = handler(ctx, msg)
				if err == nil {
					break
				}
			}

			if err != nil {
				c.IncErrors()
			} else {
				c.IncConsumed()
			}
		}
	}
}

func (c *Client) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cancel, ok := c.subscribers[topic]; ok {
		cancel()
		delete(c.subscribers, topic)
	}
	delete(c.handlers, topic)
	delete(c.topics, topic)
	return nil
}

func (c *Client) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}

	c.SetState(mq.StateDisconnecting)

	c.mu.Lock()
	for _, cancel := range c.subscribers {
		cancel()
	}
	c.subscribers = make(map[string]context.CancelFunc)
	c.handlers = make(map[string]mq.Handler)
	c.topics = make(map[string][]chan *mq.Message)
	c.mu.Unlock()

	c.SetState(mq.StateDisconnected)
	return nil
}

var _ mq.Client = (*Client)(nil)
