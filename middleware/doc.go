// Package middleware 提供 HTTP/gRPC 中间件。
//
// HTTP 中间件：
//   - CORS、认证、限流、超时
//   - 请求 ID、日志、恢复、幂等性
//
// gRPC 拦截器：
//   - 链路追踪、指标采集
//   - 日志、恢复、认证
//
// 使用示例：
//
//	router.Use(middleware.RequestID())
//	router.Use(middleware.Logging(logger))
//	router.Use(middleware.Recovery())
package middleware
