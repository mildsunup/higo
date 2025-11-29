// Package response 提供统一响应格式。
//
// 核心功能：
//   - 泛型响应结构（类型安全）
//   - 分页响应
//   - 框架无关（纯数据结构）
//
// 使用示例：
//
//	// 成功响应
//	resp := response.OK(user)
//	c.JSON(200, resp)
//
//	// 分页响应
//	resp := response.Page(users, 100, 1, 20)
//	c.JSON(200, resp)
//
//	// 带追踪信息
//	resp := response.OK(data).WithRequestID(id).WithTraceID(traceID)
//	c.JSON(200, resp)
package response
