package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"

	"github.com/mildsunup/higo/mq"
)

// Config Kafka 配置
type Config struct {
	Name            string        `json:"name" yaml:"name"`
	Brokers         []string      `json:"brokers" yaml:"brokers"`
	Version         string        `json:"version" yaml:"version"`
	ClientID        string        `json:"client_id" yaml:"client_id"`
	RequiredAcks    int           `json:"required_acks" yaml:"required_acks"`
	MaxRetries      int           `json:"max_retries" yaml:"max_retries"`
	RetryBackoff    time.Duration `json:"retry_backoff" yaml:"retry_backoff"`
	FlushFrequency  time.Duration `json:"flush_frequency" yaml:"flush_frequency"`
	FlushMessages   int           `json:"flush_messages" yaml:"flush_messages"`
	EnableTLS       bool          `json:"enable_tls" yaml:"enable_tls"`
	EnableSASL      bool          `json:"enable_sasl" yaml:"enable_sasl"`
	SASLMechanism   string        `json:"sasl_mechanism" yaml:"sasl_mechanism"`
	SASLUser        string        `json:"sasl_user" yaml:"sasl_user"`
	SASLPassword    string        `json:"sasl_password" yaml:"sasl_password"`
}

// Client Kafka 客户端
type Client struct {
	*mq.Base
	config        Config
	saramaConfig  *sarama.Config
	producer      sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
	consumerGroup sarama.ConsumerGroup

	mu          sync.RWMutex
	handlers    map[string]mq.Handler
	subscribers map[string]context.CancelFunc
}

// New 创建 Kafka 客户端
func New(cfg Config) (*Client, error) {
	name := cfg.Name
	if name == "" {
		name = "kafka"
	}

	saramaCfg := sarama.NewConfig()

	// 版本
	if cfg.Version != "" {
		version, err := sarama.ParseKafkaVersion(cfg.Version)
		if err != nil {
			return nil, fmt.Errorf("kafka: invalid version: %w", err)
		}
		saramaCfg.Version = version
	}

	// 生产者配置
	saramaCfg.Producer.RequiredAcks = sarama.RequiredAcks(cfg.RequiredAcks)
	saramaCfg.Producer.Retry.Max = cfg.MaxRetries
	if cfg.RetryBackoff > 0 {
		saramaCfg.Producer.Retry.Backoff = cfg.RetryBackoff
	}
	if cfg.FlushFrequency > 0 {
		saramaCfg.Producer.Flush.Frequency = cfg.FlushFrequency
	}
	if cfg.FlushMessages > 0 {
		saramaCfg.Producer.Flush.Messages = cfg.FlushMessages
	}
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Return.Errors = true

	// 消费者配置
	saramaCfg.Consumer.Return.Errors = true
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	// ClientID
	if cfg.ClientID != "" {
		saramaCfg.ClientID = cfg.ClientID
	}

	// SASL
	if cfg.EnableSASL {
		saramaCfg.Net.SASL.Enable = true
		saramaCfg.Net.SASL.User = cfg.SASLUser
		saramaCfg.Net.SASL.Password = cfg.SASLPassword
		switch cfg.SASLMechanism {
		case "PLAIN":
			saramaCfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		case "SCRAM-SHA-256":
			saramaCfg.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		case "SCRAM-SHA-512":
			saramaCfg.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		}
	}

	// TLS
	if cfg.EnableTLS {
		saramaCfg.Net.TLS.Enable = true
	}

	return &Client{
		Base:         mq.NewBase(name, mq.TypeKafka),
		config:       cfg,
		saramaConfig: saramaCfg,
		handlers:     make(map[string]mq.Handler),
		subscribers:  make(map[string]context.CancelFunc),
	}, nil
}

func (c *Client) Connect(ctx context.Context) error {
	if !c.CompareAndSwapState(mq.StateDisconnected, mq.StateConnecting) {
		return fmt.Errorf("kafka: invalid state for connect")
	}

	// 创建同步生产者
	producer, err := sarama.NewSyncProducer(c.config.Brokers, c.saramaConfig)
	if err != nil {
		c.SetState(mq.StateDisconnected)
		return fmt.Errorf("kafka: create producer failed: %w", err)
	}
	c.producer = producer

	// 创建异步生产者
	asyncProducer, err := sarama.NewAsyncProducer(c.config.Brokers, c.saramaConfig)
	if err != nil {
		_ = producer.Close()
		c.SetState(mq.StateDisconnected)
		return fmt.Errorf("kafka: create async producer failed: %w", err)
	}
	c.asyncProducer = asyncProducer

	c.SetState(mq.StateConnected)
	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	if c.State() != mq.StateConnected {
		return fmt.Errorf("kafka: not connected")
	}

	// 通过获取 broker 列表来检查连接
	client, err := sarama.NewClient(c.config.Brokers, c.saramaConfig)
	if err != nil {
		return fmt.Errorf("kafka: ping failed: %w", err)
	}
	defer client.Close()

	return nil
}

func (c *Client) Publish(ctx context.Context, topic string, value []byte, opts ...mq.PublishOption) (*mq.PublishResult, error) {
	if c.producer == nil {
		return nil, fmt.Errorf("kafka: producer not initialized")
	}

	var options mq.PublishOptions
	for _, opt := range opts {
		opt(&options)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}

	if options.Key != "" {
		msg.Key = sarama.StringEncoder(options.Key)
	}

	if options.Partition != nil {
		msg.Partition = *options.Partition
	}

	for k, v := range options.Headers {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	partition, offset, err := c.producer.SendMessage(msg)
	if err != nil {
		c.IncErrors()
		return nil, fmt.Errorf("kafka: publish failed: %w", err)
	}

	c.IncPublished()
	return &mq.PublishResult{
		MessageID: fmt.Sprintf("%d-%d", partition, offset),
		Partition: partition,
		Offset:    offset,
	}, nil
}

func (c *Client) PublishAsync(ctx context.Context, topic string, value []byte, callback func(*mq.PublishResult, error), opts ...mq.PublishOption) {
	if c.asyncProducer == nil {
		if callback != nil {
			callback(nil, fmt.Errorf("kafka: async producer not initialized"))
		}
		return
	}

	var options mq.PublishOptions
	for _, opt := range opts {
		opt(&options)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}

	if options.Key != "" {
		msg.Key = sarama.StringEncoder(options.Key)
	}

	for k, v := range options.Headers {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	c.asyncProducer.Input() <- msg

	// 异步处理结果
	if callback != nil {
		go func() {
			select {
			case success := <-c.asyncProducer.Successes():
				c.IncPublished()
				callback(&mq.PublishResult{
					MessageID: fmt.Sprintf("%d-%d", success.Partition, success.Offset),
					Partition: success.Partition,
					Offset:    success.Offset,
				}, nil)
			case err := <-c.asyncProducer.Errors():
				c.IncErrors()
				callback(nil, err.Err)
			}
		}()
	}
}

func (c *Client) Subscribe(ctx context.Context, topic string, handler mq.Handler, opts ...mq.SubscribeOption) error {
	options := mq.DefaultSubscribeOptions()
	for _, opt := range opts {
		opt(&options)
	}

	if options.Group == "" {
		options.Group = c.config.ClientID + "-group"
	}

	consumerGroup, err := sarama.NewConsumerGroup(c.config.Brokers, options.Group, c.saramaConfig)
	if err != nil {
		return fmt.Errorf("kafka: create consumer group failed: %w", err)
	}

	c.mu.Lock()
	c.handlers[topic] = handler
	c.consumerGroup = consumerGroup
	c.mu.Unlock()

	subCtx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	c.subscribers[topic] = cancel
	c.mu.Unlock()

	// 启动消费
	go func() {
		h := &consumerGroupHandler{
			client:  c,
			handler: handler,
			options: options,
		}

		for {
			select {
			case <-subCtx.Done():
				return
			default:
				if err := consumerGroup.Consume(subCtx, []string{topic}, h); err != nil {
					c.IncErrors()
				}
			}
		}
	}()

	return nil
}

func (c *Client) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cancel, ok := c.subscribers[topic]; ok {
		cancel()
		delete(c.subscribers, topic)
	}
	delete(c.handlers, topic)

	if c.consumerGroup != nil {
		return c.consumerGroup.Close()
	}
	return nil
}

func (c *Client) Close() error {
	c.SetState(mq.StateDisconnecting)

	c.mu.Lock()
	for _, cancel := range c.subscribers {
		cancel()
	}
	c.subscribers = make(map[string]context.CancelFunc)
	c.handlers = make(map[string]mq.Handler)
	c.mu.Unlock()

	var errs []error

	if c.producer != nil {
		if err := c.producer.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.asyncProducer != nil {
		if err := c.asyncProducer.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.consumerGroup != nil {
		if err := c.consumerGroup.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	c.SetState(mq.StateDisconnected)

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// consumerGroupHandler 实现 sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	client  *Client
	handler mq.Handler
	options mq.SubscribeOptions
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		mqMsg := &mq.Message{
			ID:        fmt.Sprintf("%d-%d", msg.Partition, msg.Offset),
			Topic:     msg.Topic,
			Key:       string(msg.Key),
			Value:     msg.Value,
			Headers:   make(map[string]string),
			Timestamp: msg.Timestamp,
			Raw:       msg,
		}

		for _, header := range msg.Headers {
			mqMsg.Headers[string(header.Key)] = string(header.Value)
		}

		var err error
		for attempt := 0; attempt <= h.options.MaxRetries; attempt++ {
			if attempt > 0 {
				h.client.IncRetries()
				time.Sleep(h.options.RetryDelay)
			}

			err = h.handler(session.Context(), mqMsg)
			if err == nil {
				break
			}
		}

		if err != nil {
			h.client.IncErrors()
		} else {
			h.client.IncConsumed()
			if h.options.AutoAck {
				session.MarkMessage(msg, "")
			}
		}
	}
	return nil
}

var _ mq.Client = (*Client)(nil)
