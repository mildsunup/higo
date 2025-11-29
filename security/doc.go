// Package security 提供安全工具。
//
// 核心功能：
//   - 限流器（令牌桶/漏桶）
//   - 密码强度验证
//
// 使用示例：
//
//	limiter := security.NewTokenBucket(rate, burst)
//	if limiter.Allow() {
//	    // 处理请求
//	}
package security
