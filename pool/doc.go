// Package pool 提供对象池。
//
// 核心功能：
//   - Worker Pool（协程池）
//   - Buffer Pool（字节缓冲池）
//   - Object Pool（通用对象池）
//
// 使用示例：
//
//	wp := pool.NewWorkerPool(size)
//	wp.Submit(func() { /* task */ })
//	defer wp.Stop()
package pool
