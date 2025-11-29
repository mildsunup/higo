// Package server 提供 HTTP/gRPC 服务器抽象。
//
// 核心功能：
//   - HTTP 服务器（基于 Gin）
//   - gRPC 服务器
//   - 多路复用（同端口同时支持 HTTP/gRPC）
//   - 服务器组管理
//
// 使用示例：
//
//	http := server.NewHTTP(cfg)
//	grpc := server.NewGRPC(cfg)
//	group := server.NewGroup(http, grpc)
//	group.Start(ctx)
package server
