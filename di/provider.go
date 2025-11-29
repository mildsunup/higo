package di

import (
	"context"
	"sync"
)

// Provider 泛型 Provider 接口
type Provider[T any] interface {
	Provide(ctx context.Context) (T, error)
}

// ProviderFunc 函数式 Provider
type ProviderFunc[T any] func(ctx context.Context) (T, error)

func (f ProviderFunc[T]) Provide(ctx context.Context) (T, error) {
	return f(ctx)
}

// SingletonProvider 单例 Provider
type SingletonProvider[T any] struct {
	once     sync.Once
	instance T
	err      error
	factory  func(context.Context) (T, error)
}

// NewSingleton 创建单例 Provider
func NewSingleton[T any](factory func(context.Context) (T, error)) *SingletonProvider[T] {
	return &SingletonProvider[T]{factory: factory}
}

func (p *SingletonProvider[T]) Provide(ctx context.Context) (T, error) {
	p.once.Do(func() {
		p.instance, p.err = p.factory(ctx)
	})
	return p.instance, p.err
}

// LazyProvider 延迟初始化 Provider
type LazyProvider[T any] struct {
	mu       sync.RWMutex
	instance T
	init     bool
	factory  func(context.Context) (T, error)
}

// NewLazy 创建延迟 Provider
func NewLazy[T any](factory func(context.Context) (T, error)) *LazyProvider[T] {
	return &LazyProvider[T]{factory: factory}
}

func (p *LazyProvider[T]) Provide(ctx context.Context) (T, error) {
	p.mu.RLock()
	if p.init {
		instance := p.instance
		p.mu.RUnlock()
		return instance, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.init {
		return p.instance, nil
	}

	var err error
	p.instance, err = p.factory(ctx)
	if err != nil {
		var zero T
		return zero, err
	}
	p.init = true
	return p.instance, nil
}

// Reset 重置延迟 Provider（用于测试）
func (p *LazyProvider[T]) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	var zero T
	p.instance = zero
	p.init = false
}

// CachedProvider 带缓存的 Provider
type CachedProvider[T any] struct {
	mu       sync.RWMutex
	cache    map[string]T
	factory  func(ctx context.Context, key string) (T, error)
}

// NewCached 创建缓存 Provider
func NewCached[T any](factory func(ctx context.Context, key string) (T, error)) *CachedProvider[T] {
	return &CachedProvider[T]{
		cache:   make(map[string]T),
		factory: factory,
	}
}

func (p *CachedProvider[T]) Get(ctx context.Context, key string) (T, error) {
	p.mu.RLock()
	if v, ok := p.cache[key]; ok {
		p.mu.RUnlock()
		return v, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()
	if v, ok := p.cache[key]; ok {
		return v, nil
	}

	v, err := p.factory(ctx, key)
	if err != nil {
		var zero T
		return zero, err
	}
	p.cache[key] = v
	return v, nil
}
