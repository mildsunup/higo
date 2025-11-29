package server

import (
	"context"
	"sync"
)

// Group 服务器组，管理多个服务器的生命周期
type Group struct {
	servers []Server
	mu      sync.RWMutex
}

// NewGroup 创建服务器组
func NewGroup(servers ...Server) *Group {
	return &Group{servers: servers}
}

// Add 添加服务器
func (g *Group) Add(s Server) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.servers = append(g.servers, s)
}

// Start 启动所有服务器
func (g *Group) Start(ctx context.Context) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, s := range g.servers {
		if err := s.Start(ctx); err != nil {
			// 启动失败，停止已启动的服务器
			g.stopServers(ctx, g.servers[:len(g.servers)-1])
			return err
		}
	}
	return nil
}

// Stop 停止所有服务器（逆序）
func (g *Group) Stop(ctx context.Context) error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.stopServers(ctx, g.servers)
}

func (g *Group) stopServers(ctx context.Context, servers []Server) error {
	var lastErr error
	// 逆序停止
	for i := len(servers) - 1; i >= 0; i-- {
		if err := servers[i].Stop(ctx); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// Servers 返回所有服务器
func (g *Group) Servers() []Server {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]Server, len(g.servers))
	copy(result, g.servers)
	return result
}
