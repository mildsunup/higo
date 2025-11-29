package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type manager struct {
	mu       sync.RWMutex
	storages map[string]Storage
	order    []string // 保持注册顺序
}

// NewManager 创建存储管理器
func NewManager() Manager {
	return &manager{
		storages: make(map[string]Storage),
	}
}

func (m *manager) Register(s Storage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := s.Name()
	if _, exists := m.storages[name]; exists {
		return fmt.Errorf("storage %q already registered", name)
	}

	m.storages[name] = s
	m.order = append(m.order, name)
	return nil
}

func (m *manager) Get(name string) (Storage, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.storages[name]
	return s, ok
}

func (m *manager) MustGet(name string) Storage {
	s, ok := m.Get(name)
	if !ok {
		panic(fmt.Sprintf("storage %q not found", name))
	}
	return s
}

func (m *manager) GetByType(typ Type) []Storage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Storage
	for _, s := range m.storages {
		if s.Type() == typ {
			result = append(result, s)
		}
	}
	return result
}

// ConnectAll 并发连接所有存储
func (m *manager) ConnectAll(ctx context.Context) error {
	m.mu.RLock()
	storages := make([]Storage, 0, len(m.storages))
	for _, name := range m.order {
		storages = append(storages, m.storages[name])
	}
	m.mu.RUnlock()

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	for _, s := range storages {
		wg.Add(1)
		go func(s Storage) {
			defer wg.Done()
			if err := s.Connect(ctx); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", s.Name(), err))
				mu.Unlock()
			}
		}(s)
	}

	wg.Wait()
	return errors.Join(errs...)
}

// CloseAll 按注册逆序关闭
func (m *manager) CloseAll(ctx context.Context) error {
	m.mu.RLock()
	order := make([]string, len(m.order))
	copy(order, m.order)
	m.mu.RUnlock()

	var errs []error
	// 逆序关闭
	for i := len(order) - 1; i >= 0; i-- {
		if s, ok := m.Get(order[i]); ok {
			if err := s.Close(ctx); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", s.Name(), err))
			}
		}
	}
	return errors.Join(errs...)
}

func (m *manager) HealthCheck(ctx context.Context) []HealthStatus {
	m.mu.RLock()
	storages := make([]Storage, 0, len(m.storages))
	for _, name := range m.order {
		storages = append(storages, m.storages[name])
	}
	m.mu.RUnlock()

	results := make([]HealthStatus, len(storages))
	var wg sync.WaitGroup

	for i, s := range storages {
		wg.Add(1)
		go func(idx int, s Storage) {
			defer wg.Done()
			results[idx] = checkHealth(ctx, s)
		}(i, s)
	}

	wg.Wait()
	return results
}

func checkHealth(ctx context.Context, s Storage) HealthStatus {
	status := HealthStatus{
		Name:      s.Name(),
		Type:      s.Type(),
		State:     s.State().String(),
		CheckedAt: time.Now(),
	}

	start := time.Now()
	err := s.Ping(ctx)
	status.Latency = time.Since(start)
	status.Healthy = err == nil

	if err != nil {
		status.Error = err.Error()
	}

	// 获取连接池统计
	if sp, ok := s.(StatsProvider); ok {
		stats := sp.Stats()
		status.Stats = &stats
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
