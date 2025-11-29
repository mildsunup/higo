// Package lock 提供分布式锁。
//
// 核心功能：
//   - 基于 Redis/内存的分布式锁实现
//   - 自动续期、超时释放
//
// 使用示例：
//
//	lock := lock.NewRedis(client, "key", ttl)
//	if err := lock.Lock(ctx); err == nil {
//	    defer lock.Unlock(ctx)
//	    // 执行临界区代码
//	}
package lock
