package pool

import (
	"sync"
	"sync/atomic"
)

// Object 泛型对象池
type Object[T any] struct {
	pool    sync.Pool
	factory func() T
	reset   func(T)
	gets    atomic.Int64
	puts    atomic.Int64
	hits    atomic.Int64
	misses  atomic.Int64
}

// NewObject 创建对象池
// factory: 创建新对象的函数
// reset: 重置对象的函数（可选，用于归还时清理）
func NewObject[T any](factory func() T, reset func(T)) *Object[T] {
	p := &Object[T]{
		factory: factory,
		reset:   reset,
	}
	p.pool.New = func() any {
		return factory()
	}
	return p
}

// Get 获取对象
func (p *Object[T]) Get() T {
	p.gets.Add(1)
	obj := p.pool.Get()
	if obj == nil {
		p.misses.Add(1)
		return p.factory()
	}
	p.hits.Add(1)
	return obj.(T)
}

// Put 归还对象
func (p *Object[T]) Put(obj T) {
	p.puts.Add(1)
	if p.reset != nil {
		p.reset(obj)
	}
	p.pool.Put(obj)
}

// Stats 返回统计信息
func (p *Object[T]) Stats() Stats {
	return Stats{
		Gets:   p.gets.Load(),
		Puts:   p.puts.Load(),
		Hits:   p.hits.Load(),
		Misses: p.misses.Load(),
		InUse:  p.gets.Load() - p.puts.Load(),
	}
}

var _ Pool[any] = (*Object[any])(nil)
