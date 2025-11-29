package storage

import (
	"context"
	"sync/atomic"
)

// Base 存储基类，提供通用功能
type Base struct {
	name  string
	typ   Type
	state atomic.Int32
}

// NewBase 创建基类
func NewBase(name string, typ Type) *Base {
	b := &Base{name: name, typ: typ}
	b.state.Store(int32(StateDisconnected))
	return b
}

// Name 返回存储名称
func (b *Base) Name() string { return b.name }

// Type 返回存储类型
func (b *Base) Type() Type { return b.typ }

// State 返回连接状态
func (b *Base) State() State { return State(b.state.Load()) }

// SetState 设置状态
func (b *Base) SetState(s State) { b.state.Store(int32(s)) }

// CompareAndSwapState CAS 状态
func (b *Base) CompareAndSwapState(old, new State) bool {
	return b.state.CompareAndSwap(int32(old), int32(new))
}

// Manager 存储管理器接口
type Manager interface {
	// Register 注册存储
	Register(storage Storage) error
	// Get 获取存储
	Get(name string) (Storage, bool)
	// MustGet 获取存储，不存在则 panic
	MustGet(name string) Storage
	// GetByType 按类型获取
	GetByType(typ Type) []Storage
	// ConnectAll 连接所有存储
	ConnectAll(ctx context.Context) error
	// CloseAll 关闭所有存储
	CloseAll(ctx context.Context) error
	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) []HealthStatus
	// List 列出所有存储名称
	List() []string
}
