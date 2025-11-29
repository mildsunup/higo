// Package mq 提供消息队列抽象层。
//
// 支持的消息队列：
//   - Kafka
//   - RabbitMQ
//   - Memory（内存队列，用于测试）
//
// 核心功能：
//   - 统一的生产者/消费者接口
//   - 消息发布/订阅、异步处理
//   - 链路追踪和指标采集
//
// 使用示例：
//
//	client := kafka.New("events", cfg)
//	client.Publish(ctx, "topic", data)
//	client.Subscribe(ctx, "topic", handler)
package mq
