package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/otel/trace"

	"higo/storage"
)

// Config MongoDB 配置
type Config struct {
	Name           string        `json:"name" yaml:"name"`
	URI            string        `json:"uri" yaml:"uri"`
	Database       string        `json:"database" yaml:"database"`
	MaxPoolSize    uint64        `json:"max_pool_size" yaml:"max_pool_size"`
	MinPoolSize    uint64        `json:"min_pool_size" yaml:"min_pool_size"`
	ConnectTimeout time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
	Tracer         trace.TracerProvider
}

// Storage MongoDB 存储
type Storage struct {
	*storage.Base
	client   *mongo.Client
	database *mongo.Database
	config   Config
}

// New 创建 MongoDB 存储
func New(cfg Config) *Storage {
	name := cfg.Name
	if name == "" {
		name = "mongodb"
	}
	return &Storage{
		Base:   storage.NewBase(name, storage.TypeMongoDB),
		config: cfg,
	}
}

func (s *Storage) Connect(ctx context.Context) error {
	if !s.CompareAndSwapState(storage.StateDisconnected, storage.StateConnecting) {
		return fmt.Errorf("mongodb: invalid state for connect")
	}

	connectCtx := ctx
	if s.config.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		connectCtx, cancel = context.WithTimeout(ctx, s.config.ConnectTimeout)
		defer cancel()
	}

	opts := options.Client().ApplyURI(s.config.URI)
	if s.config.MaxPoolSize > 0 {
		opts.SetMaxPoolSize(s.config.MaxPoolSize)
	}
	if s.config.MinPoolSize > 0 {
		opts.SetMinPoolSize(s.config.MinPoolSize)
	}

	// OpenTelemetry 追踪
	if s.config.Tracer != nil {
		opts.SetMonitor(otelmongo.NewMonitor(
			otelmongo.WithTracerProvider(s.config.Tracer),
		))
	}

	client, err := mongo.Connect(connectCtx, opts)
	if err != nil {
		s.SetState(storage.StateDisconnected)
		return fmt.Errorf("mongodb: connect failed: %w", err)
	}

	if err := client.Ping(connectCtx, nil); err != nil {
		s.SetState(storage.StateDisconnected)
		return fmt.Errorf("mongodb: ping failed: %w", err)
	}

	s.client = client
	s.database = client.Database(s.config.Database)
	s.SetState(storage.StateConnected)
	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	if s.client == nil {
		return fmt.Errorf("mongodb: not connected")
	}
	return s.client.Ping(ctx, nil)
}

func (s *Storage) Close(ctx context.Context) error {
	if s.client == nil {
		return nil
	}
	s.SetState(storage.StateDisconnecting)
	err := s.client.Disconnect(ctx)
	s.SetState(storage.StateDisconnected)
	s.client = nil
	s.database = nil
	return err
}

// Client 返回 MongoDB 客户端
func (s *Storage) Client() *mongo.Client { return s.client }

// Database 返回数据库实例
func (s *Storage) Database() *mongo.Database { return s.database }

var _ storage.Storage = (*Storage)(nil)
