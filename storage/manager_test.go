package storage

import (
	"context"
	"errors"
	"testing"
)

func TestManager_Register(t *testing.T) {
	m := NewManager()
	mock := newMockStorage()

	if err := m.Register(mock); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 重复注册
	if err := m.Register(mock); err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestManager_Get(t *testing.T) {
	m := NewManager()
	mock := newMockStorage()
	_ = m.Register(mock)

	s, ok := m.Get("mock")
	if !ok {
		t.Fatal("expected to find storage")
	}
	if s.Name() != "mock" {
		t.Fatalf("expected name 'mock', got %s", s.Name())
	}

	_, ok = m.Get("nonexistent")
	if ok {
		t.Fatal("expected not to find storage")
	}
}

func TestManager_GetByType(t *testing.T) {
	m := NewManager()
	mock1 := &mockStorage{Base: NewBase("mock1", TypeMySQL)}
	mock2 := &mockStorage{Base: NewBase("mock2", TypeMySQL)}
	mock3 := &mockStorage{Base: NewBase("mock3", TypeRedis)}

	_ = m.Register(mock1)
	_ = m.Register(mock2)
	_ = m.Register(mock3)

	mysqlStorages := m.GetByType(TypeMySQL)
	if len(mysqlStorages) != 2 {
		t.Fatalf("expected 2 MySQL storages, got %d", len(mysqlStorages))
	}
}

func TestManager_ConnectAll(t *testing.T) {
	m := NewManager()
	mock1 := newMockStorage()
	mock1.Base = NewBase("mock1", TypeUnknown)
	mock2 := newMockStorage()
	mock2.Base = NewBase("mock2", TypeUnknown)

	_ = m.Register(mock1)
	_ = m.Register(mock2)

	if err := m.ConnectAll(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mock1.connectCount != 1 || mock2.connectCount != 1 {
		t.Fatal("expected both storages to be connected")
	}
}

func TestManager_ConnectAll_Error(t *testing.T) {
	m := NewManager()
	mock1 := newMockStorage()
	mock1.Base = NewBase("mock1", TypeUnknown)
	mock2 := newMockStorage()
	mock2.Base = NewBase("mock2", TypeUnknown)
	mock2.connectErr = errors.New("connect failed")

	_ = m.Register(mock1)
	_ = m.Register(mock2)

	err := m.ConnectAll(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestManager_HealthCheck(t *testing.T) {
	m := NewManager()
	mock1 := newMockStorage()
	mock1.Base = NewBase("mock1", TypeMySQL)
	mock2 := newMockStorage()
	mock2.Base = NewBase("mock2", TypeRedis)
	mock2.pingErr = errors.New("ping failed")

	_ = m.Register(mock1)
	_ = m.Register(mock2)

	results := m.HealthCheck(context.Background())
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// 检查结果
	for _, r := range results {
		if r.Name == "mock1" && !r.Healthy {
			t.Error("mock1 should be healthy")
		}
		if r.Name == "mock2" && r.Healthy {
			t.Error("mock2 should be unhealthy")
		}
	}
}

func TestManager_List(t *testing.T) {
	m := NewManager()
	mock1 := &mockStorage{Base: NewBase("a", TypeUnknown)}
	mock2 := &mockStorage{Base: NewBase("b", TypeUnknown)}

	_ = m.Register(mock1)
	_ = m.Register(mock2)

	list := m.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
	// 保持注册顺序
	if list[0] != "a" || list[1] != "b" {
		t.Fatalf("expected order [a, b], got %v", list)
	}
}
