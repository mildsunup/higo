package mq_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"higo/mq"
	"higo/mq/memory"
)

func TestMemoryClient_PublishSubscribe(t *testing.T) {
	client := memory.New("test")

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer client.Close()

	var received atomic.Int32
	handler := func(ctx context.Context, msg *mq.Message) error {
		received.Add(1)
		return nil
	}

	if err := client.Subscribe(ctx, "test-topic", handler); err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// 发布消息
	for i := 0; i < 5; i++ {
		_, err := client.Publish(ctx, "test-topic", []byte("hello"))
		if err != nil {
			t.Fatalf("publish failed: %v", err)
		}
	}

	// 等待消费
	time.Sleep(100 * time.Millisecond)

	if received.Load() != 5 {
		t.Errorf("expected 5 messages, got %d", received.Load())
	}
}

func TestMemoryClient_PublishWithOptions(t *testing.T) {
	client := memory.New("test")

	ctx := context.Background()
	_ = client.Connect(ctx)
	defer client.Close()

	var receivedKey string
	var receivedHeaders map[string]string

	handler := func(ctx context.Context, msg *mq.Message) error {
		receivedKey = msg.Key
		receivedHeaders = msg.Headers
		return nil
	}

	_ = client.Subscribe(ctx, "test-topic", handler)

	_, err := client.Publish(ctx, "test-topic", []byte("hello"),
		mq.WithKey("my-key"),
		mq.WithHeaders(map[string]string{"x-custom": "value"}),
	)
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if receivedKey != "my-key" {
		t.Errorf("expected key 'my-key', got %s", receivedKey)
	}
	if receivedHeaders["x-custom"] != "value" {
		t.Errorf("expected header 'x-custom'='value', got %v", receivedHeaders)
	}
}

func TestManager_Register(t *testing.T) {
	mgr := mq.NewManager()
	client := memory.New("test")

	if err := mgr.Register(client); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// 重复注册
	if err := mgr.Register(client); err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestManager_Get(t *testing.T) {
	mgr := mq.NewManager()
	client := memory.New("test")
	_ = mgr.Register(client)

	c, ok := mgr.Get("test")
	if !ok {
		t.Fatal("expected to find client")
	}
	if c.Name() != "test" {
		t.Errorf("expected name 'test', got %s", c.Name())
	}

	_, ok = mgr.Get("nonexistent")
	if ok {
		t.Fatal("expected not to find client")
	}
}

func TestManager_GetByType(t *testing.T) {
	mgr := mq.NewManager()
	_ = mgr.Register(memory.New("mem1"))
	_ = mgr.Register(memory.New("mem2"))

	clients := mgr.GetByType(mq.TypeMemory)
	if len(clients) != 2 {
		t.Errorf("expected 2 memory clients, got %d", len(clients))
	}

	clients = mgr.GetByType(mq.TypeKafka)
	if len(clients) != 0 {
		t.Errorf("expected 0 kafka clients, got %d", len(clients))
	}
}

func TestManager_ConnectAll(t *testing.T) {
	mgr := mq.NewManager()
	_ = mgr.Register(memory.New("mem1"))
	_ = mgr.Register(memory.New("mem2"))

	ctx := context.Background()
	if err := mgr.ConnectAll(ctx); err != nil {
		t.Fatalf("connect all failed: %v", err)
	}

	// 验证所有客户端已连接
	for _, name := range mgr.List() {
		c, _ := mgr.Get(name)
		if c.State() != mq.StateConnected {
			t.Errorf("client %s not connected", name)
		}
	}
}

func TestManager_HealthCheck(t *testing.T) {
	mgr := mq.NewManager()
	client := memory.New("test")
	_ = mgr.Register(client)
	_ = client.Connect(context.Background())

	results := mgr.HealthCheck(context.Background())
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if !results[0].Healthy {
		t.Error("expected healthy status")
	}
}

func TestBuilder(t *testing.T) {
	client := memory.New("test")

	// 使用 builder 添加装饰器 (不带 metrics)
	wrapped := mq.NewBuilder(client).Build()

	if wrapped.Name() != "test" {
		t.Errorf("expected name 'test', got %s", wrapped.Name())
	}

	// 解包
	unwrapped := mq.Unwrap(wrapped)
	if unwrapped != client {
		t.Error("unwrap should return original client")
	}
}
