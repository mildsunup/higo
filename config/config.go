package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config 应用程序主配置结构
type Config struct {
	App           AppConfig           `yaml:"app" mapstructure:"app"`
	Auth          AuthConfig          `yaml:"auth" mapstructure:"auth"`
	Server        ServerConfig        `yaml:"server" mapstructure:"server"`
	Redis         RedisConfig         `yaml:"redis" mapstructure:"redis"` // 顶层 Redis（缓存/EventBus）
	Logger        LoggerConfig        `yaml:"logger" mapstructure:"logger"`
	Observability ObservabilityConfig `yaml:"observability" mapstructure:"observability"`
	Storage       StorageConfig       `yaml:"storage" mapstructure:"storage"`
	MQ            MQConfig            `yaml:"mq" mapstructure:"mq"`
}

// AppConfig 应用程序基础配置
type AppConfig struct {
	Name        string `yaml:"name" mapstructure:"name"`
	Env         string `yaml:"env" mapstructure:"env"`
	Version     string `yaml:"version" mapstructure:"version"`
	Description string `yaml:"description" mapstructure:"description"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret string        `yaml:"jwt_secret" mapstructure:"jwt_secret"`
	JWTExpiry time.Duration `yaml:"jwt_expiry" mapstructure:"jwt_expiry"`
}

// RedisConfig 顶层 Redis 配置（用于缓存和 EventBus）
type RedisConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	Addr     string `yaml:"addr" mapstructure:"addr"`
	Password string `yaml:"password" mapstructure:"password"`
	DB       int    `yaml:"db" mapstructure:"db"`
	PoolSize int    `yaml:"pool_size" mapstructure:"pool_size"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	HTTP HTTPConfig `yaml:"http" mapstructure:"http"`
	GRPC GRPCConfig `yaml:"grpc" mapstructure:"grpc"`
}

// HTTPConfig HTTP 服务器配置
type HTTPConfig struct {
	Enabled      bool          `yaml:"enabled" mapstructure:"enabled"`
	Port         string        `yaml:"port" mapstructure:"port"`
	Mode         string        `yaml:"mode" mapstructure:"mode"`
	ReadTimeout  time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	EnableH2C    bool          `yaml:"enable_h2c" mapstructure:"enable_h2c"`
	ServiceName  string        `yaml:"service_name" mapstructure:"service_name"`
}

// GRPCConfig gRPC 服务器配置
type GRPCConfig struct {
	Enabled           bool          `yaml:"enabled" mapstructure:"enabled"`
	Port              string        `yaml:"port" mapstructure:"port"`
	MaxRecvMsgSize    int           `yaml:"max_recv_msg_size" mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize    int           `yaml:"max_send_msg_size" mapstructure:"max_send_msg_size"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout" mapstructure:"connection_timeout"`
	ServiceName       string        `yaml:"service_name" mapstructure:"service_name"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `yaml:"level" mapstructure:"level"`
	Format     string `yaml:"format" mapstructure:"format"`
	OutputPath string `yaml:"output_path" mapstructure:"output_path"`
}

// ObservabilityConfig 可观测性配置
type ObservabilityConfig struct {
	Metrics MetricsConfig `yaml:"metrics" mapstructure:"metrics"`
	Tracing TracingConfig `yaml:"tracing" mapstructure:"tracing"`
}

// MetricsConfig 监控配置
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Path    string `yaml:"path" mapstructure:"path"`
	Port    string `yaml:"port" mapstructure:"port"`
}

// TracingConfig 追踪配置
type TracingConfig struct {
	Enabled     bool    `yaml:"enabled" mapstructure:"enabled"`
	ServiceName string  `yaml:"service_name" mapstructure:"service_name"`
	Exporter    string  `yaml:"exporter" mapstructure:"exporter"`
	Endpoint    string  `yaml:"endpoint" mapstructure:"endpoint"`
	SampleRate  float64 `yaml:"sample_rate" mapstructure:"sample_rate"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	EnableTracing bool                       `yaml:"enable_tracing" mapstructure:"enable_tracing"`
	MySQL         MySQLStorageConfig         `yaml:"mysql" mapstructure:"mysql"`
	Redis         RedisStorageConfig         `yaml:"redis" mapstructure:"redis"`
	MongoDB       MongoDBStorageConfig       `yaml:"mongodb" mapstructure:"mongodb"`
	ClickHouse    ClickHouseStorageConfig    `yaml:"clickhouse" mapstructure:"clickhouse"`
	Elasticsearch ElasticsearchStorageConfig `yaml:"elasticsearch" mapstructure:"elasticsearch"`
}

// MySQLStorageConfig MySQL 配置
type MySQLStorageConfig struct {
	Enabled         bool          `yaml:"enabled" mapstructure:"enabled"`
	DSN             string        `yaml:"dsn" mapstructure:"dsn"`
	Replicas        []string      `yaml:"replicas" mapstructure:"replicas"`
	MaxOpenConns    int           `yaml:"max_open_conns" mapstructure:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
	LogLevel        int           `yaml:"log_level" mapstructure:"log_level"`
}

// RedisStorageConfig Redis 配置
type RedisStorageConfig struct {
	Enabled      bool   `yaml:"enabled" mapstructure:"enabled"`
	Addr         string `yaml:"addr" mapstructure:"addr"`
	Password     string `yaml:"password" mapstructure:"password"`
	DB           int    `yaml:"db" mapstructure:"db"`
	MaxRetries   int    `yaml:"max_retries" mapstructure:"max_retries"`
	PoolSize     int    `yaml:"pool_size" mapstructure:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns" mapstructure:"min_idle_conns"`
}

// MongoDBStorageConfig MongoDB 配置
type MongoDBStorageConfig struct {
	Enabled        bool          `yaml:"enabled" mapstructure:"enabled"`
	URI            string        `yaml:"uri" mapstructure:"uri"`
	Database       string        `yaml:"database" mapstructure:"database"`
	MaxPoolSize    uint64        `yaml:"max_pool_size" mapstructure:"max_pool_size"`
	MinPoolSize    uint64        `yaml:"min_pool_size" mapstructure:"min_pool_size"`
	ConnectTimeout time.Duration `yaml:"connect_timeout" mapstructure:"connect_timeout"`
}

// ClickHouseStorageConfig ClickHouse 配置
type ClickHouseStorageConfig struct {
	Enabled         bool          `yaml:"enabled" mapstructure:"enabled"`
	DSN             string        `yaml:"dsn" mapstructure:"dsn"`
	MaxOpenConns    int           `yaml:"max_open_conns" mapstructure:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
}

// ElasticsearchStorageConfig Elasticsearch 配置
type ElasticsearchStorageConfig struct {
	Enabled   bool     `yaml:"enabled" mapstructure:"enabled"`
	Addresses []string `yaml:"addresses" mapstructure:"addresses"`
	Username  string   `yaml:"username" mapstructure:"username"`
	Password  string   `yaml:"password" mapstructure:"password"`
}

// MQConfig 消息队列配置
type MQConfig struct {
	Kafka    KafkaConfig    `yaml:"kafka" mapstructure:"kafka"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq" mapstructure:"rabbitmq"`
}

// KafkaConfig Kafka 配置
type KafkaConfig struct {
	Enabled  bool     `yaml:"enabled" mapstructure:"enabled"`
	Brokers  []string `yaml:"brokers" mapstructure:"brokers"`
	ClientID string   `yaml:"client_id" mapstructure:"client_id"`
}

// RabbitMQConfig RabbitMQ 配置
type RabbitMQConfig struct {
	Enabled      bool   `yaml:"enabled" mapstructure:"enabled"`
	URL          string `yaml:"url" mapstructure:"url"`
	ExchangeName string `yaml:"exchange_name" mapstructure:"exchange_name"`
	ExchangeType string `yaml:"exchange_type" mapstructure:"exchange_type"`
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	if !c.Server.HTTP.Enabled && !c.Server.GRPC.Enabled {
		return fmt.Errorf("at least one server (HTTP or gRPC) must be enabled")
	}

	if c.Server.HTTP.Enabled && c.Server.HTTP.Port == "" {
		return fmt.Errorf("http.port is required when HTTP is enabled")
	}

	if c.Server.GRPC.Enabled && c.Server.GRPC.Port == "" {
		return fmt.Errorf("grpc.port is required when gRPC is enabled")
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if c.Logger.Level != "" && !validLevels[c.Logger.Level] {
		return fmt.Errorf("logger.level must be debug, info, warn, or error")
	}

	return nil
}

// GetEnv 获取当前环境
func GetEnv() string {
	if env := os.Getenv("APP_ENV"); env != "" {
		return env
	}
	if env := os.Getenv("GO_ENV"); env != "" {
		return env
	}
	return "development"
}

// GetConfigPath 根据环境获取配置路径
func GetConfigPath(env string) string {
	switch strings.ToLower(env) {
	case "dev", "development":
		return "configs/config.dev.yaml"
	case "staging", "stage":
		return "configs/config.staging.yaml"
	case "prod", "production":
		return "configs/config.prod.yaml"
	default:
		return "configs/config.dev.yaml"
	}
}

// Load 从文件加载配置
func Load(configPath string) (*Config, error) {
	return LoadWithOptions(context.Background(), configPath)
}

// LoadByEnv 根据环境加载配置
func LoadByEnv(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = GetConfigPath(GetEnv())
	}
	return Load(configPath)
}

// LoadWithOptions 使用选项加载配置
func LoadWithOptions(ctx context.Context, configPath string, opts ...Option) (*Config, error) {
	// 默认提供者：文件 + 环境变量
	providers := []Provider{
		NewFileProvider(configPath),
		NewEnvProvider("HIGO"),
	}

	// 检查是否使用远程配置
	if source := os.Getenv("CONFIG_SOURCE"); source != "" {
		switch strings.ToLower(source) {
		case "consul":
			endpoint := os.Getenv("CONFIG_REMOTE_ADDR")
			path := os.Getenv("CONFIG_REMOTE_PATH")
			if endpoint != "" && path != "" {
				providers = append([]Provider{NewRemoteProvider(RemoteConsul, endpoint, path)}, providers...)
			}
		case "etcd", "etcd3":
			endpoint := os.Getenv("CONFIG_REMOTE_ADDR")
			path := os.Getenv("CONFIG_REMOTE_PATH")
			if endpoint != "" && path != "" {
				providers = append([]Provider{NewRemoteProvider(RemoteEtcd3, endpoint, path)}, providers...)
			}
		}
	}

	// 合并选项
	allOpts := make([]Option, 0, len(providers)+len(opts))
	for _, p := range providers {
		allOpts = append(allOpts, WithProvider(p))
	}
	allOpts = append(allOpts, opts...)

	// 添加环境变量解析器
	allOpts = append(allOpts, WithSecretResolver(NewEnvSecretResolver("")))

	loader := NewLoader(allOpts...)

	var cfg Config
	if err := loader.Load(ctx, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Default 返回默认配置
func Default() *Config {
	return &Config{
		App: AppConfig{
			Name:    "higo",
			Env:     "development",
			Version: "1.0.0",
		},
		Server: ServerConfig{
			HTTP: HTTPConfig{
				Enabled:      true,
				Port:         "8080",
				Mode:         "debug",
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
				EnableH2C:    true,
				ServiceName:  "higo-http",
			},
			GRPC: GRPCConfig{
				Enabled:           false,
				Port:              "9091",
				MaxRecvMsgSize:    4 << 20,
				MaxSendMsgSize:    4 << 20,
				ConnectionTimeout: 10 * time.Second,
				ServiceName:       "higo-grpc",
			},
		},
		Logger: LoggerConfig{
			Level:      "info",
			Format:     "json",
			OutputPath: "stdout",
		},
		Observability: ObservabilityConfig{
			Metrics: MetricsConfig{
				Enabled: true,
				Path:    "/metrics",
			},
			Tracing: TracingConfig{
				Enabled:    false,
				SampleRate: 0.1,
			},
		},
	}
}
