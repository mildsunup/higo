// Package cache 提供缓存抽象层。
//
// 核心功能：
//   - 多级缓存（内存 + Redis）
//   - 缓存穿透/击穿/雪崩防护
//   - 序列化策略（JSON/MessagePack）
//   - 统计信息（命中率、键数量）
//
// 使用示例：
//
//	c := cache.NewRedis(client, cache.WithPrefix("app:"))
//	err := c.Set(ctx, "key", value, time.Hour)
//	err = c.Get(ctx, "key", &dest)
package cache
