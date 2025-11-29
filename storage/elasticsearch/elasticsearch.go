package elasticsearch

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"

	"github.com/mildsunup/higo/storage"
)

// Config Elasticsearch 配置
type Config struct {
	Name       string   `json:"name" yaml:"name"`
	Addresses  []string `json:"addresses" yaml:"addresses"`
	Username   string   `json:"username" yaml:"username"`
	Password   string   `json:"password" yaml:"password"`
	MaxRetries int      `json:"max_retries" yaml:"max_retries"`
}

// Storage Elasticsearch 存储
type Storage struct {
	*storage.Base
	client *elasticsearch.Client
	config Config
}

// New 创建 Elasticsearch 存储
func New(cfg Config) *Storage {
	name := cfg.Name
	if name == "" {
		name = "elasticsearch"
	}
	return &Storage{
		Base:   storage.NewBase(name, storage.TypeElasticsearch),
		config: cfg,
	}
}

func (s *Storage) Connect(ctx context.Context) error {
	if !s.CompareAndSwapState(storage.StateDisconnected, storage.StateConnecting) {
		return fmt.Errorf("elasticsearch: invalid state for connect")
	}

	cfg := elasticsearch.Config{
		Addresses:  s.config.Addresses,
		Username:   s.config.Username,
		Password:   s.config.Password,
		MaxRetries: s.config.MaxRetries,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		s.SetState(storage.StateDisconnected)
		return fmt.Errorf("elasticsearch: create client failed: %w", err)
	}

	res, err := client.Ping(client.Ping.WithContext(ctx))
	if err != nil {
		s.SetState(storage.StateDisconnected)
		return fmt.Errorf("elasticsearch: ping failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		s.SetState(storage.StateDisconnected)
		return fmt.Errorf("elasticsearch: ping error: %s", res.Status())
	}

	s.client = client
	s.SetState(storage.StateConnected)
	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	if s.client == nil {
		return fmt.Errorf("elasticsearch: not connected")
	}

	res, err := s.client.Ping(s.client.Ping.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch: ping error: %s", res.Status())
	}
	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	s.SetState(storage.StateDisconnecting)
	s.client = nil
	s.SetState(storage.StateDisconnected)
	return nil
}

// Client 返回 Elasticsearch 客户端
func (s *Storage) Client() *elasticsearch.Client { return s.client }

var _ storage.Storage = (*Storage)(nil)
