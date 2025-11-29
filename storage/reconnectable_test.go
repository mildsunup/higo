package storage

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockStorage struct {
	*Base
	connectErr   error
	pingErr      error
	connectCount int
}

func newMockStorage() *mockStorage {
	return &mockStorage{Base: NewBase("mock", TypeUnknown)}
}

func (m *mockStorage) Connect(ctx context.Context) error {
	m.connectCount++
	if m.connectErr != nil {
		return m.connectErr
	}
	m.SetState(StateConnected)
	return nil
}

func (m *mockStorage) Ping(ctx context.Context) error {
	return m.pingErr
}

func (m *mockStorage) Close(ctx context.Context) error {
	m.SetState(StateDisconnected)
	return nil
}

func TestReconnectable_Connect(t *testing.T) {
	mock := newMockStorage()
	r := NewReconnectable(mock, ReconnectConfig{
		MaxRetries:      2,
		InitialInterval: time.Millisecond,
		MaxInterval:     10 * time.Millisecond,
		Multiplier:      2,
	})

	// 成功连接
	if err := r.Connect(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.connectCount != 1 {
		t.Fatalf("expected 1 connect, got %d", mock.connectCount)
	}
}

func TestReconnectable_ConnectRetry(t *testing.T) {
	mock := newMockStorage()
	mock.connectErr = errors.New("connection refused")

	r := NewReconnectable(mock, ReconnectConfig{
		MaxRetries:      2,
		InitialInterval: time.Millisecond,
		MaxInterval:     10 * time.Millisecond,
		Multiplier:      2,
	})

	err := r.Connect(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	// 1 initial + 2 retries = 3
	if mock.connectCount != 3 {
		t.Fatalf("expected 3 connect attempts, got %d", mock.connectCount)
	}
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		err    error
		expect bool
	}{
		{nil, false},
		{errors.New("connection refused"), true},
		{errors.New("timeout"), true},
		{errors.New("EOF"), true},
		{errors.New("some other error"), false},
	}

	for _, tt := range tests {
		got := isConnectionError(tt.err)
		if got != tt.expect {
			t.Errorf("isConnectionError(%v) = %v, want %v", tt.err, got, tt.expect)
		}
	}
}
