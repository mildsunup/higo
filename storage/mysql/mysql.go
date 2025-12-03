package mysql

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"gorm.io/plugin/opentelemetry/tracing"

	"github.com/mildsunup/higo/storage"
)

// Config MySQL 配置
type Config struct {
	Name            string              `json:"name" yaml:"name"`
	DSN             string              `json:"dsn" yaml:"dsn"`
	Replicas        []string            `json:"replicas" yaml:"replicas"`
	MaxOpenConns    int                 `json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int                 `json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration       `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration       `json:"conn_max_idle_time" yaml:"conn_max_idle_time"`
	LogLevel        gormlogger.LogLevel `json:"log_level" yaml:"log_level"`           // 1=Silent, 2=Error, 3=Warn, 4=Info
	SlowThreshold   time.Duration       `json:"slow_threshold" yaml:"slow_threshold"` // 慢查询阈值，默认 200ms
}

// Logger 日志接口
type Logger interface {
	Info(ctx context.Context, msg string, fields ...any)
	Warn(ctx context.Context, msg string, fields ...any)
	Error(ctx context.Context, msg string, fields ...any)
}

// Storage MySQL 存储
type Storage struct {
	*storage.Base
	db     *gorm.DB
	config Config
	logger Logger
	tracer trace.TracerProvider
}

// Option MySQL 存储选项
type Option func(*Storage)

// WithLogger 设置日志
func WithLogger(logger Logger) Option {
	return func(s *Storage) {
		s.logger = logger
	}
}

// WithTracer 设置追踪
func WithTracer(tracer trace.TracerProvider) Option {
	return func(s *Storage) {
		s.tracer = tracer
	}
}

// WithReplicas 设置只读副本
func WithReplicas(replicas ...string) Option {
	return func(s *Storage) {
		s.config.Replicas = replicas
	}
}

// New 创建 MySQL 存储
func New(cfg Config, opts ...Option) *Storage {
	name := cfg.Name
	if name == "" {
		name = "mysql"
	}
	if cfg.SlowThreshold <= 0 {
		cfg.SlowThreshold = 200 * time.Millisecond
	}

	s := &Storage{
		Base:   storage.NewBase(name, storage.TypeMySQL),
		config: cfg,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Storage) Connect(ctx context.Context) error {
	if !s.CompareAndSwapState(storage.StateDisconnected, storage.StateConnecting) {
		return fmt.Errorf("mysql: invalid state for connect")
	}

	// 创建 GORM logger
	var gormLog gormlogger.Interface
	if s.logger != nil {
		gormLog = newSlowQueryLogger(s.logger, s.config.SlowThreshold, s.config.LogLevel)
	} else {
		gormLog = gormlogger.Default.LogMode(s.config.LogLevel)
	}

	gormCfg := &gorm.Config{
		Logger: gormLog,
	}

	db, err := gorm.Open(mysql.Open(s.config.DSN), gormCfg)
	if err != nil {
		s.SetState(storage.StateDisconnected)
		return fmt.Errorf("mysql: connect failed: %w", err)
	}

	// OpenTelemetry 追踪
	if s.tracer != nil {
		if err := db.Use(tracing.NewPlugin(
			tracing.WithTracerProvider(s.tracer),
		)); err != nil {
			s.SetState(storage.StateDisconnected)
			return fmt.Errorf("mysql: setup tracing failed: %w", err)
		}
	}

	// 读写分离
	if len(s.config.Replicas) > 0 {
		replicas := make([]gorm.Dialector, len(s.config.Replicas))
		for i, dsn := range s.config.Replicas {
			replicas[i] = mysql.Open(dsn)
		}
		if err := db.Use(dbresolver.Register(dbresolver.Config{
			Sources:  []gorm.Dialector{mysql.Open(s.config.DSN)},
			Replicas: replicas,
			Policy:   dbresolver.RandomPolicy{},
		})); err != nil {
			s.SetState(storage.StateDisconnected)
			return fmt.Errorf("mysql: setup dbresolver failed: %w", err)
		}
	}

	// 连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		s.SetState(storage.StateDisconnected)
		return fmt.Errorf("mysql: get sql.DB failed: %w", err)
	}

	if s.config.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(s.config.MaxOpenConns)
	}
	if s.config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(s.config.MaxIdleConns)
	}
	if s.config.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(s.config.ConnMaxLifetime)
	}
	if s.config.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(s.config.ConnMaxIdleTime)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		s.SetState(storage.StateDisconnected)
		return fmt.Errorf("mysql: ping failed: %w", err)
	}

	s.db = db
	s.SetState(storage.StateConnected)
	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("mysql: not connected")
	}
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (s *Storage) Close(ctx context.Context) error {
	if s.db == nil {
		return nil
	}
	s.SetState(storage.StateDisconnecting)
	sqlDB, err := s.db.DB()
	if err != nil {
		s.SetState(storage.StateDisconnected)
		return err
	}
	err = sqlDB.Close()
	s.SetState(storage.StateDisconnected)
	s.db = nil
	return err
}

// DB 返回 GORM 实例
func (s *Storage) DB() *gorm.DB { return s.db }

// Stats 返回连接池统计
func (s *Storage) Stats() storage.Stats {
	if s.db == nil {
		return storage.Stats{}
	}
	sqlDB, err := s.db.DB()
	if err != nil {
		return storage.Stats{}
	}
	stats := sqlDB.Stats()
	return storage.Stats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
	}
}

var (
	_ storage.Storage       = (*Storage)(nil)
	_ storage.StatsProvider = (*Storage)(nil)
)

// --- 慢查询日志 ---

type slowQueryLogger struct {
	logger        Logger
	slowThreshold time.Duration
	logLevel      gormlogger.LogLevel
}

func newSlowQueryLogger(l Logger, threshold time.Duration, level gormlogger.LogLevel) *slowQueryLogger {
	return &slowQueryLogger{
		logger:        l,
		slowThreshold: threshold,
		logLevel:      level,
	}
}

func (l *slowQueryLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &slowQueryLogger{
		logger:        l.logger,
		slowThreshold: l.slowThreshold,
		logLevel:      level,
	}
}

func (l *slowQueryLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.logLevel >= gormlogger.Info {
		l.logger.Info(ctx, msg, data...)
	}
}

func (l *slowQueryLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.logLevel >= gormlogger.Warn {
		l.logger.Warn(ctx, msg, data...)
	}
}

func (l *slowQueryLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.logLevel >= gormlogger.Error {
		l.logger.Error(ctx, msg, data...)
	}
}

func (l *slowQueryLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 错误日志
	if err != nil && l.logLevel >= gormlogger.Error {
		l.logger.Error(ctx, "sql error",
			"error", err.Error(),
			"elapsed", elapsed.String(),
			"rows", rows,
			"sql", sql,
		)
		return
	}

	// 慢查询日志
	if elapsed > l.slowThreshold && l.slowThreshold > 0 && l.logLevel >= gormlogger.Warn {
		l.logger.Warn(ctx, "slow sql",
			"elapsed", elapsed.String(),
			"threshold", l.slowThreshold.String(),
			"rows", rows,
			"sql", sql,
		)
		return
	}

	// 普通日志
	if l.logLevel >= gormlogger.Info {
		l.logger.Info(ctx, "sql",
			"elapsed", elapsed.String(),
			"rows", rows,
			"sql", sql,
		)
	}
}

var _ gormlogger.Interface = (*slowQueryLogger)(nil)
