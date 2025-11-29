package ddd

import "sync"

// AggregateRoot 聚合根基类
// 聚合根是一致性边界，负责维护聚合内的不变量
type AggregateRoot[ID Identifier] struct {
	EntityBase[ID]
	version int64
	events  []DomainEvent
	mu      sync.Mutex
}

// NewAggregateRoot 创建聚合根
func NewAggregateRoot[ID Identifier](id ID) AggregateRoot[ID] {
	return AggregateRoot[ID]{
		EntityBase: NewEntityBase(id),
		version:    0,
	}
}

// Version 获取版本号（用于乐观锁）
func (a *AggregateRoot[ID]) Version() int64 { return a.version }

// SetVersion 设置版本号
func (a *AggregateRoot[ID]) SetVersion(v int64) { a.version = v }

// IncrementVersion 递增版本号
func (a *AggregateRoot[ID]) IncrementVersion() { a.version++ }

// RaiseEvent 发布领域事件
func (a *AggregateRoot[ID]) RaiseEvent(event DomainEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = append(a.events, event)
}

// PullEvents 拉取并清空事件（发布后调用）
func (a *AggregateRoot[ID]) PullEvents() []DomainEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	events := a.events
	a.events = nil
	return events
}

// GetEvents 获取事件（不清空）
func (a *AggregateRoot[ID]) GetEvents() []DomainEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := make([]DomainEvent, len(a.events))
	copy(result, a.events)
	return result
}

// ClearEvents 清空事件
func (a *AggregateRoot[ID]) ClearEvents() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = nil
}

// HasPendingEvents 是否有待发布事件
func (a *AggregateRoot[ID]) HasPendingEvents() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.events) > 0
}
