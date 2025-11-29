// Package aop 提供面向切面编程（AOP）支持。
//
// 核心功能：
//   - 拦截器链（Interceptor Chain）
//   - 方法调用拦截（Invocation）
//   - 泛型处理器包装
//
// 使用示例：
//
//	chain := aop.NewChain(loggingInterceptor, tracingInterceptor)
//	handler := aop.Wrap(chain, "method", originalHandler)
package aop
