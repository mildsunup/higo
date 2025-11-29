package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/mildsunup/higo/mq"
)

// Config RabbitMQ 配置
type Config struct {
	Name            string        `json:"name" yaml:"name"`
	URL             string        `json:"url" yaml:"url"` // amqp://user:pass@host:port/vhost
	ExchangeName    string        `json:"exchange_name" yaml:"exchange_name"`
	ExchangeType    string        `json:"exchange_type" yaml:"exchange_type"` // direct, fanout, topic, headers
	Durable         bool          `json:"durable" yaml:"durable"`
	AutoDelete      bool          `json:"auto_delete" yaml:"auto_delete"`
	PrefetchCount   int           `json:"prefetch_count" yaml:"prefetch_count"`
	ReconnectDelay  time.Duration `json:"reconnect_delay" yaml:"reconnect_delay"`
	PublishTimeout  time.Duration `json:"publish_timeout" yaml:"publish_timeout"`
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		ExchangeType:   "topic",
		Durable:        true,
		PrefetchCount:  10,
		ReconnectDelay: 5 * time.Second,
		PublishTimeout: 5 * time.Second,
	}
}

// Client RabbitMQ 客户端
type Client struct {
	*mq.Base
	config  Config
	conn    *amqp.Connection
	channel *amqp.Channel

	mu          sync.RWMutex
	subscribers map[string]context.CancelFunc
	closed      chan struct{}
}

// New 创建 RabbitMQ 客户端
func New(cfg Config) *Client {
	name := cfg.Name
	if name == "" {
		name = "rabbitmq"
	}
	if cfg.ExchangeType == "" {
		cfg.ExchangeType = "topic"
	}
	if cfg.PrefetchCount == 0 {
		cfg.PrefetchCount = 10
	}
	if cfg.PublishTimeout == 0 {
		cfg.PublishTimeout = 5 * time.Second
	}

	return &Client{
		Base:        mq.NewBase(name, mq.TypeRabbitMQ),
		config:      cfg,
		subscribers: make(map[string]context.CancelFunc),
		closed:      make(chan struct{}),
	}
}

func (c *Client) Connect(ctx context.Context) error {
	if !c.CompareAndSwapState(mq.StateDisconnected, mq.StateConnecting) {
		return fmt.Errorf("rabbitmq: invalid state for connect")
	}

	conn, err := amqp.Dial(c.config.URL)
	if err != nil {
		c.SetState(mq.StateDisconnected)
		return fmt.Errorf("rabbitmq: dial failed: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		c.SetState(mq.StateDisconnected)
		return fmt.Errorf("rabbitmq: create channel failed: %w", err)
	}

	// 声明 Exchange
	if c.config.ExchangeName != "" {
		if err := channel.ExchangeDeclare(
			c.config.ExchangeName,
			c.config.ExchangeType,
			c.config.Durable,
			c.config.AutoDelete,
			false, // internal
			false, // no-wait
			nil,
		); err != nil {
			channel.Close()
			conn.Close()
			c.SetState(mq.StateDisconnected)
			return fmt.Errorf("rabbitmq: declare exchange failed: %w", err)
		}
	}

	// 设置 QoS
	if err := channel.Qos(c.config.PrefetchCount, 0, false); err != nil {
		channel.Close()
		conn.Close()
		c.SetState(mq.StateDisconnected)
		return fmt.Errorf("rabbitmq: set qos failed: %w", err)
	}

	c.conn = conn
	c.channel = channel
	c.SetState(mq.StateConnected)

	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	if c.conn == nil || c.conn.IsClosed() {
		return fmt.Errorf("rabbitmq: not connected")
	}
	return nil
}

func (c *Client) Publish(ctx context.Context, topic string, value []byte, opts ...mq.PublishOption) (*mq.PublishResult, error) {
	if c.channel == nil {
		return nil, fmt.Errorf("rabbitmq: channel not initialized")
	}

	var options mq.PublishOptions
	for _, opt := range opts {
		opt(&options)
	}

	msg := amqp.Publishing{
		ContentType:  "application/octet-stream",
		Body:         value,
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
	}

	if options.Key != "" {
		msg.MessageId = options.Key
	}

	if options.Headers != nil {
		msg.Headers = make(amqp.Table)
		for k, v := range options.Headers {
			msg.Headers[k] = v
		}
	}

	// 使用带超时的 context
	pubCtx, cancel := context.WithTimeout(ctx, c.config.PublishTimeout)
	defer cancel()

	routingKey := topic
	exchange := c.config.ExchangeName

	if err := c.channel.PublishWithContext(
		pubCtx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		msg,
	); err != nil {
		c.IncErrors()
		return nil, fmt.Errorf("rabbitmq: publish failed: %w", err)
	}

	c.IncPublished()
	return &mq.PublishResult{
		MessageID: msg.MessageId,
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
	if c.channel == nil {
		return fmt.Errorf("rabbitmq: channel not initialized")
	}

	options := mq.DefaultSubscribeOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// 队列名称
	queueName := topic
	if options.Group != "" {
		queueName = fmt.Sprintf("%s.%s", topic, options.Group)
	}

	// 声明队列
	queue, err := c.channel.QueueDeclare(
		queueName,
		c.config.Durable,
		c.config.AutoDelete,
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("rabbitmq: declare queue failed: %w", err)
	}

	// 绑定队列到 Exchange
	if c.config.ExchangeName != "" {
		if err := c.channel.QueueBind(
			queue.Name,
			topic, // routing key
			c.config.ExchangeName,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("rabbitmq: bind queue failed: %w", err)
		}
	}

	// 开始消费
	deliveries, err := c.channel.Consume(
		queue.Name,
		"",             // consumer tag
		options.AutoAck,
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("rabbitmq: consume failed: %w", err)
	}

	subCtx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	c.subscribers[topic] = cancel
	c.mu.Unlock()

	// 启动消费协程
	for i := 0; i < options.Concurrency; i++ {
		go c.consume(subCtx, deliveries, handler, options)
	}

	return nil
}

func (c *Client) consume(ctx context.Context, deliveries <-chan amqp.Delivery, handler mq.Handler, opts mq.SubscribeOptions) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.closed:
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}

			msg := &mq.Message{
				ID:        d.MessageId,
				Topic:     d.RoutingKey,
				Value:     d.Body,
				Headers:   make(map[string]string),
				Timestamp: d.Timestamp,
				Raw:       &d,
			}

			for k, v := range d.Headers {
				if s, ok := v.(string); ok {
					msg.Headers[k] = s
				}
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
				if !opts.AutoAck {
					_ = d.Nack(false, true) // requeue
				}
			} else {
				c.IncConsumed()
				if !opts.AutoAck {
					_ = d.Ack(false)
				}
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
	return nil
}

func (c *Client) Close() error {
	c.SetState(mq.StateDisconnecting)

	close(c.closed)

	c.mu.Lock()
	for _, cancel := range c.subscribers {
		cancel()
	}
	c.subscribers = make(map[string]context.CancelFunc)
	c.mu.Unlock()

	var errs []error

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	c.SetState(mq.StateDisconnected)

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

var _ mq.Client = (*Client)(nil)
