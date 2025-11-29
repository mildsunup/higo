package mq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type manager struct {
	mu      sync.RWMutex
	clients map[string]Client
	order   []string
}

// NewManager 创建 MQ 管理器
func NewManager() Manager {
	return &manager{
		clients: make(map[string]Client),
	}
}

func (m *manager) Register(c Client) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := c.Name()
	if _, exists := m.clients[name]; exists {
		return fmt.Errorf("mq client %q already registered", name)
	}

	m.clients[name] = c
	m.order = append(m.order, name)
	return nil
}

func (m *manager) Get(name string) (Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clients[name]
	return c, ok
}

func (m *manager) GetByType(typ Type) []Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Client
	for _, c := range m.clients {
		if c.Type() == typ {
			result = append(result, c)
		}
	}
	return result
}

func (m *manager) ConnectAll(ctx context.Context) error {
	m.mu.RLock()
	clients := make([]Client, 0, len(m.clients))
	for _, name := range m.order {
		clients = append(clients, m.clients[name])
	}
	m.mu.RUnlock()

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	for _, c := range clients {
		wg.Add(1)
		go func(c Client) {
			defer wg.Done()
			if err := c.Connect(ctx); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", c.Name(), err))
				mu.Unlock()
			}
		}(c)
	}

	wg.Wait()
	return errors.Join(errs...)
}

func (m *manager) CloseAll(ctx context.Context) error {
	m.mu.RLock()
	order := make([]string, len(m.order))
	copy(order, m.order)
	m.mu.RUnlock()

	var errs []error
	for i := len(order) - 1; i >= 0; i-- {
		if c, ok := m.Get(order[i]); ok {
			if err := c.Close(); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", c.Name(), err))
			}
		}
	}
	return errors.Join(errs...)
}

func (m *manager) HealthCheck(ctx context.Context) []HealthStatus {
	m.mu.RLock()
	clients := make([]Client, 0, len(m.clients))
	for _, name := range m.order {
		clients = append(clients, m.clients[name])
	}
	m.mu.RUnlock()

	results := make([]HealthStatus, len(clients))
	var wg sync.WaitGroup

	for i, c := range clients {
		wg.Add(1)
		go func(idx int, c Client) {
			defer wg.Done()
			results[idx] = checkHealth(ctx, c)
		}(i, c)
	}

	wg.Wait()
	return results
}

func checkHealth(ctx context.Context, c Client) HealthStatus {
	status := HealthStatus{
		Name:      c.Name(),
		Type:      c.Type(),
		State:     c.State().String(),
		CheckedAt: time.Now(),
	}

	start := time.Now()
	err := c.Ping(ctx)
	status.Latency = time.Since(start)
	status.Healthy = err == nil

	if err != nil {
		status.Error = err.Error()
	}
	return status
}

func (m *manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.order))
	copy(result, m.order)
	return result
}
