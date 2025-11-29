// Package observability 提供可观测性支持。
//
// 核心功能：
//   - 链路追踪（OpenTelemetry）
//   - 指标采集（Prometheus）
//   - 自动 Span 注入
//
// 使用示例：
//
//	obs, err := observability.New(cfg)
//	defer obs.Shutdown(ctx)
//	ctx, span := obs.Tracer().Start(ctx, "operation")
//	defer span.End()
package observability
