// Package storage 提供数据存储抽象层。
//
// 支持的存储类型：
//   - MySQL/PostgreSQL
//   - Redis
//   - MongoDB
//   - ClickHouse
//   - Elasticsearch
//
// 核心功能：
//   - 统一的存储接口
//   - 连接池管理、健康检查、重连机制
//   - 链路追踪和指标采集
//
// 使用示例：
//
//	mgr := storage.NewManager()
//	db := mysql.New("main", cfg)
//	mgr.Register(db)
//	mgr.ConnectAll(ctx)
package storage
