// Package eventbus 提供进程内事件总线。
//
// 核心功能：
//   - 发布/订阅模式
//   - 同步/异步事件分发
//
// 使用示例：
//
//	bus := eventbus.New()
//	bus.Subscribe("event", handler)
//	bus.Publish("event", data)
package eventbus
