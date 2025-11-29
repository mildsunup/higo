// Package errors 提供错误处理。
//
// 核心功能：
//   - 错误码定义（HTTP/gRPC）
//   - 错误构造器
//   - 错误响应转换
//
// 使用示例：
//
//	err := errors.New(errors.CodeNotFound, "user not found")
//	err = errors.Wrap(err, errors.CodeInternal, "failed to get user")
//	httpCode := err.HTTPStatus()
package errors
