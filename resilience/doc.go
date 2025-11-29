// Package resilience 提供弹性模式。
//
// 核心功能：
//   - 熔断器（Circuit Breaker）
//   - 重试策略（Retry）
//   - 限流器（Rate Limiter）
//
// 使用示例：
//
//	cb := resilience.NewCircuitBreaker(threshold, timeout)
//	err := cb.Execute(func() error {
//	    return doSomething()
//	})
package resilience
