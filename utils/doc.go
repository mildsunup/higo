// Package utils 提供通用工具函数。
//
// 核心功能：
//   - 字符串、切片、Map 操作
//   - 时间、数字处理
//   - 指针辅助、异步工具
//   - ID 生成（UUID/Snowflake）
//
// 使用示例：
//
//	id := utils.UUID()
//	ptr := utils.Ptr("value")
//	result := utils.Map(slice, func(v int) int { return v * 2 })
package utils
