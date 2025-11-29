// Package logger 提供结构化日志。
//
// 核心功能：
//   - 基于 Zap 的高性能日志
//   - 日志级别、格式、输出配置
//   - 链路追踪上下文注入
//
// 使用示例：
//
//	log := logger.New(cfg)
//	log.Info(ctx, "message", logger.String("key", "value"))
//	log.Error(ctx, "error", logger.Error(err))
package logger
