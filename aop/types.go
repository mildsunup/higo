// Package aop 提供面向切面编程支持
// 设计原则：零反射、类型安全、高性能、易扩展
package aop

import "context"

// Invocation 方法调用上下文
type Invocation struct {
	ctx      context.Context
	method   string
	target   any
	args     []any
	result   any
	err      error
	metadata map[string]any
	index    int
	chain    []Interceptor
}

// Context 获取上下文
func (i *Invocation) Context() context.Context { return i.ctx }

// SetContext 设置上下文
func (i *Invocation) SetContext(ctx context.Context) { i.ctx = ctx }

// Method 获取方法名
func (i *Invocation) Method() string { return i.method }

// Target 获取目标对象
func (i *Invocation) Target() any { return i.target }

// Args 获取参数
func (i *Invocation) Args() []any { return i.args }

// Result 获取结果
func (i *Invocation) Result() any { return i.result }

// SetResult 设置结果
func (i *Invocation) SetResult(result any) { i.result = result }

// Error 获取错误
func (i *Invocation) Error() error { return i.err }

// SetError 设置错误
func (i *Invocation) SetError(err error) { i.err = err }

// Set 设置元数据
func (i *Invocation) Set(key string, val any) {
	if i.metadata == nil {
		i.metadata = make(map[string]any)
	}
	i.metadata[key] = val
}

// Get 获取元数据
func (i *Invocation) Get(key string) (any, bool) {
	if i.metadata == nil {
		return nil, false
	}
	v, ok := i.metadata[key]
	return v, ok
}

// Proceed 执行下一个拦截器或目标方法
func (i *Invocation) Proceed() error {
	if i.index < len(i.chain) {
		interceptor := i.chain[i.index]
		i.index++
		return interceptor.Intercept(i)
	}
	return nil
}

// Interceptor 拦截器接口
type Interceptor interface {
	Intercept(inv *Invocation) error
}

// InterceptorFunc 函数式拦截器适配器
type InterceptorFunc func(*Invocation) error

// Intercept 实现 Interceptor 接口
func (f InterceptorFunc) Intercept(inv *Invocation) error {
	return f(inv)
}
