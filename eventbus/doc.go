// Package eventbus 提供领域事件发布能力。
//
// 核心功能：
//   - 基于消息队列的事件发布
//   - 同步/异步发布模式
//   - 事件信封封装
//
// 使用示例：
//
//	// 创建事件总线
//	bus := eventbus.New(mqProducer)
//
//	// 自定义配置
//	bus := eventbus.New(mqProducer,
//	    eventbus.WithTopic("my-events"),
//	    eventbus.WithIDFunc(customIDGen),
//	)
//
//	// 发布事件
//	err := bus.Publish(ctx, myEvent)
//	bus.PublishAsync(ctx, myEvent)
package eventbus
