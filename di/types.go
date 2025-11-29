// Package di 提供依赖注入基础设施
// 设计原则：配合 Wire 编译时注入，运行时只做组件注册
package di

// Scope 依赖作用域
type Scope int

const (
	Singleton Scope = iota // 单例：整个应用生命周期
	Prototype              // 原型：每次请求新实例
	Request                // 请求级：每个请求一个实例
)

// Closer 关闭接口
type Closer interface {
	Close() error
}

// Named 命名接口
type Named interface {
	Name() string
}
