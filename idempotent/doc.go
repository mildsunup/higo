// Package idempotent 提供幂等性保证。
//
// 核心功能：
//   - 基于 Redis 的幂等性令牌
//   - 防止重复提交
//
// 使用示例：
//
//	idem := idempotent.New(redis)
//	if err := idem.Check(ctx, token); err == nil {
//	    // 执行业务逻辑
//	    idem.Consume(ctx, token)
//	}
package idempotent
