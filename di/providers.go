package di

import (
	"time"

	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm/logger"

	"higo/aop"
	"higo/config"
	pkglogger "higo/logger"
	"higo/observability"
	"higo/storage"
	"higo/storage/mongodb"
	"higo/storage/mysql"
	"higo/storage/redis"
)

// ProvideLogger 提供应用日志组件
func ProvideLogger(cfg *config.Config) (pkglogger.Logger, error) {
	return pkglogger.New(pkglogger.Config{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		Output:     cfg.Logger.OutputPath,
		AddCaller:  true,
		CallerSkip: 2,
	})
}

// ProvideAOPChain provides AOP interceptor chain with default aspects
func ProvideAOPChain(log pkglogger.Logger) *aop.Chain {
	return aop.NewChain(
		aop.Recovery(func(r any) {
			log.Error(nil, "panic recovered", pkglogger.Any("panic", r))
		}),
	)
}

// ProvideObservability provides observability instance
func ProvideObservability(cfg *config.Config) (*observability.Observability, error) {
	return observability.New(observability.Config{
		ServiceName:    cfg.Observability.Tracing.ServiceName,
		ServiceVersion: cfg.App.Version,
		Environment:    cfg.App.Env,
		Tracing: observability.TracingConfig{
			Enabled:  cfg.Observability.Tracing.Enabled,
			Exporter: cfg.Observability.Tracing.Exporter,
			Endpoint: cfg.Observability.Tracing.Endpoint,
			Insecure: true,
			Sampler:  "ratio",
			Ratio:    cfg.Observability.Tracing.SampleRate,
		},
		Metrics: observability.MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
	})
}

// ProvideTracerProvider provides OpenTelemetry tracer provider
func ProvideTracerProvider(obs *observability.Observability) trace.TracerProvider {
	return obs.TracerProvider()
}

// ProvideMySQL provides MySQL storage instance
func ProvideMySQL(cfg *config.Config, tp trace.TracerProvider) *mysql.Storage {
	if !cfg.Storage.MySQL.Enabled {
		return nil
	}

	mysqlConfig := mysql.Config{
		Name:            "mysql",
		DSN:             cfg.Storage.MySQL.DSN,
		MaxOpenConns:    cfg.Storage.MySQL.MaxOpenConns,
		MaxIdleConns:    cfg.Storage.MySQL.MaxIdleConns,
		ConnMaxLifetime: cfg.Storage.MySQL.ConnMaxLifetime,
		LogLevel:        logger.LogLevel(cfg.Storage.MySQL.LogLevel),
		Replicas:        cfg.Storage.MySQL.Replicas,
	}

	if cfg.Storage.EnableTracing {
		mysqlConfig.Tracer = tp
	}

	return mysql.New(mysqlConfig)
}

// ProvideRedis provides Redis storage instance
func ProvideRedis(cfg *config.Config, tp trace.TracerProvider) *redis.Storage {
	if !cfg.Storage.Redis.Enabled {
		return nil
	}

	redisConfig := redis.Config{
		Name:         "redis",
		Addr:         cfg.Storage.Redis.Addr,
		Password:     cfg.Storage.Redis.Password,
		DB:           cfg.Storage.Redis.DB,
		MaxRetries:   cfg.Storage.Redis.MaxRetries,
		PoolSize:     cfg.Storage.Redis.PoolSize,
		MinIdleConns: cfg.Storage.Redis.MinIdleConns,
	}

	if cfg.Storage.EnableTracing {
		redisConfig.Tracer = tp
	}

	return redis.New(redisConfig)
}

// ProvideMongoDB provides MongoDB storage instance
func ProvideMongoDB(cfg *config.Config, tp trace.TracerProvider) *mongodb.Storage {
	if !cfg.Storage.MongoDB.Enabled {
		return nil
	}

	mongoConfig := mongodb.Config{
		Name:           "mongodb",
		URI:            cfg.Storage.MongoDB.URI,
		Database:       cfg.Storage.MongoDB.Database,
		MaxPoolSize:    cfg.Storage.MongoDB.MaxPoolSize,
		MinPoolSize:    cfg.Storage.MongoDB.MinPoolSize,
		ConnectTimeout: time.Duration(cfg.Storage.MongoDB.ConnectTimeout) * time.Second,
	}

	if cfg.Storage.EnableTracing {
		mongoConfig.Tracer = tp
	}

	return mongodb.New(mongoConfig)
}

// ProvideStorageMetrics provides storage metrics
func ProvideStorageMetrics(obs *observability.Observability) *storage.Metrics {
	return storage.NewMetrics(obs.Metrics())
}

// ProvideStorageManager provides the storage manager with registered storages
func ProvideStorageManager(
	mysqlStore *mysql.Storage,
	redisStore *redis.Storage,
	mongoStore *mongodb.Storage,
	cfg *config.Config,
	tp trace.TracerProvider,
	metrics *storage.Metrics,
	log pkglogger.Logger) (storage.Manager, error) {

	mgr := storage.NewManager()

	var tracer trace.Tracer
	if cfg.Storage.EnableTracing {
		tracer = tp.Tracer("higo.storage")
	}

	// Register MySQL
	if mysqlStore != nil {
		s := buildStorage(mysqlStore, tracer, metrics)
		if err := mgr.Register(s); err != nil {
			return nil, err
		}
		log.Info(nil, "Registered MySQL storage")
	}

	// Register Redis
	if redisStore != nil {
		s := buildStorage(redisStore, tracer, metrics)
		if err := mgr.Register(s); err != nil {
			return nil, err
		}
		log.Info(nil, "Registered Redis storage")
	}

	// Register MongoDB
	if mongoStore != nil {
		s := buildStorage(mongoStore, tracer, metrics)
		if err := mgr.Register(s); err != nil {
			return nil, err
		}
		log.Info(nil, "Registered MongoDB storage")
	}

	return mgr, nil
}

func buildStorage(s storage.Storage, tracer trace.Tracer, metrics *storage.Metrics) storage.Storage {
	b := storage.NewBuilder(s)
	if tracer != nil {
		b = b.WithTracing(tracer)
	}
	if metrics != nil {
		b = b.WithMetrics(metrics)
	}
	return b.Build()
}

// ProvideConfig 提供配置
func ProvideConfig(configPath string) (*config.Config, error) {
	return config.LoadByEnv(configPath)
}

// StorageSet is the Provider Set for storage
var StorageSet = wire.NewSet(
	ProvideMongoDB,
)

// ConfigSet 是配置相关的 Provider Set
var ConfigSet = wire.NewSet(
	ProvideConfig,
)

// InfraSet 是基础设施的 Provider Set
var InfraSet = wire.NewSet(
	ProvideLogger,
	ProvideObservability,
	ProvideTracerProvider,
	ProvideStorageMetrics,
	ProvideStorageManager,
	ProvideMySQL,
	ProvideRedis,
	ProvideMongoDB,
	ProvideAOPChain,
)
