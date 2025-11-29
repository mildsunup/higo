package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemory_GetSet(t *testing.T) {
	cache, err := NewMemory(DefaultMemoryConfig())
	if err != nil {
		t.Fatalf("create memory cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// Set
	if err := cache.Set(ctx, "key1", "value1", time.Minute); err != nil {
		t.Fatalf("set: %v", err)
	}

	// 等待 ristretto 异步写入
	time.Sleep(10 * time.Millisecond)

	// Get
	var result string
	if err := cache.Get(ctx, "key1", &result); err != nil {
		t.Fatalf("get: %v", err)
	}

	if result != "value1" {
		t.Errorf("expected 'value1', got %s", result)
	}
}

func TestMemory_NotFound(t *testing.T) {
	cache, _ := NewMemory(DefaultMemoryConfig())
	defer cache.Close()

	var result string
	err := cache.Get(context.Background(), "nonexistent", &result)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemory_Delete(t *testing.T) {
	cache, _ := NewMemory(DefaultMemoryConfig())
	defer cache.Close()

	ctx := context.Background()
	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	time.Sleep(10 * time.Millisecond)

	_ = cache.Delete(ctx, "key1")

	var result string
	err := cache.Get(ctx, "key1", &result)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMemory_Exists(t *testing.T) {
	cache, _ := NewMemory(DefaultMemoryConfig())
	defer cache.Close()

	ctx := context.Background()
	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	time.Sleep(10 * time.Millisecond)

	exists, _ := cache.Exists(ctx, "key1")
	if !exists {
		t.Error("expected key to exist")
	}

	exists, _ = cache.Exists(ctx, "nonexistent")
	if exists {
		t.Error("expected key to not exist")
	}
}

func TestMemory_Stats(t *testing.T) {
	cache, _ := NewMemory(DefaultMemoryConfig())
	defer cache.Close()

	ctx := context.Background()
	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	time.Sleep(10 * time.Millisecond)

	var result string
	_ = cache.Get(ctx, "key1", &result) // hit
	_ = cache.Get(ctx, "key2", &result) // miss

	stats := cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
}

func TestMemory_WithPrefix(t *testing.T) {
	cache, _ := NewMemory(DefaultMemoryConfig(), WithPrefix("test"))
	defer cache.Close()

	ctx := context.Background()
	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	time.Sleep(10 * time.Millisecond)

	var result string
	err := cache.Get(ctx, "key1", &result)
	if err != nil {
		t.Fatalf("get with prefix: %v", err)
	}
	if result != "value1" {
		t.Errorf("expected 'value1', got %s", result)
	}
}

func TestMemory_StructValue(t *testing.T) {
	cache, _ := NewMemory(DefaultMemoryConfig())
	defer cache.Close()

	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	ctx := context.Background()
	user := User{ID: 1, Name: "test"}
	_ = cache.Set(ctx, "user:1", user, time.Minute)
	time.Sleep(10 * time.Millisecond)

	var result User
	err := cache.Get(ctx, "user:1", &result)
	if err != nil {
		t.Fatalf("get struct: %v", err)
	}

	if result.ID != 1 || result.Name != "test" {
		t.Errorf("expected {1, test}, got %+v", result)
	}
}

func TestMultiLevel_GetSet(t *testing.T) {
	l1, _ := NewMemory(DefaultMemoryConfig())
	l2, _ := NewMemory(DefaultMemoryConfig()) // 用内存模拟 L2

	cache := NewMultiLevel(MultiLevelConfig{L1: l1, L2: l2})
	defer cache.Close()

	ctx := context.Background()

	// Set
	if err := cache.Set(ctx, "key1", "value1", time.Minute); err != nil {
		t.Fatalf("set: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	// Get
	var result string
	if err := cache.Get(ctx, "key1", &result); err != nil {
		t.Fatalf("get: %v", err)
	}

	if result != "value1" {
		t.Errorf("expected 'value1', got %s", result)
	}
}

func TestMultiLevel_L2Fallback(t *testing.T) {
	l1, _ := NewMemory(DefaultMemoryConfig())
	l2, _ := NewMemory(DefaultMemoryConfig())

	cache := NewMultiLevel(MultiLevelConfig{L1: l1, L2: l2})
	defer cache.Close()

	ctx := context.Background()

	// 直接写入 L2
	_ = l2.Set(ctx, "key1", "value1", time.Minute)
	time.Sleep(10 * time.Millisecond)

	// 从 MultiLevel 获取（应该从 L2 获取并回填 L1）
	var result string
	if err := cache.Get(ctx, "key1", &result); err != nil {
		t.Fatalf("get from L2: %v", err)
	}

	if result != "value1" {
		t.Errorf("expected 'value1', got %s", result)
	}

	// 验证 L1 已回填
	time.Sleep(10 * time.Millisecond)
	var l1Result string
	if err := l1.Get(ctx, "key1", &l1Result); err != nil {
		t.Fatalf("L1 should have been populated: %v", err)
	}
}

func TestMultiLevel_GetOrLoad(t *testing.T) {
	l1, _ := NewMemory(DefaultMemoryConfig())
	l2, _ := NewMemory(DefaultMemoryConfig())

	cache := NewMultiLevel(MultiLevelConfig{L1: l1, L2: l2})
	defer cache.Close()

	ctx := context.Background()
	loadCount := 0

	loader := func(ctx context.Context) (any, error) {
		loadCount++
		return "loaded_value", nil
	}

	var result string
	err := cache.GetOrLoad(ctx, "key1", &result, loader, time.Minute)
	if err != nil {
		t.Fatalf("GetOrLoad: %v", err)
	}

	if result != "loaded_value" {
		t.Errorf("expected 'loaded_value', got %s", result)
	}

	if loadCount != 1 {
		t.Errorf("expected loader to be called once, got %d", loadCount)
	}

	// 再次调用，应该从缓存获取
	time.Sleep(10 * time.Millisecond)
	var result2 string
	_ = cache.GetOrLoad(ctx, "key1", &result2, loader, time.Minute)

	if loadCount != 1 {
		t.Errorf("loader should not be called again, got %d", loadCount)
	}
}

func TestSerializer_JSON(t *testing.T) {
	s := &JSONSerializer{}

	type Data struct {
		Name string `json:"name"`
	}

	data := Data{Name: "test"}
	bytes, err := s.Marshal(data)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var result Data
	if err := s.Unmarshal(bytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if result.Name != "test" {
		t.Errorf("expected 'test', got %s", result.Name)
	}
}
