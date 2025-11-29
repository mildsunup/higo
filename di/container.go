package di

import (
	"sync"
)

// Container 依赖注入容器
// 只负责组件注册和获取，生命周期由 runtime.App 管理
type Container struct {
	mu         sync.RWMutex
	components map[string]any
}

// NewContainer 创建容器
func NewContainer() *Container {
	return &Container{
		components: make(map[string]any),
	}
}

// Register 注册组件
func (c *Container) Register(name string, component any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.components[name] = component
}

// Get 获取组件
func (c *Container) Get(name string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.components[name]
	return v, ok
}

// MustGet 获取组件，不存在则 panic
func (c *Container) MustGet(name string) any {
	v, ok := c.Get(name)
	if !ok {
		panic("di: component not found: " + name)
	}
	return v
}

// Has 检查组件是否存在
func (c *Container) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.components[name]
	return ok
}

// All 返回所有组件
func (c *Container) All() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]any, len(c.components))
	for k, v := range c.components {
		result[k] = v
	}
	return result
}

// Remove 移除组件
func (c *Container) Remove(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.components, name)
}

// Clear 清空所有组件
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.components = make(map[string]any)
}
