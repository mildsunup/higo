package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 全局 context 提取器
var contextExtractors []ContextExtractor

// RegisterExtractor 注册 context 提取器
func RegisterExtractor(e ContextExtractor) {
	contextExtractors = append(contextExtractors, e)
}

// zapLogger zap 实现
type zapLogger struct {
	base       *zap.Logger
	level      zap.AtomicLevel
	extractors []ContextExtractor
}

// Option 日志选项
type Option func(*zapLogger)

// WithExtractors 设置 context 提取器
func WithExtractors(extractors ...ContextExtractor) Option {
	return func(l *zapLogger) {
		l.extractors = append(l.extractors, extractors...)
	}
}

// New 创建 Logger
func New(cfg Config, opts ...Option) (Logger, error) {
	level := zap.NewAtomicLevelAt(toZapLevel(ParseLevel(cfg.Level)))

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.EncodeLevel = zapcore.LowercaseLevelEncoder

	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encCfg)
	}

	writer := buildWriter(cfg)
	core := zapcore.NewCore(encoder, writer, level)

	zapOpts := []zap.Option{}
	if cfg.AddCaller {
		zapOpts = append(zapOpts, zap.AddCaller(), zap.AddCallerSkip(cfg.CallerSkip))
	}

	l := &zapLogger{
		base:       zap.New(core, zapOpts...),
		level:      level,
		extractors: contextExtractors,
	}

	for _, opt := range opts {
		opt(l)
	}

	return l, nil
}

func (l *zapLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, zapcore.DebugLevel, msg, fields)
}

func (l *zapLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, zapcore.InfoLevel, msg, fields)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, zapcore.WarnLevel, msg, fields)
}

func (l *zapLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, zapcore.ErrorLevel, msg, fields)
}

func (l *zapLogger) log(ctx context.Context, level zapcore.Level, msg string, fields []Field) {
	if l.base == nil {
		return
	}

	// 提取 context 字段
	allFields := l.extractContext(ctx)
	allFields = append(allFields, toZapFields(fields)...)

	switch level {
	case zapcore.DebugLevel:
		l.base.Debug(msg, allFields...)
	case zapcore.InfoLevel:
		l.base.Info(msg, allFields...)
	case zapcore.WarnLevel:
		l.base.Warn(msg, allFields...)
	case zapcore.ErrorLevel:
		l.base.Error(msg, allFields...)
	}
}

func (l *zapLogger) extractContext(ctx context.Context) []zap.Field {
	if ctx == nil {
		return nil
	}
	var fields []zap.Field
	for _, e := range l.extractors {
		for _, f := range e(ctx) {
			fields = append(fields, toZapField(f))
		}
	}
	return fields
}

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{
		base:       l.base.With(toZapFields(fields)...),
		level:      l.level,
		extractors: l.extractors,
	}
}

func (l *zapLogger) WithLevel(level Level) Logger {
	newLevel := zap.NewAtomicLevelAt(toZapLevel(level))
	return &zapLogger{
		base:       l.base,
		level:      newLevel,
		extractors: l.extractors,
	}
}

func (l *zapLogger) Sync() error {
	if l.base == nil {
		return nil
	}
	return l.base.Sync()
}

// Zap 返回底层 zap.Logger（用于需要原生 zap 的场景）
func (l *zapLogger) Zap() *zap.Logger {
	return l.base
}

// --- 辅助函数 ---

func buildWriter(cfg Config) zapcore.WriteSyncer {
	switch cfg.Output {
	case "", "stdout":
		return zapcore.AddSync(os.Stdout)
	case "stderr":
		return zapcore.AddSync(os.Stderr)
	default:
		// 文件输出，使用 lumberjack 自动分割
		rot := cfg.Rotation
		maxSize := rot.MaxSize
		if maxSize <= 0 {
			maxSize = 100
		}
		maxBackups := rot.MaxBackups
		if maxBackups <= 0 {
			maxBackups = 3
		}
		maxAge := rot.MaxAge
		if maxAge <= 0 {
			maxAge = 7
		}
		return zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Output,
			MaxSize:    maxSize,    // MB
			MaxBackups: maxBackups, // 保留文件数
			MaxAge:     maxAge,     // 保留天数
			Compress:   rot.Compress,
		})
	}
}

func toZapLevel(l Level) zapcore.Level {
	switch l {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func toZapField(f Field) zap.Field {
	switch v := f.Value.(type) {
	case string:
		return zap.String(f.Key, v)
	case int:
		return zap.Int(f.Key, v)
	case int64:
		return zap.Int64(f.Key, v)
	case float64:
		return zap.Float64(f.Key, v)
	case bool:
		return zap.Bool(f.Key, v)
	case error:
		return zap.Error(v)
	default:
		return zap.Any(f.Key, v)
	}
}

func toZapFields(fields []Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	result := make([]zap.Field, len(fields))
	for i, f := range fields {
		result[i] = toZapField(f)
	}
	return result
}

// Nop 返回空实现
func Nop() Logger { return &nopLogger{} }

type nopLogger struct{}

func (l *nopLogger) Debug(ctx context.Context, msg string, fields ...Field) {}
func (l *nopLogger) Info(ctx context.Context, msg string, fields ...Field)  {}
func (l *nopLogger) Warn(ctx context.Context, msg string, fields ...Field)  {}
func (l *nopLogger) Error(ctx context.Context, msg string, fields ...Field) {}
func (l *nopLogger) With(fields ...Field) Logger                            { return l }
func (l *nopLogger) WithLevel(level Level) Logger                           { return l }
func (l *nopLogger) Sync() error                                            { return nil }
