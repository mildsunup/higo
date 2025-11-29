package errors

import (
	"errors"
	"fmt"
	"runtime"
)

// Error 应用错误
type Error struct {
	code     Code
	message  string
	cause    error
	metadata map[string]any
	stack    []uintptr
}

// New 创建错误
func New(code Code, message string) *Error {
	return &Error{
		code:    code,
		message: message,
		stack:   callers(),
	}
}

// Newf 创建格式化错误
func Newf(code Code, format string, args ...any) *Error {
	return &Error{
		code:    code,
		message: fmt.Sprintf(format, args...),
		stack:   callers(),
	}
}

// FromCode 从错误码创建错误
func FromCode(code Code) *Error {
	return &Error{
		code:    code,
		message: code.Message(),
		stack:   callers(),
	}
}

// Wrap 包装错误
func Wrap(err error, code Code, message string) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		code:    code,
		message: message,
		cause:   err,
		stack:   callers(),
	}
}

// Wrapf 包装错误（格式化）
func Wrapf(err error, code Code, format string, args ...any) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		code:    code,
		message: fmt.Sprintf(format, args...),
		cause:   err,
		stack:   callers(),
	}
}

// Error 实现 error 接口
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

// Unwrap 返回底层错误
func (e *Error) Unwrap() error {
	return e.cause
}

// Code 返回错误码
func (e *Error) Code() Code {
	return e.code
}

// Message 返回错误消息
func (e *Error) Message() string {
	return e.message
}

// HTTPStatus 返回 HTTP 状态码
func (e *Error) HTTPStatus() int {
	return e.code.HTTPStatus()
}

// WithMeta 添加元数据
func (e *Error) WithMeta(key string, value any) *Error {
	if e.metadata == nil {
		e.metadata = make(map[string]any)
	}
	e.metadata[key] = value
	return e
}

// WithMetadata 批量添加元数据
func (e *Error) WithMetadata(meta map[string]any) *Error {
	if e.metadata == nil {
		e.metadata = make(map[string]any)
	}
	for k, v := range meta {
		e.metadata[k] = v
	}
	return e
}

// Metadata 返回元数据
func (e *Error) Metadata() map[string]any {
	return e.metadata
}

// GetMeta 获取元数据
func (e *Error) GetMeta(key string) any {
	if e.metadata == nil {
		return nil
	}
	return e.metadata[key]
}

// StackTrace 返回堆栈信息
func (e *Error) StackTrace() []string {
	if e.stack == nil {
		return nil
	}

	frames := runtime.CallersFrames(e.stack)
	var trace []string

	for {
		frame, more := frames.Next()
		trace = append(trace, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}
	return trace
}

// Is 判断错误是否匹配
func (e *Error) Is(target error) bool {
	if target == nil {
		return false
	}

	var t *Error
	if errors.As(target, &t) {
		return e.code == t.code
	}
	return false
}

// callers 获取调用栈
func callers() []uintptr {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[:n]
}

// --- 标准库兼容 ---

// Is 检查错误链中是否包含目标错误
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 将错误转换为目标类型
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Join 合并多个错误
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// --- 辅助函数 ---

// GetCode 从错误中提取错误码
func GetCode(err error) Code {
	if err == nil {
		return OK
	}
	var e *Error
	if errors.As(err, &e) {
		return e.code
	}
	return Unknown
}

// GetHTTPStatus 从错误中提取 HTTP 状态码
func GetHTTPStatus(err error) int {
	if err == nil {
		return 200
	}
	var e *Error
	if errors.As(err, &e) {
		return e.HTTPStatus()
	}
	return 500
}

// GetMessage 从错误中提取消息
func GetMessage(err error) string {
	if err == nil {
		return ""
	}
	var e *Error
	if errors.As(err, &e) {
		return e.message
	}
	return err.Error()
}

// IsCode 检查错误是否为指定错误码
func IsCode(err error, code Code) bool {
	return GetCode(err) == code
}
